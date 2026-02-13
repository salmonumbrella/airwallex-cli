package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/batch"
	"github.com/salmonumbrella/airwallex-cli/internal/iocontext"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
)

// openSecretsStore is a variable that can be overridden in tests
var openSecretsStore = secrets.OpenDefault

// mustMarkRequired marks a flag as required, panicking on error.
// Use for flags that are definitely defined - panics indicate programmer error.
func mustMarkRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(fmt.Sprintf("flag %q not defined: %v", name, err))
	}
}

// newClientForCreds is a variable that can be overridden in tests.
var newClientForCreds = func(creds secrets.Credentials) (*api.Client, error) {
	if creds.AccountID != "" {
		return api.NewClientWithAccount(creds.ClientID, creds.APIKey, creds.AccountID)
	}
	return api.NewClient(creds.ClientID, creds.APIKey)
}

// getClient creates an API client from the current account
func getClient(ctx context.Context) (*api.Client, error) {
	account, err := requireAccount(ctx)
	if err != nil {
		return nil, err
	}

	store, err := openSecretsStore()
	if err != nil {
		return nil, err
	}

	creds, err := store.Get(account)
	if err != nil {
		return nil, fmt.Errorf("account not found: %s", account)
	}

	return newClientForCreds(creds)
}

// convertDateToRFC3339 converts a date string in YYYY-MM-DD format to RFC3339 format
// with time set to 00:00:00 UTC
func convertDateToRFC3339(dateStr string) (string, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", fmt.Errorf("expected format YYYY-MM-DD, got %q", dateStr)
	}
	// Convert to RFC3339 format with UTC timezone
	return t.UTC().Format(time.RFC3339), nil
}

// convertDateToRFC3339End converts a date string in YYYY-MM-DD format to RFC3339 format
// with time set to 23:59:59 UTC (inclusive end-of-day).
func convertDateToRFC3339End(dateStr string) (string, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", fmt.Errorf("expected format YYYY-MM-DD, got %q", dateStr)
	}
	endOfDay := t.UTC().Add(24*time.Hour - time.Second)
	return endOfDay.Format(time.RFC3339), nil
}

// validateDate validates that a date string is in YYYY-MM-DD format
func validateDate(dateStr string) error {
	if dateStr == "" {
		return nil // empty is valid (optional parameter)
	}
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format: expected YYYY-MM-DD, got %q", dateStr)
	}
	return nil
}

// validateAmount validates that an amount is positive
func validateAmount(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive, got %.2f", amount)
	}
	return nil
}

// validateCurrency validates that a currency code is 3 uppercase letters
func validateCurrency(currency string) error {
	if currency == "" {
		return nil // empty is valid (optional parameter)
	}
	if len(currency) != 3 {
		return fmt.Errorf("currency must be 3 letters, got %q", currency)
	}
	for _, r := range currency {
		if r < 'A' || r > 'Z' {
			return fmt.Errorf("currency must be uppercase letters, got %q", currency)
		}
	}
	return nil
}

// validateDateRange validates that from date is before to date when both are provided
func validateDateRange(from, to string) error {
	if from == "" || to == "" {
		return nil // only validate if both are provided
	}

	fromTime, err := time.Parse("2006-01-02", from)
	if err != nil {
		return fmt.Errorf("invalid from date: expected YYYY-MM-DD, got %q", from)
	}

	toTime, err := time.Parse("2006-01-02", to)
	if err != nil {
		return fmt.Errorf("invalid to date: expected YYYY-MM-DD, got %q", to)
	}

	if fromTime.After(toTime) {
		return fmt.Errorf("from date (%s) must be before or equal to to date (%s)", from, to)
	}

	return nil
}

// NormalizeIDArg accepts an ID or an ID embedded in a URL/path and returns just
// the canonical ID portion (the last path segment, without query/fragment).
//
// This is a "desire path" helper: agents often pass full URLs from webhooks or
// copied links instead of bare IDs.
func NormalizeIDArg(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	if i := strings.IndexAny(s, "?#"); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimRight(s, "/")
	if j := strings.LastIndex(s, "/"); j >= 0 && j < len(s)-1 {
		s = s[j+1:]
	}
	return s
}

