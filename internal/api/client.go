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
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BaseURL    = "https://api.airwallex.com"
	APIVersion = "2025-11-11"

	// DefaultHTTPTimeout is the default timeout for HTTP requests.
	DefaultHTTPTimeout = 30 * time.Second

	// DefaultOperationTimeout is the default timeout for API operations.
	DefaultOperationTimeout = 45 * time.Second

	// TokenRefreshBuffer is how long before expiry to refresh the token.
	TokenRefreshBuffer = 60 * time.Second

	// MaxRateLimitRetries is the maximum number of retries on 429 responses.
	MaxRateLimitRetries = 3

	// Max5xxRetries is the maximum retries for server errors on idempotent requests.
	Max5xxRetries = 1

	// RateLimitBaseDelay is the initial delay for rate limit exponential backoff.
	RateLimitBaseDelay = 1 * time.Second

	// ServerErrorRetryDelay is the delay before retrying on 5xx errors.
	ServerErrorRetryDelay = 1 * time.Second

	// IdempotencyKeyBytes is the number of random bytes for idempotency keys.
	IdempotencyKeyBytes = 16

	// MaxIdleConns is the maximum number of idle connections.
	MaxIdleConns = 100

	// MaxConnsPerHost is the maximum connections per host.
	MaxConnsPerHost = 10

	// IdleConnTimeout is how long to keep idle connections.
	IdleConnTimeout = 90 * time.Second

	// CircuitBreakerThreshold is the number of consecutive 5xx errors to open the circuit.
	CircuitBreakerThreshold = 5

	// CircuitBreakerResetTime is how long to wait before attempting to close the circuit.
	CircuitBreakerResetTime = 30 * time.Second
)

var (
	rateLimitBaseDelay    = RateLimitBaseDelay
	serverErrorRetryDelay = ServerErrorRetryDelay
)

// withDefaultTimeout adds a timeout to the context if none exists.
func withDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	// Check if context already has a deadline
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, DefaultOperationTimeout)
}

// closeBody closes the response body and logs any error.
func closeBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			slog.Debug("failed to close response body", "error", err)
		}
	}
}

// parseRetryAfter computes a delay from Retry-After header.
// Supports seconds or HTTP-date formats. Returns ok=false if header is invalid.
func parseRetryAfter(header string, now time.Time) (delay time.Duration, ok bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(header); err == nil {
		if seconds < 0 {
			return 0, false
		}
		return time.Duration(seconds) * time.Second, true
	}
	if t, err := http.ParseTime(header); err == nil {
		if t.Before(now) {
			return 0, true
		}
		return t.Sub(now), true
	}
	return 0, false
}

type circuitBreaker struct {
	mu          sync.Mutex
	failures    int
	lastFailure time.Time
	open        bool
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	wasOpen := cb.open
	cb.failures = 0
	cb.open = false
	if wasOpen {
		slog.Info("circuit breaker reset")
	}
}

func (cb *circuitBreaker) recordFailure() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= CircuitBreakerThreshold {
		cb.open = true
		slog.Warn("circuit breaker opened", "failures", cb.failures)
		return true // circuit just opened
	}
	return false
}

func (cb *circuitBreaker) isOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if !cb.open {
		return false
	}
	// Check if reset time has passed
	if time.Since(cb.lastFailure) > CircuitBreakerResetTime {
		cb.open = false
		cb.failures = 0
		return false
	}
	return true
}

type Client struct {
	baseURL        string
	clientID       string
	apiKey         string
	accountID      string // Optional: for x-login-as header (multi-account API keys)
	token          *TokenCache
	tokenMu        sync.RWMutex
	httpClient     *http.Client
	circuitBreaker *circuitBreaker
}

type TokenCache struct {
	Token     string
	ExpiresAt time.Time
}

func NewClient(clientID, apiKey string) (*Client, error) {
	return newClient(BaseURL, clientID, apiKey, "", true)
}

// NewClientWithAccount creates a client with an account ID for x-login-as header.
// Use this when your API key has access to multiple accounts.
func NewClientWithAccount(clientID, apiKey, accountID string) (*Client, error) {
	return newClient(BaseURL, clientID, apiKey, accountID, true)
}

// NewClientWithBaseURL creates a client with a custom base URL (primarily for tests).
func NewClientWithBaseURL(baseURL, clientID, apiKey string) (*Client, error) {
	return newClient(baseURL, clientID, apiKey, "", false)
}

