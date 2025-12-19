package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	mathrand "math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BaseURL    = "https://api.airwallex.com"
	APIVersion = "2025-11-11"
)

type Client struct {
	baseURL    string
	clientID   string
	apiKey     string
	accountID  string // Optional: for x-login-as header (multi-account API keys)
	token      *TokenCache
	tokenMu    sync.RWMutex
	httpClient *http.Client
}

type TokenCache struct {
	Token     string
	ExpiresAt time.Time
}

func NewClient(clientID, apiKey string) (*Client, error) {
	if !strings.HasPrefix(BaseURL, "https://") {
		return nil, fmt.Errorf("api base URL must use HTTPS")
	}
	return &Client{
		baseURL:  BaseURL,
		clientID: clientID,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		},
	}, nil
}

// NewClientWithAccount creates a client with an account ID for x-login-as header.
// Use this when your API key has access to multiple accounts.
func NewClientWithAccount(clientID, apiKey, accountID string) (*Client, error) {
	if !strings.HasPrefix(BaseURL, "https://") {
		return nil, fmt.Errorf("api base URL must use HTTPS")
	}
	return &Client{
		baseURL:   BaseURL,
		clientID:  clientID,
		apiKey:    apiKey,
		accountID: accountID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		},
	}, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if err := c.ensureValidToken(ctx); err != nil {
		return nil, fmt.Errorf("auth failed: %w", err)
	}

	c.tokenMu.RLock()
	token := c.token.Token
	c.tokenMu.RUnlock()

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("x-api-version", APIVersion)
	req.Header.Set("Content-Type", "application/json")
	return c.doWithRetry(ctx, req)
}

// doWithRetry executes the request with retry logic:
//   - 429: exponential backoff with jitter, max 3 retries (safe for all methods)
//     Respects Retry-After header if present
//   - 5xx: single retry after 1s, ONLY for idempotent methods (GET, HEAD, OPTIONS)
//   - 4xx: no retry
func (c *Client) doWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	// Separate retry counters for different error types
	retries429 := 0
	retries5xx := 0
	maxRetries := 3

	// Determine if the method is idempotent
	isIdempotent := req.Method == "GET" || req.Method == "HEAD" || req.Method == "OPTIONS"

	for {
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		// 4xx errors (except 429): no retry
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return resp, nil
		}

		// 429 rate limit: exponential backoff with jitter
		// Safe to retry for all methods because the request wasn't processed
		if resp.StatusCode == 429 {
			if retries429 >= maxRetries {
				return resp, nil
			}

			// Calculate backoff: 1s, 2s, 4s with jitter
			baseDelay := time.Duration(1<<retries429) * time.Second
			jitter := time.Duration(mathrand.Int63n(int64(baseDelay / 2)))
			delay := baseDelay + jitter

			// Check for Retry-After header (can be seconds or HTTP date)
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					delay = time.Duration(seconds) * time.Second
				}
				// Could also parse HTTP date format, but seconds is more common
			}

			slog.Info("rate limited, retrying", "delay", delay, "attempt", retries429+1, "max_retries", maxRetries)

			resp.Body.Close()

			// Replay request body if available
			if req.GetBody != nil {
				req.Body, err = req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("failed to replay request body: %w", err)
				}
			}

			// Context-aware sleep for cancellation support
			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			retries429++
			continue
		}

		// 5xx errors: retry once after 1s, ONLY for idempotent operations
		// Non-idempotent operations (POST, PUT, DELETE, PATCH) could have been partially
		// processed, so retrying could cause duplicates (e.g., duplicate transfers)
		if resp.StatusCode >= 500 {
			// Don't retry non-idempotent operations on 5xx
			if !isIdempotent {
				return resp, nil
			}

			// Only retry once
			if retries5xx > 0 {
				return resp, nil
			}
			resp.Body.Close()

			// Replay request body if available
			if req.GetBody != nil {
				req.Body, err = req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("failed to replay request body: %w", err)
				}
			}

			time.Sleep(1 * time.Second)
			retries5xx++
			continue
		}

		// Success or other status codes
		return resp, nil
	}
}

func (c *Client) ensureValidToken(ctx context.Context) error {
	c.tokenMu.RLock()
	valid := c.token != nil && time.Now().Add(60*time.Second).Before(c.token.ExpiresAt)
	c.tokenMu.RUnlock()

	if valid {
		return nil
	}
	return c.fetchToken(ctx)
}

func (c *Client) fetchToken(ctx context.Context) error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// Double-check pattern: another goroutine might have fetched while we waited
	if c.token != nil && time.Now().Add(60*time.Second).Before(c.token.ExpiresAt) {
		return nil
	}

	url := c.baseURL + "/api/v1/authentication/login"

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-client-id", c.clientID)
	req.Header.Set("x-api-key", c.apiKey)
	if c.accountID != "" {
		req.Header.Set("x-login-as", c.accountID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		apiErr := ParseAPIError(body)
		return fmt.Errorf("authentication failed: %s", apiErr.Error())
	}

	var result struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expires_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// Airwallex returns timestamps like "2025-12-17T08:25:19+0000" (no colon in tz)
	// Try RFC3339 first, then fallback to Airwallex format
	expiresAt, err := time.Parse(time.RFC3339, result.ExpiresAt)
	if err != nil {
		// Airwallex format: "2006-01-02T15:04:05-0700"
		expiresAt, err = time.Parse("2006-01-02T15:04:05-0700", result.ExpiresAt)
		if err != nil {
			return fmt.Errorf("parsing expires_at %q: %w", result.ExpiresAt, err)
		}
	}

	c.token = &TokenCache{
		Token:     result.Token,
		ExpiresAt: expiresAt,
	}
	return nil
}

// generateIdempotencyKey creates a unique key for idempotent operations.
func generateIdempotencyKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate idempotency key: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// isFinancialOperation checks if the path is a financial operation that needs idempotency.
func isFinancialOperation(path string) bool {
	financialPaths := []string{
		"/api/v1/transfers/create",
		"/api/v1/issuing/cards/create",
		"/api/v1/beneficiaries/create",
	}
	for _, fp := range financialPaths {
		if strings.Contains(path, fp) {
			return true
		}
	}
	return false
}

func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(ctx, req)
}

func (c *Client) Post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	var getBody func() (io.ReadCloser, error)
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
		// Enable body replay for retries
		getBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(data)), nil
		}
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	// Add idempotency key for financial operations
	if isFinancialOperation(path) {
		idempotencyKey, err := generateIdempotencyKey()
		if err != nil {
			return nil, err
		}
		req.Header.Set("x-idempotency-key", idempotencyKey)
	}

	req.GetBody = getBody
	return c.Do(ctx, req)
}
