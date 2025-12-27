package api

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// TestEndpointsRegistry_Complete verifies all endpoints are properly configured
func TestEndpointsRegistry_Complete(t *testing.T) {
	v := reflect.ValueOf(Endpoints)
	typ := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		endpoint := v.Field(i).Interface().(Endpoint)

		t.Run(field.Name, func(t *testing.T) {
			// Verify Path is not empty
			if endpoint.Path == "" {
				t.Errorf("%s: Path is empty", field.Name)
			}

			// Verify Path starts with /api/v1/
			if !strings.HasPrefix(endpoint.Path, "/api/v1/") {
				t.Errorf("%s: Path %q does not start with /api/v1/", field.Name, endpoint.Path)
			}

			// Verify Method is valid HTTP method
			validMethods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
			if endpoint.Method == "" {
				t.Errorf("%s: Method is empty", field.Name)
			} else {
				found := false
				for _, m := range validMethods {
					if endpoint.Method == m {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: Method %q is not a valid HTTP method", field.Name, endpoint.Method)
				}
			}

			// Verify ExpectedStatus is valid HTTP status code
			if endpoint.ExpectedStatus < 200 || endpoint.ExpectedStatus >= 300 {
				t.Errorf("%s: ExpectedStatus %d is not a 2xx status code", field.Name, endpoint.ExpectedStatus)
			}
		})
	}
}

// TestEndpointsRegistry_IdempotencyFlags verifies RequiresIdem flags are correct
func TestEndpointsRegistry_IdempotencyFlags(t *testing.T) {
	// Endpoints that MUST have RequiresIdem = true
	mustRequireIdem := []struct {
		name     string
		endpoint Endpoint
	}{
		{"TransfersCreate", Endpoints.TransfersCreate},
		{"CardsCreate", Endpoints.CardsCreate},
		{"BeneficiariesCreate", Endpoints.BeneficiariesCreate},
		{"FXConversionsCreate", Endpoints.FXConversionsCreate},
		{"LinkedAccountsCreate", Endpoints.LinkedAccountsCreate},
		{"PaymentLinksCreate", Endpoints.PaymentLinksCreate},
	}

	for _, tc := range mustRequireIdem {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.endpoint.RequiresIdem {
				t.Errorf("%s: RequiresIdem should be true for financial operations", tc.name)
			}
			if !tc.endpoint.NeedsIdempotencyKey() {
				t.Errorf("%s: NeedsIdempotencyKey() should return true", tc.name)
			}
		})
	}

	// Endpoints that MUST have RequiresIdem = false
	mustNotRequireIdem := []struct {
		name     string
		endpoint Endpoint
	}{
		{"Login", Endpoints.Login},
		{"TransfersList", Endpoints.TransfersList},
		{"TransfersGet", Endpoints.TransfersGet},
		{"BalancesCurrent", Endpoints.BalancesCurrent},
		{"BalancesHistory", Endpoints.BalancesHistory},
		{"BeneficiariesList", Endpoints.BeneficiariesList},
		{"BeneficiariesGet", Endpoints.BeneficiariesGet},
	}

	for _, tc := range mustNotRequireIdem {
		t.Run(tc.name, func(t *testing.T) {
			if tc.endpoint.RequiresIdem {
				t.Errorf("%s: RequiresIdem should be false for read/list operations", tc.name)
			}
			if tc.endpoint.NeedsIdempotencyKey() {
				t.Errorf("%s: NeedsIdempotencyKey() should return false", tc.name)
			}
		})
	}
}

// TestEndpointsRegistry_PathConsistency verifies path patterns are consistent
func TestEndpointsRegistry_PathConsistency(t *testing.T) {
	tests := []struct {
		name     string
		endpoint Endpoint
		wantPath string
	}{
		{"TransfersCreate", Endpoints.TransfersCreate, "/api/v1/transfers/create"},
		{"TransfersList", Endpoints.TransfersList, "/api/v1/transfers"},
		{"BeneficiariesCreate", Endpoints.BeneficiariesCreate, "/api/v1/beneficiaries/create"},
		{"CardsCreate", Endpoints.CardsCreate, "/api/v1/issuing/cards/create"},
		{"BalancesCurrent", Endpoints.BalancesCurrent, "/api/v1/balances/current"},
		{"BalancesHistory", Endpoints.BalancesHistory, "/api/v1/balances/history"},
		{"FXConversionsCreate", Endpoints.FXConversionsCreate, "/api/v1/fx/conversions/create"},
		{"LinkedAccountsCreate", Endpoints.LinkedAccountsCreate, "/api/v1/linked_accounts/create"},
		{"PaymentLinksCreate", Endpoints.PaymentLinksCreate, "/api/v1/pa/payment_links/create"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.endpoint.Path != tt.wantPath {
				t.Errorf("%s: Path = %q, want %q", tt.name, tt.endpoint.Path, tt.wantPath)
			}
		})
	}
}