// isTerminal is a variable that can be overridden in tests
var isTerminal = func() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// ConfirmOrYes prompts for confirmation unless --yes/--force flag is set.
// Returns true if confirmed, false if declined.
// Returns an error if stdin is not a TTY and confirmation is needed.
func ConfirmOrYes(ctx context.Context, prompt string) (bool, error) {
	// If --yes or --force flag is set, skip confirmation
	if outfmt.GetYes(ctx) {
		return true, nil
	}
	if outfmt.GetNoInput(ctx) {
		return false, fmt.Errorf("cannot prompt for confirmation: input disabled by --no-input (use --yes to skip)")
	}

	// Check if stdin is a terminal
	if !isTerminal() {
		return false, fmt.Errorf("cannot prompt for confirmation: stdin is not a terminal (use --yes to skip)")
	}

	// Get IO from context
	io := iocontext.GetIO(ctx)

	// Print prompt to stderr (so it doesn't interfere with stdout output)
	_, _ = fmt.Fprint(io.ErrOut, prompt+" [y/N]: ")

	// Read response from stdin
	reader := bufio.NewReader(io.In)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

func readJSONPayload(data, fromFile string) (map[string]interface{}, error) {
	if data != "" && fromFile != "" {
		return nil, fmt.Errorf("use only one of --data or --from-file")
	}

	var reader io.Reader
	switch {
	case fromFile != "":
		if fromFile == "-" {
			reader = os.Stdin
		} else {
			//nolint:gosec // G304: filename comes from user input, intentional
			f, err := os.Open(fromFile)
			if err != nil {
				return nil, fmt.Errorf("failed to open file: %w", err)
			}
			defer func() { _ = f.Close() }()
			reader = f
		}
	case data != "":
		reader = strings.NewReader(data)
	default:
		return nil, fmt.Errorf("provide --data or --from-file")
	}

	limitedReader := io.LimitReader(reader, batch.MaxInputSize+1)
	payload, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON payload: %w", err)
	}
	if len(payload) > batch.MaxInputSize {
		return nil, fmt.Errorf("input too large: exceeds maximum size of %d bytes", batch.MaxInputSize)
	}

	var result map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("invalid JSON object: %w", err)
	}
	return result, nil
}

func readOptionalJSONPayload(data, fromFile string) (map[string]interface{}, error) {
	if data == "" && fromFile == "" {
		return nil, nil
	}
	return readJSONPayload(data, fromFile)
}

