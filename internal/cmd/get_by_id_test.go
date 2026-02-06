package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
)

func TestFetchByID_CardVsCardholder(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Register mock handlers for card and cardholder endpoints.
	// We return minimal JSON so the response parses successfully.
	testMockServer.Handle("GET", "/api/v1/issuing/cards/card_abc123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"card_id": "card_abc123"})
	})
	testMockServer.Handle("GET", "/api/v1/issuing/cardholders/card_holder_abc123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"cardholder_id": "card_holder_abc123"})
	})
	testMockServer.Handle("GET", "/api/v1/issuing/cardholders/cardholder_xyz789", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"cardholder_id": "cardholder_xyz789"})
	})

	tests := []struct {
		name        string
		id          string
		wantCmdPfx  string // prefix of the canonical command
		wantErrPart string // if non-empty, expect error containing this
	}{
		{
			name:       "card_holder_ prefix routes to GetCardholder",
			id:         "card_holder_abc123",
			wantCmdPfx: "airwallex cardholders get card_holder_abc123",
		},
		{
			name:       "cardholder_ prefix routes to GetCardholder",
			id:         "cardholder_xyz789",
			wantCmdPfx: "airwallex cardholders get cardholder_xyz789",
		},
		{
			name:       "card_ prefix (not card_holder_) routes to GetCard",
			id:         "card_abc123",
			wantCmdPfx: "airwallex cards get card_abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := api.NewClientWithBaseURL(testMockServer.URL(), "test-client-id", "test-api-key")
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			_, canonicalCmd, err := fetchByID(context.Background(), client, tt.id)

			if tt.wantErrPart != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrPart)
				}
				if !strings.Contains(err.Error(), tt.wantErrPart) {
					t.Errorf("expected error containing %q, got %q", tt.wantErrPart, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if canonicalCmd != tt.wantCmdPfx {
				t.Errorf("canonical command = %q, want %q", canonicalCmd, tt.wantCmdPfx)
			}
		})
	}
}

func TestFetchByID_PrefixRouting(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Register mock handlers for a selection of resource types.
	handlers := map[string]string{
		"/api/v1/transfers/tfr_001":                     `{"id":"tfr_001"}`,
		"/api/v1/beneficiaries/ben_002":                 `{"id":"ben_002"}`,
		"/api/v1/issuing/transactions/txn_003":          `{"id":"txn_003"}`,
		"/api/v1/issuing/transaction_disputes/disp_004": `{"id":"disp_004"}`,
	}
	for path, body := range handlers {
		b := body // capture
		testMockServer.Handle("GET", path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(b))
		})
	}

	tests := []struct {
		name    string
		id      string
		wantCmd string
	}{
		{"transfer", "tfr_001", "airwallex transfers get tfr_001"},
		{"beneficiary", "ben_002", "airwallex beneficiaries get ben_002"},
		{"transaction", "txn_003", "airwallex transactions get txn_003"},
		{"dispute", "disp_004", "airwallex disputes get disp_004"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := api.NewClientWithBaseURL(testMockServer.URL(), "test-client-id", "test-api-key")
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			_, canonicalCmd, err := fetchByID(context.Background(), client, tt.id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if canonicalCmd != tt.wantCmd {
				t.Errorf("canonical command = %q, want %q", canonicalCmd, tt.wantCmd)
			}
		})
	}
}

func TestFetchByID_UnknownPrefix(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	client, err := api.NewClientWithBaseURL(testMockServer.URL(), "test-client-id", "test-api-key")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, _, err = fetchByID(context.Background(), client, "zzz_unknown_id")
	if err == nil {
		t.Fatal("expected error for unknown prefix, got nil")
	}
	if !strings.Contains(err.Error(), "unknown id") {
		t.Errorf("expected error containing %q, got %q", "unknown id", err.Error())
	}
}
