# 🏦 Airwallex CLI — Banking in your terminal.

Airwallex in your terminal. Manage cards, send payouts, convert currencies, manage transfers, FX, beneficiaries, deposits, and webhooks 

## Features

- **Authentication** - authenticate once, tokens refresh indefinitely
- **Balances** - view account balances and global accounts
- **Benificiaries** - create and manage beneficiaries 
- **Cards** - create and manage cards, cardholders, and view transactions
- **Deposits** - view and manage deposit records
- **FX** - create foreign exchange quotes, convert currencies
- **Multiple accounts** - manage multiple Airwallex accounts
- **Payment links** - create and manage payment links
- **Reports** - generate and download financial reports, wire confirmations
- **Transactions** - view transaction history
- **Transfers** - send payouts and manage recipients
- **Webhooks** - configure and manage webhook endpoints

## Installation

### Homebrew

```bash
brew install salmonumbrella/tap/airwallex-cli
```

## Quick Start

### 1. Authenticate

Choose one of two methods:

**Browser:**
```bash
airwallex auth login
```

**Terminal:**
```bash
airwallex auth add my-account
# You'll be prompted securely for Client ID and API Key
```

### 2. Test Authentication

```bash
airwallex auth test --account my-account
```

## Configuration

### Account Selection

Specify the account using either a flag or environment variable:

```bash
# Via flag
airwallex balances --account my-account

# Via environment
export AWX_ACCOUNT=my-account
airwallex balances
```

### Environment Variables

- `AWX_ACCOUNT` - Default account name to use
- `AWX_OUTPUT` - Output format: `text` (default) or `json`
- `AWX_COLOR` - Color mode: `auto` (default), `always`, or `never`
- `NO_COLOR` - Set to any value to disable colors (standard convention)

## Security

### Credential Storage

Credentials are stored securely in your system's keychain:
- **macOS**: Keychain Access
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **Windows**: Credential Manager

## Rate Limiting

The Airwallex API enforces rate limits to ensure service stability. The CLI automatically handles rate limiting with:

- **Exponential backoff** - Retries with increasing delays (1s, 2s, 4s) plus jitter to avoid thundering herd
- **Retry-After header respect** - Honors the API's suggested retry timing when provided
- **Maximum retry attempts** - Up to 3 retries on 429 (Too Many Requests) responses
- **Circuit breaker** - After 5 consecutive server errors (5xx), requests are blocked for 30 seconds to prevent cascading failures

## Commands

### Authentication

```bash
airwallex auth login                     # Authenticate via browser (recommended)
airwallex auth add <name>                # Add credentials manually (prompts securely)
airwallex auth list                      # List configured accounts
airwallex auth remove <name>             # Remove account
airwallex auth test [--account <name>]   # Test credentials
```

### Balances & Accounts

```bash
airwallex balances                              # View current balances
airwallex balances history [--currency <c>] [--from <date>] [--to <date>]
airwallex accounts list                         # List global accounts
airwallex accounts get <accountId>              # Get account details
```

### Issuing - Cards

```bash
airwallex issuing cards list [--status <status>] [--cardholder-id <id>]
airwallex issuing cards get <cardId>
airwallex issuing cards create --cardholder-id <id> --form-factor VIRTUAL|PHYSICAL ...
airwallex issuing cards update <cardId> [--nickname <name>] [--status ACTIVE|INACTIVE|CLOSED]
airwallex issuing cards activate <cardId>
airwallex issuing cards details <cardId>        # Sensitive: full PAN, CVV, expiry
airwallex issuing cards limits <cardId>         # View spending limits and remaining balance
```

### Issuing - Cardholders

```bash
airwallex issuing cardholders list
airwallex issuing cardholders get <cardholderId>
airwallex issuing cardholders create --type INDIVIDUAL|DELEGATE --email <email> ...
airwallex issuing cardholders update <cardholderId> [--email <email>] ...
```

### Issuing - Transactions

```bash
airwallex issuing transactions list [--card-id <id>] [--from <date>] [--to <date>]
airwallex issuing transactions get <transactionId>
```

### Issuing - Authorizations

```bash
airwallex issuing authorizations list [--status <status>] [--card-id <id>] [--billing-currency <c>] \
  [--digital-wallet-token-id <id>] [--lifecycle-id <id>] [--retrieval-ref <ref>] \
  [--from <date>] [--to <date>]
airwallex issuing authorizations get <transactionId>
```

