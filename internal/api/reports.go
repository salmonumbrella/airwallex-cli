package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"
)

// Report types
const (
	ReportTypeAccountStatement = "ACCOUNT_STATEMENT_REPORT"
	ReportTypeBalanceActivity  = "BALANCE_ACTIVITY_REPORT"
	ReportTypeTransactionRecon = "TRANSACTION_RECON_REPORT"
	ReportTypeSettlement       = "SETTLEMENT_REPORT"
)

// File formats
const (
	FileFormatPDF   = "PDF"
	FileFormatCSV   = "CSV"
	FileFormatExcel = "EXCEL"
)

// Report statuses
const (
	ReportStatusPending   = "PENDING"
	ReportStatusCompleted = "COMPLETED"
	ReportStatusFailed    = "FAILED"
)

type FinancialReport struct {
	ID               string   `json:"id"`
	Type             string   `json:"type"`
	Status           string   `json:"status"`
	FileFormat       string   `json:"file_format"`
	FromDate         string   `json:"from_date"`
	ToDate           string   `json:"to_date"`
	Currencies       []string `json:"currencies,omitempty"`
	TransactionTypes []string `json:"transaction_types,omitempty"`
	ReportVersion    string   `json:"report_version,omitempty"`
	CreatedAt        string   `json:"created_at"`
	ReportExpiresAt  string   `json:"report_expires_at,omitempty"`
	ErrorMessage     string   `json:"error_message,omitempty"`
}

type FinancialReportsResponse struct {
	Items   []FinancialReport `json:"items"`
	HasMore bool              `json:"has_more"`
}

type CreateReportRequest struct {
	Type             string   `json:"type"`
	FromDate         string   `json:"from_date"`
	ToDate           string   `json:"to_date"`
	FileFormat       string   `json:"file_format"`
	Currencies       []string `json:"currencies,omitempty"`
	TransactionTypes []string `json:"transaction_types,omitempty"`
	ReportVersion    string   `json:"report_version,omitempty"`
	TimeZone         string   `json:"time_zone,omitempty"`
}

// CreateFinancialReport creates a new financial report (async)
func (c *Client) CreateFinancialReport(ctx context.Context, req *CreateReportRequest) (*FinancialReport, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, "/api/v1/finance/financial_reports/create", req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var report FinancialReport
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return nil, err
	}
	return &report, nil
}

// ListFinancialReports lists all financial reports
func (c *Client) ListFinancialReports(ctx context.Context, pageNum, pageSize int) (*FinancialReportsResponse, error) {
	params := url.Values{}
	// Airwallex API requires both page_num and page_size together
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1 // API uses 1-based page numbering
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v1/finance/financial_reports"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var result FinancialReportsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFinancialReport gets a single report by ID
func (c *Client) GetFinancialReport(ctx context.Context, reportID string) (*FinancialReport, error) {
	if err := ValidateResourceID(reportID, "report"); err != nil {
		return nil, err
	}
	resp, err := c.Get(ctx, "/api/v1/finance/financial_reports/"+url.PathEscape(reportID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var report FinancialReport
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return nil, err
	}
	return &report, nil
}

// DownloadFinancialReport downloads report content as bytes
// Returns: (content bytes, content-type header, error)
func (c *Client) DownloadFinancialReport(ctx context.Context, reportID string) ([]byte, string, error) {
	if err := ValidateResourceID(reportID, "report"); err != nil {
		return nil, "", err
	}
	resp, err := c.Get(ctx, "/api/v1/finance/financial_reports/"+url.PathEscape(reportID)+"/content")
	if err != nil {
		return nil, "", err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", ParseAPIError(body)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read report content: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	return content, contentType, nil
}

// WaitForReport polls until report is complete or failed (helper method)
func (c *Client) WaitForReport(ctx context.Context, reportID string, timeout time.Duration) (*FinancialReport, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	if err := ValidateResourceID(reportID, "report"); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	attempt := 0
	maxAttempts := int(timeout / (2 * time.Second))

	for time.Now().Before(deadline) {
		report, err := c.GetFinancialReport(ctx, reportID)
		if err != nil {
			return nil, err
		}

		switch report.Status {
		case ReportStatusCompleted:
			return report, nil
		case ReportStatusFailed:
			if report.ErrorMessage != "" {
				return report, fmt.Errorf("report generation failed: %s", report.ErrorMessage)
			}
			return report, fmt.Errorf("report generation failed")
		case ReportStatusPending:
			// Continue polling
		default:
			return report, fmt.Errorf("unexpected report status: %s", report.Status)
		}

		// Exponential backoff: 2s, 4s, 8s, then stay at 8s
		delay := time.Duration(1<<min(attempt, 2)) * 2 * time.Second
		if time.Now().Add(delay).After(deadline) {
			delay = time.Until(deadline)
			if delay <= 0 {
				break
			}
		}

		time.Sleep(delay)
		attempt++

		if attempt > maxAttempts {
			break
		}
	}

	return nil, fmt.Errorf("timeout waiting for report to complete")
}
