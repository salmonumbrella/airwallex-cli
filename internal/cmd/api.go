package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func newAPICmd() *cobra.Command {
	var (
		method      string
		data        string
		dataFile    string
		headers     []string
		queryParams []string
		silent      bool
		include     bool
	)

	cmd := &cobra.Command{
		Use:   "api [method] <endpoint>",
		Short: "Make raw API requests",
		Long: `Make authenticated requests to any Airwallex API endpoint.

The endpoint should start with / (e.g., /api/v1/balances/current).
Authentication is handled automatically using your configured account.

An optional HTTP method (get, post, put, patch, delete) can be provided
as the first argument instead of using -X.

Examples:
  # GET current balances
  airwallex api /api/v1/balances/current

  # GET with method as positional arg
  airwallex api get /api/v1/balances/current

  # GET with query parameters
  airwallex api /api/v1/transfers -q status=COMPLETED -q page_size=10

  # GET with query shorthand (extra key=value args are treated as -q)
  airwallex api get /api/v1/financial_transactions from_created_at=2025-06-01T00:00:00+0000 to_created_at=2025-06-30T23:59:59+0000 page_size=100

  # POST with inline JSON
  airwallex api post /api/v1/beneficiaries -d '{"beneficiary": {...}}'

  # POST with -X flag
  airwallex api /api/v1/beneficiaries -X POST -d '{"beneficiary": {...}}'

  # POST with file
  airwallex api post /api/v1/transfers --data-file transfer.json

  # Include response headers
  airwallex api /api/v1/balances/current -i`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedMethod, endpoint, resolvedQueryParams, err := parseAPIInvocation(cmd, args, method, queryParams)
			if err != nil {
				return err
			}
			method = resolvedMethod
			queryParams = resolvedQueryParams

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			// Build request body
			var body io.Reader
			if data != "" {
				body = strings.NewReader(data)
			} else if dataFile != "" {
				if dataFile == "-" {
					body = os.Stdin
				} else {
					f, err := os.Open(dataFile)
					if err != nil {
						return fmt.Errorf("failed to open data file: %w", err)
					}
					defer func() { _ = f.Close() }()
					body = f
				}
			}

			// Build URL with query params (properly encoded)
			reqURL := client.BaseURL() + endpoint
			if len(queryParams) > 0 {
				params := url.Values{}
				for _, qp := range queryParams {
					parts := strings.SplitN(qp, "=", 2)
					if len(parts) == 2 {
						params.Add(parts[0], parts[1])
					} else {
						// Handle key without value (e.g., "flag" becomes "flag=")
						params.Add(parts[0], "")
					}
				}
				reqURL += "?" + params.Encode()
			}

			// Create request
			req, err := http.NewRequestWithContext(cmd.Context(), method, reqURL, body)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			// Add custom headers
			for _, h := range headers {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) == 2 {
					req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				}
			}

			// Execute request
			resp, err := client.Do(cmd.Context(), req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			// Read response
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			if silent {
				// Still return error for non-2xx status codes
				if resp.StatusCode >= 400 {
					return fmt.Errorf("request failed with status %d", resp.StatusCode)
				}
				return nil
			}

			// Print headers if requested
			if include {
				errOut := cmd.ErrOrStderr()
				_, _ = fmt.Fprintf(errOut, "HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
				for k, v := range resp.Header {
					_, _ = fmt.Fprintf(errOut, "%s: %s\n", k, strings.Join(v, ", "))
				}
				_, _ = fmt.Fprintln(errOut)
			}

			// Output response body
			out := cmd.OutOrStdout()
			if outfmt.IsJSON(cmd.Context()) || isJSONResponse(resp) {
				// Pretty-print JSON
				var prettyJSON interface{}
				if err := json.Unmarshal(respBody, &prettyJSON); err == nil {
					if writeErr := outfmt.WriteJSONFiltered(out, prettyJSON, outfmt.GetQuery(cmd.Context())); writeErr != nil {
						return writeErr
					}
				} else {
					// Not valid JSON, output raw
					_, _ = fmt.Fprintln(out, string(respBody))
				}
			} else {
				// Raw output
				_, _ = fmt.Fprintln(out, string(respBody))
			}

			// Return error for non-2xx status codes
			if resp.StatusCode >= 400 {
				return fmt.Errorf("request failed with status %d", resp.StatusCode)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&method, "method", "X", "GET", "HTTP method")
	cmd.Flags().StringVarP(&data, "data", "d", "", "Request body (JSON)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Read request body from file (- for stdin)")
	cmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "Custom headers (key: value)")
	cmd.Flags().StringArrayVarP(&queryParams, "query", "q", nil, "Query parameters (key=value)")
	cmd.Flags().BoolVarP(&silent, "silent", "s", false, "Don't print response body")
	cmd.Flags().BoolVarP(&include, "include", "i", false, "Include response headers in output")

	return cmd
}

func isJSONResponse(resp *http.Response) bool {
	ct := resp.Header.Get("Content-Type")
	return strings.Contains(ct, "application/json")
}

func parseAPIInvocation(cmd *cobra.Command, args []string, method string, queryParams []string) (string, string, []string, error) {
	resolvedMethod := method
	var endpoint string
	var extraArgs []string

	switch {
	case len(args) >= 2 && isHTTPMethod(args[0]):
		// "api get /api/v1/..." — use first arg as method (unless -X was explicit)
		if !cmd.Flags().Changed("method") {
			resolvedMethod = strings.ToUpper(args[0])
		}
		endpoint = args[1]
		extraArgs = args[2:]
	case len(args) >= 2 && !looksLikeEndpoint(args[0]) && !strings.Contains(args[1], "="):
		return "", "", nil, fmt.Errorf("unknown HTTP method %q; expected get, post, put, patch, delete, head, or options", args[0])
	default:
		endpoint = args[0]
		extraArgs = args[1:]
	}

	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}

	resolvedQueryParams, err := appendQueryShorthand(queryParams, extraArgs)
	if err != nil {
		return "", "", nil, err
	}

	resolvedQueryParams, remapped := remapFinancialTransactionsQueryParams(endpoint, resolvedQueryParams)
	if remapped {
		errOut := cmd.ErrOrStderr()
		_, _ = fmt.Fprintln(errOut, "warning: remapped from_posted_at/to_posted_at to from_created_at/to_created_at for /api/v1/financial_transactions")
	}

	return resolvedMethod, endpoint, resolvedQueryParams, nil
}

