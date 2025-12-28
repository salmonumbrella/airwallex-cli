package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCreateFinancialReport_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/finance/financial_reports/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "rpt_123",
			"type": "ACCOUNT_STATEMENT_REPORT",
			"status": "PENDING",
			"file_format": "PDF",
			"from_date": "2024-01-01",
			"to_date": "2024-01-31",
			"currencies": ["USD", "EUR"],
			"created_at": "2024-01-15T10:00:00Z"
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req := &CreateReportRequest{
		Type:       ReportTypeAccountStatement,
		FromDate:   "2024-01-01",
		ToDate:     "2024-01-31",
		FileFormat: FileFormatPDF,
		Currencies: []string{"USD", "EUR"},
	}

	report, err := c.CreateFinancialReport(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateFinancialReport() error: %v", err)
	}
	if report == nil {
		t.Fatal("report is nil")
	}
	if report.ID != "rpt_123" {
		t.Errorf("id = %q, want 'rpt_123'", report.ID)
	}
	if report.Type != ReportTypeAccountStatement {
		t.Errorf("type = %q, want %q", report.Type, ReportTypeAccountStatement)
	}
	if report.Status != ReportStatusPending {
		t.Errorf("status = %q, want %q", report.Status, ReportStatusPending)
	}
	if report.FileFormat != FileFormatPDF {
		t.Errorf("file_format = %q, want %q", report.FileFormat, FileFormatPDF)
	}
	if len(report.Currencies) != 2 {
		t.Errorf("currencies count = %d, want 2", len(report.Currencies))
	}
}

func TestCreateFinancialReport_WithTransactionTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "rpt_456",
			"type": "TRANSACTION_RECON_REPORT",
			"status": "PENDING",
			"file_format": "CSV",
			"from_date": "2024-01-01",
			"to_date": "2024-01-31",
			"transaction_types": ["PAYMENT", "REFUND"],
			"created_at": "2024-01-15T10:00:00Z"
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req := &CreateReportRequest{
		Type:             ReportTypeTransactionRecon,
		FromDate:         "2024-01-01",
		ToDate:           "2024-01-31",
		FileFormat:       FileFormatCSV,
		TransactionTypes: []string{"PAYMENT", "REFUND"},
	}

	report, err := c.CreateFinancialReport(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateFinancialReport() error: %v", err)
	}
	if report == nil {
		t.Fatal("report is nil")
	}
	if report.Type != ReportTypeTransactionRecon {
		t.Errorf("type = %q, want %q", report.Type, ReportTypeTransactionRecon)
	}
	if len(report.TransactionTypes) != 2 {
		t.Errorf("transaction_types count = %d, want 2", len(report.TransactionTypes))
	}
}

func TestCreateFinancialReport_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "invalid_request",
			"message": "Invalid date range"
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req := &CreateReportRequest{
		Type:       ReportTypeAccountStatement,
		FromDate:   "2024-01-31",
		ToDate:     "2024-01-01", // Invalid: from_date after to_date
		FileFormat: FileFormatPDF,
	}

	_, err := c.CreateFinancialReport(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid request, got nil")
	}
	if !strings.Contains(err.Error(), "Invalid date range") {
		t.Errorf("error message = %q, want message containing 'Invalid date range'", err.Error())
	}
}

func TestListFinancialReports_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/finance/financial_reports" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify query parameters are correctly set
		pageNum := r.URL.Query().Get("page_num")
		pageSize := r.URL.Query().Get("page_size")

		if pageNum != "2" {
			t.Errorf("page_num = %q, want '2'", pageNum)
		}
		if pageSize != "10" {
			t.Errorf("page_size = %q, want '10'", pageSize)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "rpt_123",
					"type": "ACCOUNT_STATEMENT_REPORT",
					"status": "COMPLETED",
					"file_format": "PDF",
					"from_date": "2024-01-01",
					"to_date": "2024-01-31",
					"created_at": "2024-01-15T10:00:00Z",
					"report_expires_at": "2024-02-15T10:00:00Z"
				},
				{
					"id": "rpt_456",
					"type": "SETTLEMENT_REPORT",
					"status": "PENDING",
					"file_format": "CSV",
					"from_date": "2024-02-01",
					"to_date": "2024-02-29",
					"created_at": "2024-02-15T10:00:00Z"
				}
			],
			"has_more": true
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	result, err := c.ListFinancialReports(context.Background(), 2, 10)
	if err != nil {
		t.Fatalf("ListFinancialReports() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 2 {
		t.Errorf("items count = %d, want 2", len(result.Items))
	}
	if !result.HasMore {
		t.Error("has_more = false, want true")
	}
	if result.Items[0].ID != "rpt_123" {
		t.Errorf("items[0].id = %q, want 'rpt_123'", result.Items[0].ID)
	}
	if result.Items[0].Status != ReportStatusCompleted {
		t.Errorf("items[0].status = %q, want %q", result.Items[0].Status, ReportStatusCompleted)
	}
	if result.Items[1].Status != ReportStatusPending {
		t.Errorf("items[1].status = %q, want %q", result.Items[1].Status, ReportStatusPending)
	}
}

