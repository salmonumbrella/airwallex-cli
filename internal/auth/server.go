package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
)

var validAccountName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// clientLimit tracks attempts for a specific client
type clientLimit struct {
	count   int
	resetAt time.Time
}

// rateLimiter tracks attempts per client IP and endpoint to prevent brute-force
type rateLimiter struct {
	mu          sync.Mutex
	attempts    map[string]*clientLimit // key: "clientIP:endpoint"
	maxAttempts int
	window      time.Duration
}

// newRateLimiter creates a rate limiter with the given max attempts and time window
func newRateLimiter(maxAttempts int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		attempts:    make(map[string]*clientLimit),
		maxAttempts: maxAttempts,
		window:      window,
	}
}

// check verifies if the client has exceeded the rate limit for this endpoint
func (rl *rateLimiter) check(clientIP, endpoint string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := clientIP + ":" + endpoint
	now := time.Now()

	// Reset if window expired
	if limit, exists := rl.attempts[key]; exists && now.After(limit.resetAt) {
		delete(rl.attempts, key)
	}

	// Initialize new limit if doesn't exist
	if rl.attempts[key] == nil {
		rl.attempts[key] = &clientLimit{
			count:   1,
			resetAt: now.Add(rl.window),
		}
		return nil
	}

	// Increment and check limit
	rl.attempts[key].count++
	if rl.attempts[key].count > rl.maxAttempts {
		return fmt.Errorf("too many attempts, please try again later")
	}
	return nil
}

// startCleanup starts a background goroutine that periodically removes expired entries
func (rl *rateLimiter) startCleanup(interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rl.cleanup()
			case <-stop:
				return
			}
		}
	}()
}

// cleanup removes all expired entries from the rate limiter
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, limit := range rl.attempts {
		if now.After(limit.resetAt) {
			delete(rl.attempts, key)
		}
	}
}

// size returns the number of entries in the rate limiter (for testing)
func (rl *rateLimiter) size() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.attempts)
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// For localhost, use RemoteAddr
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// ValidateAccountName validates an account name
func ValidateAccountName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("account name cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("account name too long (max 64 characters)")
	}
	if !validAccountName.MatchString(name) {
		return fmt.Errorf("account name contains invalid characters (use only letters, numbers, dash, underscore)")
	}
	return nil
}

// ValidateClientID validates a client ID
func ValidateClientID(clientID string) error {
	if len(clientID) == 0 {
		return fmt.Errorf("client ID cannot be empty")
	}
	if len(clientID) > 128 {
		return fmt.Errorf("client ID too long (max 128 characters)")
	}
	return nil
}

// ValidateAPIKey validates an API key
func ValidateAPIKey(apiKey string) error {
	if len(apiKey) == 0 {
		return fmt.Errorf("API key cannot be empty")
	}
	if len(apiKey) > 256 {
		return fmt.Errorf("API key too long (max 256 characters)")
	}
	return nil
}

// SetupResult contains the result of a browser-based setup
type SetupResult struct {
	AccountName string
	ClientID    string
	AccountID   string
	Error       error
}

// SetupServer handles the browser-based authentication flow
type SetupServer struct {
	result        chan SetupResult
	shutdown      chan struct{}
	stopCleanup   chan struct{}
	pendingResult *SetupResult
	pendingMu     sync.Mutex
	csrfToken     string
	store         secrets.Store
	limiter       *rateLimiter
}

// NewSetupServer creates a new setup server
func NewSetupServer(store secrets.Store) (*SetupServer, error) {
	// Generate CSRF token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	stopCleanup := make(chan struct{})
	limiter := newRateLimiter(10, 15*time.Minute)
	limiter.startCleanup(5*time.Minute, stopCleanup)

	return &SetupServer{
		result:      make(chan SetupResult, 1),
		shutdown:    make(chan struct{}),
		stopCleanup: stopCleanup,
		csrfToken:   hex.EncodeToString(tokenBytes),
		store:       store,
		limiter:     limiter,
	}, nil
}

// Start starts the setup server and opens the browser
func (s *SetupServer) Start(ctx context.Context) (*SetupResult, error) {
	// Ensure cleanup goroutine is stopped when server exits
	defer close(s.stopCleanup)

	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleSetup)
	mux.HandleFunc("/validate", s.handleValidate)
	mux.HandleFunc("/submit", s.handleSubmit)
	mux.HandleFunc("/success", s.handleSuccess)
	mux.HandleFunc("/complete", s.handleComplete)

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start server in background
	go func() {
		_ = server.Serve(listener)
	}()

	// Open browser
	go func() {
		if err := openBrowser(baseURL); err != nil {
			slog.Info("failed to open browser, user can navigate manually", "url", baseURL)
		}
	}()

	// Wait for result or context cancellation
	select {
	case result := <-s.result:
		_ = server.Shutdown(context.Background())
		return &result, nil
	case <-ctx.Done():
		_ = server.Shutdown(context.Background())
		return nil, ctx.Err()
	case <-s.shutdown:
		_ = server.Shutdown(context.Background())
		s.pendingMu.Lock()
		defer s.pendingMu.Unlock()
		if s.pendingResult != nil {
			return s.pendingResult, nil
		}
		return nil, fmt.Errorf("setup cancelled")
	}
}