func looksLikeEndpoint(arg string) bool {
	return strings.HasPrefix(arg, "/") || strings.Contains(arg, "/")
}

func appendQueryShorthand(queryParams []string, extraArgs []string) ([]string, error) {
	merged := make([]string, 0, len(queryParams)+len(extraArgs))
	merged = append(merged, queryParams...)

	for _, raw := range extraArgs {
		arg := strings.TrimSpace(raw)
		if arg == "" {
			continue
		}
		arg = strings.TrimPrefix(arg, "?")
		if arg == "" {
			continue
		}

		if strings.Contains(arg, "&") {
			parts := strings.Split(arg, "&")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				if !strings.Contains(part, "=") {
					return nil, fmt.Errorf("unexpected argument %q. If this is a query parameter, pass -q key=value", raw)
				}
				merged = append(merged, part)
			}
			continue
		}

		if !strings.Contains(arg, "=") {
			return nil, fmt.Errorf("unexpected argument %q. If this is a query parameter, pass -q key=value", raw)
		}
		merged = append(merged, arg)
	}

	return merged, nil
}

func remapFinancialTransactionsQueryParams(endpoint string, queryParams []string) ([]string, bool) {
	if !isFinancialTransactionsEndpoint(endpoint) {
		return queryParams, false
	}

	hasFromCreated := false
	hasToCreated := false
	for _, qp := range queryParams {
		key, _ := splitQueryParam(qp)
		switch key {
		case "from_created_at":
			hasFromCreated = true
		case "to_created_at":
			hasToCreated = true
		}
	}

	remapped := false
	out := make([]string, 0, len(queryParams))
	for _, qp := range queryParams {
		key, value := splitQueryParam(qp)
		switch key {
		case "from_posted_at":
			remapped = true
			if hasFromCreated {
				continue
			}
			out = append(out, "from_created_at="+value)
		case "to_posted_at":
			remapped = true
			if hasToCreated {
				continue
			}
			out = append(out, "to_created_at="+value)
		default:
			out = append(out, qp)
		}
	}

	return out, remapped
}

func splitQueryParam(qp string) (string, string) {
	parts := strings.SplitN(qp, "=", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func isFinancialTransactionsEndpoint(endpoint string) bool {
	normalized := strings.TrimSuffix(strings.ToLower(endpoint), "/")
	return normalized == "/api/v1/financial_transactions"
}

func isHTTPMethod(s string) bool {
	switch strings.ToUpper(s) {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	}
	return false
}