// NewClientWithBaseURLAndAccount creates a client with a custom base URL and account ID.
func NewClientWithBaseURLAndAccount(baseURL, clientID, apiKey, accountID string) (*Client, error) {
	return newClient(baseURL, clientID, apiKey, accountID, false)
}

func newClient(baseURL, clientID, apiKey, accountID string, requireHTTPS bool) (*Client, error) {
	if err := validateBaseURL(baseURL, requireHTTPS); err != nil {
		return nil, err
	}
	return &Client{
		baseURL:   baseURL,
		clientID:  clientID,
		apiKey:    apiKey,
		accountID: accountID,
		httpClient: &http.Client{
			Timeout: DefaultHTTPTimeout,
			Transport: &http.Transport{
				MaxIdleConns:    MaxIdleConns,
				MaxConnsPerHost: MaxConnsPerHost,
				IdleConnTimeout: IdleConnTimeout,
				TLSClientConfig: &tls.Config{
					MinVersion:         tls.VersionTLS12,
					InsecureSkipVerify: false, // Explicit: always verify certificates
				},
			},
		},
		circuitBreaker: &circuitBreaker{},
	}, nil
}

func validateBaseURL(baseURL string, requireHTTPS bool) error {
	if baseURL == "" {
		return fmt.Errorf("api base URL cannot be empty")
	}
	if strings.HasPrefix(baseURL, "https://") {
		return nil
	}
	if requireHTTPS {
		return fmt.Errorf("api base URL must use HTTPS")
	}
	if strings.HasPrefix(baseURL, "http://") {
		return nil
	}
	return fmt.Errorf("api base URL must start with http:// or https://")
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

// BaseURL returns the configured base URL for the API.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// doWithRetry executes the request with retry logic:
//   - 429: exponential backoff with jitter, max 3 retries (safe for all methods)
//     Respects Retry-After header if present
//   - 5xx: single retry after 1s, ONLY for idempotent methods (GET, HEAD, OPTIONS)
//   - 4xx: no retry
//   - Circuit breaker: stops requests after 5 consecutive 5xx errors
func (c *Client) doWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Check circuit breaker before making request
	if c.circuitBreaker.isOpen() {
		return nil, fmt.Errorf("circuit breaker open: API experiencing issues, retry later")
	}

	var resp *http.Response
	var err error

	// Separate retry counters for different error types
	retries429 := 0
	retries5xx := 0

	// Determine if the method is idempotent
	isIdempotent := req.Method == "GET" || req.Method == "HEAD" || req.Method == "OPTIONS"

	for {
		// Log request details in debug mode
		slog.Debug("api request",
			"method", req.Method,
			"url", req.URL.String(),
			"has_body", req.Body != nil,
		)

		start := time.Now()
		resp, err = c.httpClient.Do(req)
		if err != nil {
			slog.Debug("api request failed", "error", err)
			return nil, err
		}

		// Log response details in debug mode
		slog.Debug("api response",
			"status", resp.StatusCode,
			"content_length", resp.ContentLength,
		)

		// 4xx errors (except 429): no retry
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return resp, nil
		}

		// 429 rate limit: exponential backoff with jitter
		// Safe to retry for all methods because the request wasn't processed
		if resp.StatusCode == 429 {
			if retries429 >= MaxRateLimitRetries {
				return resp, nil
			}

			// Calculate backoff: base, 2x, 4x with jitter
			baseDelay := rateLimitBaseDelay
			if baseDelay <= 0 {
				baseDelay = RateLimitBaseDelay
			}
			baseDelay *= time.Duration(1 << retries429)
			//nolint:gosec // G404: jitter doesn't need crypto-strength randomness
			jitter := time.Duration(mathrand.Int63n(int64(baseDelay / 2)))
			delay := baseDelay + jitter

			// Check for Retry-After header (seconds or HTTP date)
			if retryAfterDelay, ok := parseRetryAfter(resp.Header.Get("Retry-After"), time.Now()); ok {
				delay = retryAfterDelay
			}

			slog.Info("rate limited, retrying", "delay", delay, "attempt", retries429+1, "max_retries", MaxRateLimitRetries)

			closeBody(resp)

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
			// Record failure for circuit breaker
			circuitOpened := c.circuitBreaker.recordFailure()
			if circuitOpened {
				slog.Warn("circuit breaker opened", "consecutive_failures", CircuitBreakerThreshold)
			}

			// Don't retry non-idempotent operations on 5xx
			if !isIdempotent {
				return resp, nil
			}

			// Only retry once
			if retries5xx >= Max5xxRetries {
				return resp, nil
			}

			delay := serverErrorRetryDelay
			if delay <= 0 {
				delay = ServerErrorRetryDelay
			}
			slog.Info("retrying after server error", "status", resp.StatusCode, "attempt", retries5xx+1, "delay", delay)

			closeBody(resp)

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
			retries5xx++
			continue
		}

		// Success: record for circuit breaker (2xx status codes)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			duration := time.Since(start)
			c.circuitBreaker.recordSuccess()
			slog.Debug("api request completed", "method", req.Method, "url", req.URL.Path, "status", resp.StatusCode, "duration_ms", duration.Milliseconds())
		}

		return resp, nil
	}
}