func TestListFinancialReports_WithoutPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/finance/financial_reports" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify no query parameters are set when values are 0 or negative
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query parameters, got: %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [],
			"has_more": false
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	result, err := c.ListFinancialReports(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("ListFinancialReports() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 0 {
		t.Errorf("items count = %d, want 0", len(result.Items))
	}
	if result.HasMore {
		t.Error("has_more = true, want false")
	}
}

func TestListFinancialReports_PartialPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify both page_num and page_size are set (API requires both)
		pageNum := r.URL.Query().Get("page_num")
		pageSize := r.URL.Query().Get("page_size")

		if pageNum != "1" {
			t.Errorf("page_num = %q, want '1'", pageNum)
		}
		if pageSize != "5" {
			t.Errorf("page_size = %q, want '5'", pageSize)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items": [], "has_more": false}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.ListFinancialReports(context.Background(), 0, 5)
	if err != nil {
		t.Fatalf("ListFinancialReports() error: %v", err)
	}
}

func TestListFinancialReports_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [],
			"has_more": false
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	result, err := c.ListFinancialReports(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ListFinancialReports() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 0 {
		t.Errorf("items count = %d, want 0", len(result.Items))
	}
}

func TestGetFinancialReport_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/finance/financial_reports/rpt_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "rpt_123",
			"type": "BALANCE_ACTIVITY_REPORT",
			"status": "COMPLETED",
			"file_format": "EXCEL",
			"from_date": "2024-01-01",
			"to_date": "2024-01-31",
			"currencies": ["USD", "EUR", "GBP"],
			"report_version": "v2",
			"created_at": "2024-01-15T10:00:00Z",
			"report_expires_at": "2024-02-15T10:00:00Z"
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	report, err := c.GetFinancialReport(context.Background(), "rpt_123")
	if err != nil {
		t.Fatalf("GetFinancialReport() error: %v", err)
	}
	if report == nil {
		t.Fatal("report is nil")
	}
	if report.ID != "rpt_123" {
		t.Errorf("id = %q, want 'rpt_123'", report.ID)
	}
	if report.Type != ReportTypeBalanceActivity {
		t.Errorf("type = %q, want %q", report.Type, ReportTypeBalanceActivity)
	}
	if report.Status != ReportStatusCompleted {
		t.Errorf("status = %q, want %q", report.Status, ReportStatusCompleted)
	}
	if report.FileFormat != FileFormatExcel {
		t.Errorf("file_format = %q, want %q", report.FileFormat, FileFormatExcel)
	}
	if len(report.Currencies) != 3 {
		t.Errorf("currencies count = %d, want 3", len(report.Currencies))
	}
	if report.ReportVersion != "v2" {
		t.Errorf("report_version = %q, want 'v2'", report.ReportVersion)
	}
}

func TestGetFinancialReport_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Report not found"
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.GetFinancialReport(context.Background(), "rpt_nonexistent")
	if err == nil {
		t.Error("expected error for not found report, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message = %q, want message containing 'not found'", err.Error())
	}
}

func TestGetFinancialReport_InvalidID(t *testing.T) {
	c := &Client{
		baseURL:        "http://test.example.com",
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.GetFinancialReport(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty report ID, got nil")
	}

	_, err = c.GetFinancialReport(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid report ID, got nil")
	}
}

func TestGetFinancialReport_FailedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "rpt_failed",
			"type": "SETTLEMENT_REPORT",
			"status": "FAILED",
			"file_format": "PDF",
			"from_date": "2024-01-01",
			"to_date": "2024-01-31",
			"created_at": "2024-01-15T10:00:00Z",
			"error_message": "Insufficient data for report generation"
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	report, err := c.GetFinancialReport(context.Background(), "rpt_failed")
	if err != nil {
		t.Fatalf("GetFinancialReport() error: %v", err)
	}
	if report == nil {
		t.Fatal("report is nil")
	}
	if report.Status != ReportStatusFailed {
		t.Errorf("status = %q, want %q", report.Status, ReportStatusFailed)
	}
	if report.ErrorMessage != "Insufficient data for report generation" {
		t.Errorf("error_message = %q, want 'Insufficient data for report generation'", report.ErrorMessage)
	}
}

