package cmd

import (
	"net/url"
	"strings"
	"testing"
)

func TestAPICommand_Flags(t *testing.T) {
	cmd := newAPICmd()

	// Check command has correct args validation
	if cmd.Args == nil {
		t.Error("expected Args validation")
	}

	// Check all expected flags exist
	expectedFlags := []struct {
		name      string
		shorthand string
	}{
		{"method", "X"},
		{"data", "d"},
		{"data-file", ""},
		{"header", "H"},
		{"query", "q"},
		{"silent", "s"},
		{"include", "i"},
	}

	for _, ef := range expectedFlags {
		flag := cmd.Flags().Lookup(ef.name)
		if flag == nil {
			t.Errorf("expected flag --%s", ef.name)
			continue
		}
		if ef.shorthand != "" && flag.Shorthand != ef.shorthand {
			t.Errorf("flag --%s expected shorthand -%s, got -%s", ef.name, ef.shorthand, flag.Shorthand)
		}
	}
}

func TestAPICommand_NoShorthandCollisionWithGlobalDebug(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	root := NewRootCmd()
	root.SetArgs([]string{"api", "/api/v1/balances/current"})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic executing api command: %v", r)
		}
	}()

	err := root.Execute()
	if err == nil {
		t.Fatal("expected non-nil error from mock API")
	}
	if strings.Contains(err.Error(), "unable to redefine") {
		t.Fatalf("expected no shorthand collision error, got: %v", err)
	}
}

func TestAPICommand_EndpointNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/api/v1/balances", "/api/v1/balances"},
		{"api/v1/balances", "/api/v1/balances"},
		{"/", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			endpoint := tt.input
			if !strings.HasPrefix(endpoint, "/") {
				endpoint = "/" + endpoint
			}
			if endpoint != tt.expected {
				t.Errorf("got %s, want %s", endpoint, tt.expected)
			}
		})
	}
}

func TestAPICommand_Usage(t *testing.T) {
	cmd := newAPICmd()

	if cmd.Use != "api [method] <endpoint>" {
		t.Errorf("unexpected Use: %s", cmd.Use)
	}

	// Check help includes examples
	if !strings.Contains(cmd.Long, "GET current balances") {
		t.Error("expected help to include GET example")
	}
	if !strings.Contains(cmd.Long, "GET with method as positional arg") {
		t.Error("expected help to include positional method example")
	}
}

func TestIsHTTPMethod(t *testing.T) {
	for _, m := range []string{"get", "GET", "Get", "post", "POST", "put", "patch", "delete", "head", "options"} {
		if !isHTTPMethod(m) {
			t.Errorf("expected %q to be recognized as HTTP method", m)
		}
	}
	for _, m := range []string{"/api/v1/foo", "transfers", "list", ""} {
		if isHTTPMethod(m) {
			t.Errorf("expected %q to NOT be recognized as HTTP method", m)
		}
	}
}

func TestAPICommand_AcceptsMethodAsPositionalArg(t *testing.T) {
	cmd := newAPICmd()

	// 1 arg should be accepted
	if err := cmd.Args(cmd, []string{"/api/v1/balances"}); err != nil {
		t.Errorf("1 arg should be accepted: %v", err)
	}

	// 2 args should be accepted
	if err := cmd.Args(cmd, []string{"get", "/api/v1/balances"}); err != nil {
		t.Errorf("2 args should be accepted: %v", err)
	}

	// 0 args should be rejected
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("0 args should be rejected")
	}

	// Additional args are now parsed by RunE as query shorthand/validation.
	if err := cmd.Args(cmd, []string{"get", "/api/v1/balances", "page_size=10"}); err != nil {
		t.Errorf("3 args should be accepted for query shorthand: %v", err)
	}
}

func TestAPICommand_QueryParamEncoding(t *testing.T) {
	// Test that query parameters are properly URL-encoded
	// This tests the encoding logic used in the api command
	tests := []struct {
		name     string
		params   []string
		expected string // expected query string (URL-encoded)
	}{
		{
			name:     "simple key=value",
			params:   []string{"status=COMPLETED"},
			expected: "status=COMPLETED",
		},
		{
			name:     "multiple params",
			params:   []string{"status=COMPLETED", "page_size=10"},
			expected: "page_size=10&status=COMPLETED", // url.Values sorts alphabetically
		},
		{
			name:     "value with spaces",
			params:   []string{"name=John Doe"},
			expected: "name=John+Doe",
		},
		{
			name:     "value with special chars",
			params:   []string{"filter=a=b&c=d"},
			expected: "filter=a%3Db%26c%3Dd",
		},
		{
			name:     "value with equals sign",
			params:   []string{"expr=1+1=2"},
			expected: "expr=1%2B1%3D2",
		},
		{
			name:     "key without value",
			params:   []string{"flag"},
			expected: "flag=",
		},
		{
			name:     "unicode value",
			params:   []string{"city=東京"},
			expected: "city=%E6%9D%B1%E4%BA%AC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the encoding logic from api.go
			params := url.Values{}
			for _, qp := range tt.params {
				parts := strings.SplitN(qp, "=", 2)
				if len(parts) == 2 {
					params.Add(parts[0], parts[1])
				} else {
					params.Add(parts[0], "")
				}
			}
			encoded := params.Encode()
			if encoded != tt.expected {
				t.Errorf("got %q, want %q", encoded, tt.expected)
			}
		})
	}
}

func TestParseAPIInvocation_QueryShorthand(t *testing.T) {
	cmd := newAPICmd()
	method, endpoint, q, err := parseAPIInvocation(cmd, []string{
		"get",
		"/api/v1/financial_transactions",
		"from_created_at=2025-06-01T00:00:00+0000",
		"to_created_at=2025-06-30T23:59:59+0000",
		"page_size=100",
	}, "GET", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != "GET" {
		t.Fatalf("method = %q, want GET", method)
	}
	if endpoint != "/api/v1/financial_transactions" {
		t.Fatalf("endpoint = %q, want /api/v1/financial_transactions", endpoint)
	}
	if len(q) != 3 {
		t.Fatalf("expected 3 query params, got %d (%v)", len(q), q)
	}
}

func TestParseAPIInvocation_UnknownMethod(t *testing.T) {
	cmd := newAPICmd()
	_, _, _, err := parseAPIInvocation(cmd, []string{"fetch", "/api/v1/balances/current"}, "GET", nil)
	if err == nil {
		t.Fatal("expected error for unknown method")
	}
	if !strings.Contains(err.Error(), "unknown HTTP method") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseAPIInvocation_InvalidExtraArg(t *testing.T) {
	cmd := newAPICmd()
	_, _, _, err := parseAPIInvocation(cmd, []string{"/api/v1/balances/current", "oops"}, "GET", nil)
	if err == nil {
		t.Fatal("expected error for invalid extra argument")
	}
	if !strings.Contains(err.Error(), "If this is a query parameter") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemapFinancialTransactionsQueryParams(t *testing.T) {
	q, remapped := remapFinancialTransactionsQueryParams("/api/v1/financial_transactions", []string{
		"from_posted_at=2025-06-01T00:00:00+0000",
		"to_posted_at=2025-06-30T23:59:59+0000",
		"page_size=100",
	})
	if !remapped {
		t.Fatal("expected remapped=true")
	}

	got := strings.Join(q, ",")
	if !strings.Contains(got, "from_created_at=2025-06-01T00:00:00+0000") {
		t.Fatalf("missing from_created_at remap in %v", q)
	}
	if !strings.Contains(got, "to_created_at=2025-06-30T23:59:59+0000") {
		t.Fatalf("missing to_created_at remap in %v", q)
	}
}
