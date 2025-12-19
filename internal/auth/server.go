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
	pendingResult *SetupResult
	pendingMu     sync.Mutex
	csrfToken     string
	store         secrets.Store
}

// NewSetupServer creates a new setup server
func NewSetupServer(store secrets.Store) (*SetupServer, error) {
	// Generate CSRF token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	return &SetupServer{
		result:    make(chan SetupResult, 1),
		shutdown:  make(chan struct{}),
		csrfToken: hex.EncodeToString(tokenBytes),
		store:     store,
	}, nil
}

// Start starts the setup server and opens the browser
func (s *SetupServer) Start(ctx context.Context) (*SetupResult, error) {
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
		return fmt.Errorf("Account name, Client ID, and API Key are required")
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
		return fmt.Errorf("Connection failed: %v", err)
	}

	return nil
}

// handleValidate tests credentials without saving
func (s *SetupServer) handleValidate(w http.ResponseWriter, r *http.Request) {
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

	// Verify CSRF token
	providedToken := r.Header.Get("X-CSRF-Token")
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(s.csrfToken)) != 1 {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
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