func TestDownloadFinancialReport_Success(t *testing.T) {
	expectedContent := []byte("PDF file content here")
	expectedContentType := "application/pdf"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/finance/financial_reports/rpt_123/content" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", expectedContentType)
		_, _ = w.Write(expectedContent)
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	content, contentType, err := c.DownloadFinancialReport(context.Background(), "rpt_123")
	if err != nil {
		t.Fatalf("DownloadFinancialReport() error: %v", err)
	}
	if content == nil {
		t.Fatal("content is nil")
	}
	if string(content) != string(expectedContent) {
		t.Errorf("content = %q, want %q", string(content), string(expectedContent))
	}
	if contentType != expectedContentType {
		t.Errorf("content_type = %q, want %q", contentType, expectedContentType)
	}
}

func TestDownloadFinancialReport_CSVFormat(t *testing.T) {
	expectedContent := []byte("col1,col2,col3\nval1,val2,val3")
	expectedContentType := "text/csv"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", expectedContentType)
		_, _ = w.Write(expectedContent)
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	content, contentType, err := c.DownloadFinancialReport(context.Background(), "rpt_csv")
	if err != nil {
		t.Fatalf("DownloadFinancialReport() error: %v", err)
	}
	if string(content) != string(expectedContent) {
		t.Errorf("content = %q, want %q", string(content), string(expectedContent))
	}
	if contentType != expectedContentType {
		t.Errorf("content_type = %q, want %q", contentType, expectedContentType)
	}
}

func TestDownloadFinancialReport_InvalidID(t *testing.T) {
	c := &Client{
		baseURL:        "http://test.example.com",
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, _, err := c.DownloadFinancialReport(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty report ID, got nil")
	}

	_, _, err = c.DownloadFinancialReport(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid report ID, got nil")
	}
}

func TestDownloadFinancialReport_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Report not found or not yet available"
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, _, err := c.DownloadFinancialReport(context.Background(), "rpt_nonexistent")
	if err == nil {
		t.Error("expected error for not found report, got nil")
	}
}

func TestWaitForReport_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/finance/financial_reports/rpt_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")

		callCount++
		// First 2 calls: PENDING, third call: COMPLETED
		if callCount < 3 {
			_, _ = w.Write([]byte(`{
				"id": "rpt_123",
				"type": "ACCOUNT_STATEMENT_REPORT",
				"status": "PENDING",
				"file_format": "PDF",
				"from_date": "2024-01-01",
				"to_date": "2024-01-31",
				"created_at": "2024-01-15T10:00:00Z"
			}`))
		} else {
			_, _ = w.Write([]byte(`{
				"id": "rpt_123",
				"type": "ACCOUNT_STATEMENT_REPORT",
				"status": "COMPLETED",
				"file_format": "PDF",
				"from_date": "2024-01-01",
				"to_date": "2024-01-31",
				"created_at": "2024-01-15T10:00:00Z",
				"report_expires_at": "2024-02-15T10:00:00Z"
			}`))
		}
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	report, err := c.WaitForReport(context.Background(), "rpt_123", 30*time.Second)
	if err != nil {
		t.Fatalf("WaitForReport() error: %v", err)
	}
	if report == nil {
		t.Fatal("report is nil")
	}
	if report.Status != ReportStatusCompleted {
		t.Errorf("status = %q, want %q", report.Status, ReportStatusCompleted)
	}
	if callCount < 3 {
		t.Errorf("expected at least 3 API calls, got %d", callCount)
	}
}

func TestWaitForReport_Failed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "rpt_123",
			"type": "ACCOUNT_STATEMENT_REPORT",
			"status": "FAILED",
			"file_format": "PDF",
			"from_date": "2024-01-01",
			"to_date": "2024-01-31",
			"created_at": "2024-01-15T10:00:00Z",
			"error_message": "Data processing error"
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.WaitForReport(context.Background(), "rpt_123", 10*time.Second)
	if err == nil {
		t.Error("expected error for failed report, got nil")
	}
	if !strings.Contains(err.Error(), "Data processing error") {
		t.Errorf("error message = %q, want message containing 'Data processing error'", err.Error())
	}
}