### Issuing - Disputes

```bash
airwallex issuing disputes list [--status <status>] [--detailed-status <status>] [--transaction-id <id>] \
  [--reason <reason>] [--reference <ref>] [--from <date>] [--to <date>] \
  [--from-updated <date>] [--to-updated <date>]
airwallex issuing disputes get <disputeId>
airwallex issuing disputes create --data '{...}'
airwallex issuing disputes update <disputeId> --data '{...}'
airwallex issuing disputes submit <disputeId>
airwallex issuing disputes cancel <disputeId>
```

### Transfers

```bash
airwallex transfers list [--status <status>]
airwallex transfers get <transferId>
airwallex transfers create --beneficiary-id <id> --transfer-amount <n> --transfer-currency <c> ...
airwallex transfers cancel <transferId>
airwallex transfers confirmation <transferId> --output <file.pdf>  # Download wire transfer confirmation letter
```

### Beneficiaries

```bash
airwallex beneficiaries list
airwallex beneficiaries get <beneficiaryId>
airwallex beneficiaries create --entity-type COMPANY|PERSONAL --bank-country <code> ...
airwallex beneficiaries update <beneficiaryId> ...
airwallex beneficiaries delete <beneficiaryId>
airwallex beneficiaries validate --entity-type ... --bank-country ...
```

#### Supported Countries & Routing

| Country | Payment Rail | Routing Flags |
|---------|--------------|---------------|
| 🇺🇸 USA | ACH / Fedwire | `--routing-number` |
| 🇬🇧 UK | Faster Payments | `--sort-code` |
| 🇪🇺 Europe | SEPA | `--iban`, `--swift-code` |
| 🇦🇺 Australia | BECS | `--bsb` |
| 🇦🇺 Australia | PayID/NPP | `--payid-phone`, `--payid-email`, `--payid-abn` |
| 🇨🇦 Canada | EFT | `--institution-number`, `--transit-number` |
| 🇨🇦 Canada | Interac e-Transfer | `--email`, `--phone`, `--clearing-system INTERAC` |
| 🇮🇳 India | NEFT/RTGS/IMPS | `--ifsc` |
| 🇲🇽 Mexico | SPEI | `--clabe` |
| 🇯🇵 Japan | Zengin | `--zengin-bank-code`, `--zengin-branch-code` |
| 🇨🇳 China | CNAPS | `--cnaps`, `--bank-name`, `--personal-id-type`, `--legal-rep-*` |
| 🇧🇷 Brazil | PIX / TED | `--cpf`, `--cnpj`, `--bank-branch` |
| 🇰🇷 South Korea | - | `--korea-bank-code` |
| 🇸🇬 Singapore | PayNow / FAST | `--nric`, `--uen`, `--sg-bank-code`, `--paynow-vpa` |
| 🇭🇰 Hong Kong | FPS | `--hk-bank-code`, `--fps-id`, `--hkid` |
| 🇸🇪 Sweden | Bankgiro | `--clearing-number` |
| 🌍 International | SWIFT | `--swift-code`, `--iban` |

Use `--validate` to check against schema without creating. See `airwallex beneficiaries create --help` for examples.

### Payers

```bash
airwallex payers list
airwallex payers get <payerId>
airwallex payers create --data '{...}'
airwallex payers update <payerId> --data '{...}'
airwallex payers delete <payerId>
airwallex payers validate --data '{...}'
```

### FX / Conversions

```bash
airwallex fx rates --sell USD --buy EUR              # Get current rates
airwallex fx quotes create --sell-currency USD --buy-currency EUR \
  --sell-amount 10000 --validity 1h                  # Lock a rate
airwallex fx quotes get <quoteId>                    # Get quote details
airwallex fx conversions list [--status <status>]   # List conversions
airwallex fx conversions get <conversionId>         # Get conversion details
airwallex fx conversions create --sell-currency USD --buy-currency EUR \
  --sell-amount 10000 [--quote-id <id>]             # Execute conversion
```

### Deposits

```bash
airwallex deposits list [--status SETTLED|PENDING|FAILED] [--from <date>] [--to <date>]
airwallex deposits get <depositId>
```