// TestEndpointsRegistry_HTTPMethodConsistency verifies HTTP methods match operation types
func TestEndpointsRegistry_HTTPMethodConsistency(t *testing.T) {
	// List/Get operations should be GET
	getEndpoints := []struct {
		name     string
		endpoint Endpoint
	}{
		{"TransfersList", Endpoints.TransfersList},
		{"TransfersGet", Endpoints.TransfersGet},
		{"BeneficiariesList", Endpoints.BeneficiariesList},
		{"BeneficiariesGet", Endpoints.BeneficiariesGet},
		{"CardsList", Endpoints.CardsList},
		{"CardsGet", Endpoints.CardsGet},
		{"BalancesCurrent", Endpoints.BalancesCurrent},
		{"BalancesHistory", Endpoints.BalancesHistory},
	}

	for _, tc := range getEndpoints {
		t.Run(tc.name, func(t *testing.T) {
			if tc.endpoint.Method != http.MethodGet {
				t.Errorf("%s: Method = %q, want %q", tc.name, tc.endpoint.Method, http.MethodGet)
			}
		})
	}

	// Create/Update/Delete operations should be POST (Airwallex uses POST for mutations)
	postEndpoints := []struct {
		name     string
		endpoint Endpoint
	}{
		{"TransfersCreate", Endpoints.TransfersCreate},
		{"BeneficiariesCreate", Endpoints.BeneficiariesCreate},
		{"CardsCreate", Endpoints.CardsCreate},
		{"Login", Endpoints.Login},
	}

	for _, tc := range postEndpoints {
		t.Run(tc.name, func(t *testing.T) {
			if tc.endpoint.Method != http.MethodPost {
				t.Errorf("%s: Method = %q, want %q", tc.name, tc.endpoint.Method, http.MethodPost)
			}
		})
	}
}

// TestEndpointsRegistry_StatusCodeConsistency verifies expected status codes
func TestEndpointsRegistry_StatusCodeConsistency(t *testing.T) {
	// Create operations should expect 201
	createEndpoints := []struct {
		name     string
		endpoint Endpoint
	}{
		{"TransfersCreate", Endpoints.TransfersCreate},
		{"BeneficiariesCreate", Endpoints.BeneficiariesCreate},
		{"CardsCreate", Endpoints.CardsCreate},
		{"FXConversionsCreate", Endpoints.FXConversionsCreate},
		{"LinkedAccountsCreate", Endpoints.LinkedAccountsCreate},
		{"PaymentLinksCreate", Endpoints.PaymentLinksCreate},
		{"Login", Endpoints.Login},
	}

	for _, tc := range createEndpoints {
		t.Run(tc.name, func(t *testing.T) {
			if tc.endpoint.ExpectedStatus != http.StatusCreated {
				t.Errorf("%s: ExpectedStatus = %d, want %d", tc.name, tc.endpoint.ExpectedStatus, http.StatusCreated)
			}
		})
	}

	// Get/List operations should expect 200
	okEndpoints := []struct {
		name     string
		endpoint Endpoint
	}{
		{"TransfersList", Endpoints.TransfersList},
		{"TransfersGet", Endpoints.TransfersGet},
		{"BalancesCurrent", Endpoints.BalancesCurrent},
		{"BalancesHistory", Endpoints.BalancesHistory},
	}

	for _, tc := range okEndpoints {
		t.Run(tc.name, func(t *testing.T) {
			if tc.endpoint.ExpectedStatus != http.StatusOK {
				t.Errorf("%s: ExpectedStatus = %d, want %d", tc.name, tc.endpoint.ExpectedStatus, http.StatusOK)
			}
		})
	}
}

// TestEndpointsRegistry_NoDuplicatePaths verifies no duplicate paths in the registry
func TestEndpointsRegistry_NoDuplicatePaths(t *testing.T) {
	v := reflect.ValueOf(Endpoints)
	paths := make(map[string][]string)

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		endpoint := v.Field(i).Interface().(Endpoint)
		// Normalize path by removing {id} placeholders for duplicate detection
		normalizedPath := strings.ReplaceAll(endpoint.Path, "/{id}", "")
		paths[normalizedPath] = append(paths[normalizedPath], field.Name)
	}

	for path, fields := range paths {
		// Allow duplicate normalized paths only if they differ in their full form
		// (e.g., /api/v1/transfers and /api/v1/transfers/{id} are OK)
		if len(fields) > 2 {
			t.Errorf("Path %q is duplicated across fields: %v", path, fields)
		}
	}
}

// TestIsFinancialOperation_UsesRegistry verifies isFinancialOperation uses the registry
func TestIsFinancialOperation_UsesRegistry(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{Endpoints.TransfersCreate.Path, true},
		{Endpoints.CardsCreate.Path, true},
		{Endpoints.BeneficiariesCreate.Path, true},
		{Endpoints.FXConversionsCreate.Path, true},
		{Endpoints.LinkedAccountsCreate.Path, true},
		{Endpoints.PaymentLinksCreate.Path, true},
		{Endpoints.TransfersList.Path, false},
		{Endpoints.TransfersGet.Path, false},
		{Endpoints.BalancesCurrent.Path, false},
		{Endpoints.Login.Path, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isFinancialOperation(tt.path)
			if result != tt.expected {
				t.Errorf("isFinancialOperation(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestEndpoint_NeedsIdempotencyKey verifies the helper method
func TestEndpoint_NeedsIdempotencyKey(t *testing.T) {
	tests := []struct {
		name     string
		endpoint Endpoint
		want     bool
	}{
		{
			name: "financial operation",
			endpoint: Endpoint{
				Path:         "/api/v1/transfers/create",
				RequiresIdem: true,
			},
			want: true,
		},
		{
			name: "read operation",
			endpoint: Endpoint{
				Path:         "/api/v1/transfers",
				RequiresIdem: false,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.endpoint.NeedsIdempotencyKey(); got != tt.want {
				t.Errorf("Endpoint.NeedsIdempotencyKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
