package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		Use:   "api <endpoint>",
		Short: "Make raw API requests",
		Long: `Make authenticated requests to any Airwallex API endpoint.

The endpoint should start with / (e.g., /api/v1/balances/current).
Authentication is handled automatically using your configured account.

Examples:
  # GET current balances
  airwallex api /api/v1/balances/current

  # GET with query parameters
  airwallex api /api/v1/transfers -q status=COMPLETED -q page_size=10

  # POST with inline JSON
  airwallex api /api/v1/beneficiaries -X POST -d '{"beneficiary": {...}}'

  # POST with file
  airwallex api /api/v1/transfers -X POST --data-file transfer.json

  # Include response headers
  airwallex api /api/v1/balances/current -i`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := args[0]
			if !strings.HasPrefix(endpoint, "/") {
				endpoint = "/" + endpoint
			}

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

			// Build URL with query params
			url := client.BaseURL() + endpoint
			if len(queryParams) > 0 {
				url += "?"
				for i, qp := range queryParams {
					if i > 0 {
						url += "&"
					}
					url += qp
				}
			}

			// Create request
			req, err := http.NewRequestWithContext(cmd.Context(), method, url, body)
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
				fmt.Fprintf(os.Stderr, "HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
				for k, v := range resp.Header {
					fmt.Fprintf(os.Stderr, "%s: %s\n", k, strings.Join(v, ", "))
				}
				fmt.Fprintln(os.Stderr)
			}

			// Output response body
			if outfmt.IsJSON(cmd.Context()) || isJSONResponse(resp) {
				// Pretty-print JSON
				var prettyJSON interface{}
				if err := json.Unmarshal(respBody, &prettyJSON); err == nil {
					if writeErr := outfmt.WriteJSONFiltered(os.Stdout, prettyJSON, outfmt.GetQuery(cmd.Context())); writeErr != nil {
						return writeErr
					}
				} else {
					// Not valid JSON, output raw
					fmt.Println(string(respBody))
				}
			} else {
				// Raw output
				fmt.Println(string(respBody))
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