// handleSetup serves the main setup page
func (s *SetupServer) handleSetup(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.New("setup").Parse(setupTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	data := map[string]string{
		"CSRFToken": s.csrfToken,
	}

	// Set security headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")

	if err := tmpl.Execute(w, data); err != nil {
		slog.Error("setup template execution failed", "error", err)
	}
}

// validateCredentials tests credentials against the Airwallex API
func (s *SetupServer) validateCredentials(ctx context.Context, accountName, clientID, apiKey, accountID string) error {
	if accountName == "" || clientID == "" || apiKey == "" {
		return fmt.Errorf("account name, Client ID, and API Key are required")
	}

	var client *api.Client
	var err error
	if accountID != "" {
		client, err = api.NewClientWithAccount(clientID, apiKey, accountID)
	} else {
		client, err = api.NewClient(clientID, apiKey)
	}
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	if _, err := client.Get(ctx, "/api/v1/balances/current"); err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	return nil
}

// handleValidate tests credentials without saving
func (s *SetupServer) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify CSRF token FIRST (before rate limiting)
	providedToken := r.Header.Get("X-CSRF-Token")
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(s.csrfToken)) != 1 {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	// Check rate limit per client IP
	clientIP := getClientIP(r)
	if err := s.limiter.check(clientIP, "/validate"); err != nil {
		writeJSON(w, http.StatusTooManyRequests, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var req struct {
		AccountName string `json:"account_name"`
		ClientID    string `json:"client_id"`
		APIKey      string `json:"api_key"`
		AccountID   string `json:"account_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Normalize inputs
	req.AccountName = strings.TrimSpace(req.AccountName)
	req.ClientID = strings.TrimSpace(req.ClientID)
	req.APIKey = strings.TrimSpace(req.APIKey)
	req.AccountID = strings.TrimSpace(req.AccountID)

	// Validate input format
	if err := ValidateAccountName(req.AccountName); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if err := ValidateClientID(req.ClientID); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if err := ValidateAPIKey(req.APIKey); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Validate credentials
	if err := s.validateCredentials(r.Context(), req.AccountName, req.ClientID, req.APIKey, req.AccountID); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Connection successful!",
	})
}

// handleSubmit saves credentials after validation
func (s *SetupServer) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify CSRF token FIRST (before rate limiting)
	providedToken := r.Header.Get("X-CSRF-Token")
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(s.csrfToken)) != 1 {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	// Check rate limit per client IP
	clientIP := getClientIP(r)
	if err := s.limiter.check(clientIP, "/submit"); err != nil {
		writeJSON(w, http.StatusTooManyRequests, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var req struct {
		AccountName string `json:"account_name"`
		ClientID    string `json:"client_id"`
		APIKey      string `json:"api_key"`
		AccountID   string `json:"account_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Normalize inputs
	req.AccountName = strings.TrimSpace(req.AccountName)
	req.ClientID = strings.TrimSpace(req.ClientID)
	req.APIKey = strings.TrimSpace(req.APIKey)
	req.AccountID = strings.TrimSpace(req.AccountID)

	// Validate input format
	if err := ValidateAccountName(req.AccountName); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if err := ValidateClientID(req.ClientID); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if err := ValidateAPIKey(req.APIKey); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Validate credentials
	if err := s.validateCredentials(r.Context(), req.AccountName, req.ClientID, req.APIKey, req.AccountID); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Save to keychain
	err := s.store.Set(req.AccountName, secrets.Credentials{
		ClientID:  req.ClientID,
		APIKey:    req.APIKey,
		AccountID: req.AccountID,
	})
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   "Failed to save credentials to secure storage",
		})
		return
	}

	// Store pending result
	s.pendingMu.Lock()
	s.pendingResult = &SetupResult{
		AccountName: req.AccountName,
		ClientID:    req.ClientID,
		AccountID:   req.AccountID,
	}
	s.pendingMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"success":      true,
		"account_name": req.AccountName,
	})
}

// handleSuccess serves the success page
func (s *SetupServer) handleSuccess(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("success").Parse(successTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Use server state instead of URL parameter to prevent spoofing
	s.pendingMu.Lock()
	accountName := ""
	if s.pendingResult != nil {
		accountName = s.pendingResult.AccountName
	}
	s.pendingMu.Unlock()

	data := map[string]string{
		"AccountName": accountName,
		"CSRFToken":   s.csrfToken,
	}

	// Set security headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")

	if err := tmpl.Execute(w, data); err != nil {
		slog.Error("success template execution failed", "error", err)
	}
}

// handleComplete signals that setup is done
func (s *SetupServer) handleComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify CSRF token
	providedToken := r.Header.Get("X-CSRF-Token")
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(s.csrfToken)) != 1 {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	s.pendingMu.Lock()
	if s.pendingResult != nil {
		s.result <- *s.pendingResult
	}
	s.pendingMu.Unlock()
	close(s.shutdown)
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("JSON encoding failed", "error", err)
	}
}

// openBrowser opens the URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