func readTextInput(path string) (string, error) {
	var reader io.Reader
	switch path {
	case "":
		return "", fmt.Errorf("input file path is required")
	case "-":
		reader = os.Stdin
	default:
		//nolint:gosec // G304: filename comes from user input, intentional
		f, err := os.Open(path)
		if err != nil {
			return "", fmt.Errorf("failed to open file: %w", err)
		}
		defer func() { _ = f.Close() }()
		reader = f
	}

	limitedReader := io.LimitReader(reader, batch.MaxInputSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	if len(data) > batch.MaxInputSize {
		return "", fmt.Errorf("input too large: exceeds maximum size of %d bytes", batch.MaxInputSize)
	}

	return strings.TrimSpace(string(data)), nil
}

func readQueryInput(query, queryFile string) (string, error) {
	if query != "" && queryFile != "" {
		return "", fmt.Errorf("use only one of --query or --query-file")
	}
	if queryFile == "" {
		return query, nil
	}
	return readTextInput(queryFile)
}

// normalizePageSize clamps page size to the valid API range [1, 100].
func normalizePageSize(pageSize int) int {
	if pageSize < 1 {
		return 1
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

// normalizeEnumValue expands an abbreviated flag value to its canonical form.
// It does case-insensitive prefix matching against the known values.
// Returns the input unchanged if no match or ambiguous match.
func normalizeEnumValue(input string, validValues []string) string {
	if input == "" {
		return input
	}
	upper := strings.ToUpper(input)
	for _, v := range validValues {
		if strings.ToUpper(v) == upper {
			return v
		}
	}
	var match string
	count := 0
	for _, v := range validValues {
		if strings.HasPrefix(strings.ToUpper(v), upper) {
			match = v
			count++
		}
	}
	if count == 1 {
		return match
	}
	return input
}

var canonicalVerbAliases = map[string]string{
	"list":   "ls",
	"get":    "g",
	"show":   "g",
	"create": "mk",
	"update": "up",
	"edit":   "up",
	"delete": "rm",
	"remove": "rm",
	"search": "q",
	"query":  "q",
	"find":   "q",
}

func commandVerb(use string) string {
	parts := strings.Fields(strings.TrimSpace(use))
	if len(parts) == 0 {
		return ""
	}
	return strings.ToLower(parts[0])
}

func containsAlias(aliases []string, alias string) bool {
	for _, existing := range aliases {
		if existing == alias {
			return true
		}
	}
	return false
}

// addCanonicalVerbAliases appends canonical short aliases for common verb commands
// (list/get/create/update/delete/remove/search/query/find) when there is no sibling conflict.
func addCanonicalVerbAliases(cmd *cobra.Command) {
	if cmd == nil {
		return
	}

	if alias, ok := canonicalVerbAliases[commandVerb(cmd.Use)]; ok {
		addCommandAliasIfSafe(cmd, alias)
	}

	for _, sub := range cmd.Commands() {
		addCanonicalVerbAliases(sub)
	}
}

// addCommandAliasIfSafe appends alias to cmd when it does not already exist and
// does not conflict with sibling command names/aliases.
func addCommandAliasIfSafe(cmd *cobra.Command, alias string) bool {
	if cmd == nil || alias == "" {
		return false
	}
	if cmd.Name() == alias || containsAlias(cmd.Aliases, alias) {
		return false
	}

	parent := cmd.Parent()
	if parent != nil {
		for _, sibling := range parent.Commands() {
			if sibling == cmd {
				continue
			}
			if sibling.Name() == alias || containsAlias(sibling.Aliases, alias) {
				return false
			}
		}
	}

	cmd.Aliases = append(cmd.Aliases, alias)
	return true
}

type aliasFlagValue struct {
	target *pflag.Flag
}

type flagValueGetter interface {
	Get() interface{}
}

func (a *aliasFlagValue) String() string {
	if a == nil || a.target == nil || a.target.Value == nil {
		return ""
	}
	return a.target.Value.String()
}

func (a *aliasFlagValue) Set(value string) error {
	if a == nil || a.target == nil || a.target.Value == nil {
		return fmt.Errorf("alias target is not configured")
	}
	if err := a.target.Value.Set(value); err != nil {
		return err
	}
	// Ensure Cobra required-flag checks pass when alias is used.
	a.target.Changed = true
	return nil
}

func (a *aliasFlagValue) Type() string {
	if a == nil || a.target == nil || a.target.Value == nil {
		return ""
	}
	return a.target.Value.Type()
}

func (a *aliasFlagValue) Get() interface{} {
	if a == nil || a.target == nil || a.target.Value == nil {
		return nil
	}
	if getter, ok := a.target.Value.(flagValueGetter); ok {
		return getter.Get()
	}
	return a.target.Value.String()
}

// flagAlias registers a hidden alias for an existing flag.
// The alias delegates to the original flag's Value and marks it as Changed,
// so Cobra's required-flag checks pass when the alias is used.
func flagAlias(fs *pflag.FlagSet, name, alias string) {
	f := fs.Lookup(name)
	if f == nil {
		panic(fmt.Sprintf("flagAlias: flag %q not found", name))
	}

	// Append alias hint to original flag's usage text
	if strings.Contains(f.Usage, "[--") {
		// Already has an alias shown, append to the bracket
		f.Usage = strings.TrimSuffix(f.Usage, "]") + ", --" + alias + "]"
	} else {
		f.Usage += " [--" + alias + "]"
	}

	a := *f
	a.Name = alias
	a.Shorthand = ""
	a.Usage = ""
	a.Hidden = true
	a.Value = &aliasFlagValue{target: f}
	// Build a fresh annotations map so we don't inherit cobra's
	// "required" annotation from the original flag.
	a.Annotations = map[string][]string{
		"alias-of": {name},
	}
	fs.AddFlag(&a)
}

// flagOrAliasChanged returns true if the named flag or any of its hidden aliases was explicitly set.
func flagOrAliasChanged(cmd *cobra.Command, name string) bool {
	if cmd.Flags().Changed(name) {
		return true
	}
	if cmd.InheritedFlags().Changed(name) {
		return true
	}
	found := false
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if found {
			return
		}
		if ann, ok := f.Annotations["alias-of"]; ok && len(ann) > 0 && ann[0] == name {
			if cmd.Flags().Changed(f.Name) {
				found = true
			}
		}
	})
	return found
}