func TestWaitForReport_FailedNoMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "rpt_123",
			"type": "ACCOUNT_STATEMENT_REPORT",
			"status": "FAILED",
			"file_format": "PDF",
			"from_date": "2024-01-01",
			"to_date": "2024-01-31",
			"created_at": "2024-01-15T10:00:00Z"
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.WaitForReport(context.Background(), "rpt_123", 10*time.Second)
	if err == nil {
		t.Error("expected error for failed report, got nil")
	}
	if !strings.Contains(err.Error(), "report generation failed") {
		t.Errorf("error message = %q, want message containing 'report generation failed'", err.Error())
	}
}

func TestWaitForReport_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Always return PENDING
		_, _ = w.Write([]byte(`{
			"id": "rpt_123",
			"type": "ACCOUNT_STATEMENT_REPORT",
			"status": "PENDING",
			"file_format": "PDF",
			"from_date": "2024-01-01",
			"to_date": "2024-01-31",
			"created_at": "2024-01-15T10:00:00Z"
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Use short timeout to speed up test
	_, err := c.WaitForReport(context.Background(), "rpt_123", 3*time.Second)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("error message = %q, want message containing 'timeout'", err.Error())
	}
}

func TestWaitForReport_InvalidID(t *testing.T) {
	c := &Client{
		baseURL:        "http://test.example.com",
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.WaitForReport(context.Background(), "", 10*time.Second)
	if err == nil {
		t.Error("expected error for empty report ID, got nil")
	}

	_, err = c.WaitForReport(context.Background(), "invalid/id", 10*time.Second)
	if err == nil {
		t.Error("expected error for invalid report ID, got nil")
	}
}

func TestWaitForReport_UnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "rpt_123",
			"type": "ACCOUNT_STATEMENT_REPORT",
			"status": "UNKNOWN_STATUS",
			"file_format": "PDF",
			"from_date": "2024-01-01",
			"to_date": "2024-01-31",
			"created_at": "2024-01-15T10:00:00Z"
		}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.WaitForReport(context.Background(), "rpt_123", 10*time.Second)
	if err == nil {
		t.Error("expected error for unexpected status, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected report status") {
		t.Errorf("error message = %q, want message containing 'unexpected report status'", err.Error())
	}
}

func TestWaitForReport_ExponentialBackoff(t *testing.T) {
	callTimes := []time.Time{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callTimes = append(callTimes, time.Now())

		w.Header().Set("Content-Type", "application/json")
		// Complete on 4th call
		if len(callTimes) < 4 {
			_, _ = w.Write([]byte(`{
				"id": "rpt_123",
				"type": "ACCOUNT_STATEMENT_REPORT",
				"status": "PENDING",
				"file_format": "PDF",
				"from_date": "2024-01-01",
				"to_date": "2024-01-31",
				"created_at": "2024-01-15T10:00:00Z"
			}`))
		} else {
			_, _ = w.Write([]byte(`{
				"id": "rpt_123",
				"type": "ACCOUNT_STATEMENT_REPORT",
				"status": "COMPLETED",
				"file_format": "PDF",
				"from_date": "2024-01-01",
				"to_date": "2024-01-31",
				"created_at": "2024-01-15T10:00:00Z",
				"report_expires_at": "2024-02-15T10:00:00Z"
			}`))
		}
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.WaitForReport(context.Background(), "rpt_123", 30*time.Second)
	if err != nil {
		t.Fatalf("WaitForReport() error: %v", err)
	}

	if len(callTimes) < 4 {
		t.Fatalf("expected at least 4 calls, got %d", len(callTimes))
	}

	// Verify exponential backoff: 2s, 4s, 8s (with some tolerance for execution time)
	// Call 1 -> Call 2: ~2s
	// Call 2 -> Call 3: ~4s
	// Call 3 -> Call 4: ~8s
	delay1 := callTimes[1].Sub(callTimes[0])
	delay2 := callTimes[2].Sub(callTimes[1])
	delay3 := callTimes[3].Sub(callTimes[2])

	// Allow 500ms tolerance
	tolerance := 500 * time.Millisecond

	if delay1 < 2*time.Second-tolerance || delay1 > 2*time.Second+tolerance {
		t.Errorf("first delay = %v, want ~2s", delay1)
	}
	if delay2 < 4*time.Second-tolerance || delay2 > 4*time.Second+tolerance {
		t.Errorf("second delay = %v, want ~4s", delay2)
	}
	if delay3 < 8*time.Second-tolerance || delay3 > 8*time.Second+tolerance {
		t.Errorf("third delay = %v, want ~8s", delay3)
	}
}