func (c *Client) ensureValidToken(ctx context.Context) error {
	c.tokenMu.RLock()
	valid := c.token != nil && time.Now().Add(TokenRefreshBuffer).Before(c.token.ExpiresAt)
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
	if c.token != nil && time.Now().Add(TokenRefreshBuffer).Before(c.token.ExpiresAt) {
		return nil
	}

	url := c.baseURL + Endpoints.Login.Path

	req, err := http.NewRequestWithContext(ctx, Endpoints.Login.Method, url, nil)
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
	defer closeBody(resp)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		apiErr := ParseAPIError(body)
		return WrapError(req.Method, url, resp.StatusCode, fmt.Errorf("authentication failed: %s", apiErr.Error()))
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

	expiresIn := time.Until(expiresAt)
	slog.Debug("refreshing api token", "expires_in", expiresIn)

	c.token = &TokenCache{
		Token:     result.Token,
		ExpiresAt: expiresAt,
	}
	return nil
}

// generateIdempotencyKey creates a unique key for idempotent operations.
func generateIdempotencyKey() (string, error) {
	b := make([]byte, IdempotencyKeyBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate idempotency key: %w", err)
	}
	return hex.EncodeToString(b), nil
}

var (
	idempotencyPatternsOnce sync.Once
	idempotencyPatterns     []string
)

// isFinancialOperation checks if the path is a financial operation that needs idempotency.
// Uses endpoint registry metadata to avoid drift.
func isFinancialOperation(path string) bool {
	// Remove query parameters if present
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	idempotencyPatternsOnce.Do(func() {
		idempotencyPatterns = collectIdempotencyPatterns()
	})

	for _, pattern := range idempotencyPatterns {
		if matchEndpointPath(path, pattern) {
			return true
		}
	}
	return false
}

func collectIdempotencyPatterns() []string {
	var patterns []string
	v := reflect.ValueOf(Endpoints)
	for i := 0; i < v.NumField(); i++ {
		ep := v.Field(i).Interface().(Endpoint)
		if ep.RequiresIdem {
			patterns = append(patterns, ep.Path)
		}
	}
	return patterns
}

func matchEndpointPath(path, pattern string) bool {
	if !strings.Contains(pattern, "{id}") {
		return path == pattern
	}

	parts := strings.Split(pattern, "{id}")
	if len(parts) == 0 {
		return path == pattern
	}

	if !strings.HasPrefix(path, parts[0]) {
		return false
	}

	pos := len(parts[0])
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		idx := strings.Index(path[pos:], part)
		if idx < 0 {
			return false
		}
		if idx == 0 {
			return false
		}
		placeholder := path[pos : pos+idx]
		if strings.Contains(placeholder, "/") {
			return false
		}
		pos = pos + idx + len(part)
	}

	return pos == len(path)
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

func (c *Client) doJSON(ctx context.Context, method, path string, body interface{}, out interface{}, okStatuses ...int) error {
	var resp *http.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = c.Get(ctx, path)
	case http.MethodPost:
		resp, err = c.Post(ctx, path, body)
	default:
		return fmt.Errorf("unsupported method %q", method)
	}
	if err != nil {
		return err
	}
	defer closeBody(resp)

	if len(okStatuses) == 0 {
		okStatuses = []int{http.StatusOK}
	}
	statusOK := false
	for _, code := range okStatuses {
		if resp.StatusCode == code {
			statusOK = true
			break
		}
	}
	if !statusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		apiErr := ParseAPIError(bodyBytes)
		return WrapError(method, path, resp.StatusCode, NormalizeAPIError(resp.StatusCode, apiErr))
	}

	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