### Linked Accounts

```bash
airwallex linked-accounts list
airwallex linked-accounts get <accountId>
airwallex linked-accounts create --type AU_BANK --account-name "..." \
  --currency AUD --bsb "..." --account-number "..."
airwallex linked-accounts deposit <accountId> --amount 5000 --currency AUD
```

### API Field Discovery

```bash
# Discover required fields for a beneficiary
airwallex schemas beneficiary --bank-country US --entity-type COMPANY

# Discover required fields for a transfer
airwallex schemas transfer --source-currency USD --dest-currency EUR
```

### Raw API

```bash
# GET financial transactions with explicit -q flags
airwallex api get /api/v1/financial_transactions \
  -q from_created_at=2025-06-01T00:00:00+0000 \
  -q to_created_at=2025-06-30T23:59:59+0000 \
  -q page_size=100

# Query shorthand: extra key=value args are treated as query params
airwallex api get /api/v1/financial_transactions \
  from_created_at=2025-06-01T00:00:00+0000 \
  to_created_at=2025-06-30T23:59:59+0000 \
  page_size=100
```

For `/api/v1/financial_transactions`, use `from_created_at` and `to_created_at`.
If `from_posted_at`/`to_posted_at` are provided, the CLI remaps them to the created_at filters.

### Payment Links

```bash
airwallex payment-links list
airwallex payment-links get <linkId>
airwallex payment-links create --amount 100 --currency USD --description "Invoice #123"
```

### Billing

```bash
airwallex billing customers list
airwallex billing customers get <customerId>
airwallex billing customers create --data '{...}'
airwallex billing customers update <customerId> --data '{...}'

airwallex billing products list [--active true|false]
airwallex billing products get <productId>
airwallex billing products create --data '{...}'
airwallex billing products update <productId> --data '{...}'

airwallex billing prices list [--active true|false] [--currency <c>] [--product-id <id>] \
  [--recurring-period <n>] [--recurring-period-unit <unit>]
airwallex billing prices get <priceId>
airwallex billing prices create --data '{...}'
airwallex billing prices update <priceId> --data '{...}'

airwallex billing invoices list [--customer-id <id>] [--subscription-id <id>] [--status <status>] \
  [--from <date>] [--to <date>]
airwallex billing invoices get <invoiceId>
airwallex billing invoices create --data '{...}'
airwallex billing invoices preview --data '{...}'
airwallex billing invoices items list <invoiceId>
airwallex billing invoices items get <invoiceId> <itemId>

airwallex billing subscriptions list [--customer-id <id>] [--status <status>] \
  [--recurring-period <n>] [--recurring-period-unit <unit>] [--from <date>] [--to <date>]
airwallex billing subscriptions get <subscriptionId>
airwallex billing subscriptions create --data '{...}'
airwallex billing subscriptions update <subscriptionId> --data '{...}'
airwallex billing subscriptions cancel <subscriptionId> [--data '{...}']
airwallex billing subscriptions items list <subscriptionId>
airwallex billing subscriptions items get <subscriptionId> <itemId>
```

### Webhooks

```bash
airwallex webhooks list
airwallex webhooks get <webhookId>
airwallex webhooks create --url https://example.com/hook \
  --events transfer.completed,deposit.settled
airwallex webhooks delete <webhookId>
```

### Reports

```bash
airwallex reports list [--page-size <n>]
airwallex reports get <reportId>
airwallex reports account-statement --from-date <YYYY-MM-DD> --to-date <YYYY-MM-DD> \
  --currencies <CAD,USD> [--output <file>] [--wait]
airwallex reports balance-activity --from-date <YYYY-MM-DD> --to-date <YYYY-MM-DD> \
  --format CSV|EXCEL|PDF [--output <file>] [--wait]
airwallex reports transaction-recon --from-date <YYYY-MM-DD> --to-date <YYYY-MM-DD> \
  --format CSV|EXCEL [--output <file>] [--wait]
airwallex reports settlement --from-date <YYYY-MM-DD> --to-date <YYYY-MM-DD> \
  --format CSV|EXCEL [--output <file>] [--wait]
```

## Output Formats

### Text

Human-readable tables with colors and formatting:

```bash
$ airwallex balances
CURRENCY    AVAILABLE       PENDING     RESERVED
USD         12,450.00       500.00      0.00
CAD         8,200.50        0.00        150.00

$ airwallex issuing cards list
CARD_ID                              STATUS    NICKNAME        LAST4    CARDHOLDER
card_abc123...                       ACTIVE    Marketing       4242     John Doe
card_def456...                       INACTIVE  Travel          1234     Jane Smith
```

### JSON

Machine-readable output:

```bash
$ airwallex balances --output json
{
  "balances": [
    {"currency": "USD", "available": 12450.00, "pending": 500.00},
    {"currency": "CAD", "available": 8200.50, "pending": 0.00}
  ]
}
```

Data goes to stdout, errors and progress to stderr for clean piping.

## Examples

### Create a virtual card for a cardholder

```bash
# First, create a cardholder
airwallex issuing cardholders create \
  --type INDIVIDUAL \
  --email john@example.com \
  --first-name John \
  --last-name Doe

# Then create a virtual card
airwallex issuing cards create \
  --cardholder-id <cardholderId> \
  --form-factor VIRTUAL \
  --nickname "Marketing Ads"
```

### Send a transfer to a beneficiary

```bash
# List beneficiaries to find ID
airwallex beneficiaries list

# Create transfer
airwallex transfers create \
  --beneficiary-id <beneficiaryId> \
  --transfer-amount 1000.00 \
  --transfer-currency USD \
  --source-currency USD \
  --reason "payment_to_supplier" \
  --reference "INV-2024-001"
```

### View recent transactions for a card

```bash
airwallex issuing transactions list \
  --card-id <cardId> \
  --from 2024-01-01 \
  --to 2024-01-31 \
  --output json | jq '.transactions[] | select(.amount > 100)'
```

### Switch between accounts

```bash
# Check production account
airwallex balances --account prod

# Check sandbox account
airwallex balances --account sandbox

# Or set default
export AWX_ACCOUNT=prod
airwallex balances
```

### Automation

Use `--yes` to skip confirmations, `--output-limit` to cap output rows, `--sort-by` for ordering, and `--all` to auto-paginate through every page (pagination uses `--page`/`--page-size`):

```bash
# Delete a beneficiary without confirmation prompt
airwallex beneficiaries delete ben_xxx --yes

# Get the 5 most recent transfers
airwallex transfers list --page-size 5 --sort-by created_at --desc --output json

# Fetch the first 100 cards
airwallex issuing cards list --page-size 100 --output json

# Fetch ALL beneficiaries (auto-paginates through every page)
airwallex beneficiaries list --all --output json

# Pipeline: cancel all pending transfers older than 30 days
airwallex transfers list --status PENDING --output json \
  | jq -r '.items[] | select(.created_at < "2024-01-01") | .id' \
  | xargs -I{} airwallex transfers cancel {} --yes

# Agent-friendly: get latest 10 transactions sorted by amount
airwallex issuing transactions list --page-size 10 --sort-by amount --desc --output json

# Desire path: fetch any resource by ID (auto-detect type)
airwallex get tfr_123
airwallex get ben_456
airwallex get inv_123:item_456

# Desire path: verb-first routers
airwallex list transfers --page-size 5
airwallex create webhook --url https://example.com/hook --events transfer.completed
airwallex cancel tfr_123

# Agent mode: stable JSON, no color, no prompts, structured errors
AWX_AGENT=1 airwallex list transfers --page-size 5
```

### Debug Mode

Enable verbose output for troubleshooting:

```bash
airwallex --debug transfers list
# Shows: api request method=GET url=/api/v1/transfers
# Shows: api response status=200 content_length=1234
```

### Dry-Run Mode

Preview mutations before executing:

```bash
airwallex transfers create --dry-run \
  --beneficiary-id ben_xxx \
  --transfer-amount 1000 \
  --transfer-currency USD \
  --source-currency USD \
  --reference "Test" \
  --reason "payment_to_supplier"

# Output:
# [DRY-RUN] Would create transfer
# ─────────────────────────────────────
# Send 1000.00 USD to John Smith
#   Beneficiary: John Smith
#   Transfer Amount: 1000.00 USD
# ─────────────────────────────────────
# No changes made (dry-run mode)
```

### Wait for Completion

Wait for async operations to complete:

```bash
# Wait for transfer to reach final status
airwallex transfers create --wait --timeout 300 \
  --beneficiary-id ben_xxx ...
```

### Batch Operations

Create multiple resources from a file:

```bash
# From JSON file
airwallex transfers batch-create --from-file payroll.json

# From stdin
cat transfers.json | airwallex transfers batch-create

# Continue processing on errors
airwallex transfers batch-create --from-file data.json --continue-on-error
```

### JQ Filtering

Filter JSON output with JQ expressions:

List commands return `{items, has_more, next_cursor, _links}` in JSON mode, so use `.items[]` when filtering.

```bash
# Get only USD balance
airwallex balances --output json --query '.balances[] | select(.currency=="USD")'

# Extract transfer IDs
airwallex transfers list --output json --query '[.items[].id]'

# Filter by status
airwallex transfers list --output json --query '.items[] | select(.status=="PENDING")'

# Output array only (no pagination metadata)
airwallex transfers list --output json --items-only | jq '.[] | select(.status=="PENDING")'

# Load a longer query from a file
airwallex transfers list --output json --query-file ./query.jq

# Filter beneficiaries by nickname (case-insensitive)
airwallex beneficiaries list --output json --query \
  '.items[] | select((.nickname // "") | test("Jason|Jing Sen|Huang"; "i")) | {id: .id, nickname: .nickname, account_name: .beneficiary.bank_details.account_name}'
```

## Global Flags

All commands support these flags:

- `--account <name>` - Account to use (overrides AWX_ACCOUNT)
- `--output <format>` - Output format: `text` or `json` (default: text)
- `--json` - Shorthand for `--output json`
- `--color <mode>` - Color mode: `auto`, `always`, or `never` (default: auto)
- `--no-color` - Shorthand for `--color never`
- `--agent` - Agent mode: stable JSON, no color, no prompts, structured errors (or `AWX_AGENT` env)
- `--debug` - Enable debug output (shows API requests/responses)
- `--query <expr>` - JQ filter expression for JSON output
- `--query-file <path>` - Read JQ filter expression from file (use `-` for stdin)
- `--template <tmpl>` - Go template for custom output (e.g., `{{.id}}: {{.status}}`)
- `--items-only` - Output items array only for list commands (JSON mode)
- `--results-only` - Alias for `--items-only`
- `--yes`, `-y` - Skip confirmation prompts (useful for scripts and automation)
- `--force` - Alias for `--yes`
- `--output-limit <n>` - Limit number of results in output (0 = no limit)
- `--sort-by <field>` - Sort results by field name (e.g., `created_at`, `amount`)
- `--desc` - Sort descending (requires `--sort-by`)
- `--help` - Show help for any command
- `--version` - Show version information (via `airwallex version`)

## Shell Completions

Generate shell completions for your preferred shell:

### Bash

```bash
# macOS (Homebrew):
airwallex completion bash > $(brew --prefix)/etc/bash_completion.d/airwallex

# Linux:
airwallex completion bash > /etc/bash_completion.d/airwallex

# Or source directly in current session:
source <(airwallex completion bash)
```

### Zsh

```zsh
# Save to fpath:
airwallex completion zsh > "${fpath[1]}/_airwallex"

# Or add to .zshrc for auto-loading:
echo 'autoload -U compinit; compinit' >> ~/.zshrc
echo 'source <(airwallex completion zsh)' >> ~/.zshrc
```

### Fish

```fish
airwallex completion fish > ~/.config/fish/completions/airwallex.fish
```

### PowerShell

```powershell
# Load for current session:
airwallex completion powershell | Out-String | Invoke-Expression

# Or add to profile for persistence:
airwallex completion powershell >> $PROFILE
```

## Development

After cloning, install git hooks:

```bash
make setup
```

This installs [lefthook](https://github.com/evilmartians/lefthook) pre-commit and pre-push hooks for linting and testing.

### Future Infrastructure

The following packages are ready for integration but not yet used:

- **`internal/schemacache`** - Local caching for beneficiary schemas with TTL. Reduces API calls by caching schemas after fetching.
- **`internal/reqbuilder`** - Converts flat CLI flag paths (e.g., `beneficiary.bank_details.account_name`) into nested JSON structures for API requests.

## License

MIT

## Links

- [Airwallex API Documentation](https://www.airwallex.com/docs/api)
