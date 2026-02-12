# Airwallex CLI - Claude Instructions

## Interacting with Airwallex

The CLI is available as both `airwallex` and `awx`. Use `--jq 'expr'` for jq filtering (auto-enables JSON output).

### Command Shortcuts

All commands have short aliases. Use them for faster invocations:

| Long form | Short form |
|-----------|------------|
| `awx transfers list` | `awx tr ls` |
| `awx beneficiaries list` | `awx ben ls` |
| `awx issuing cards list` | `awx is cd ls` |
| `awx deposits list` | `awx dep ls` |
| `awx balances history` | `awx b hist` |
| `awx webhooks create` | `awx wh cr` |
| `awx fx conversions list` | `awx fx cv ls` |

Single-letter flag shortcuts: `-o` (output), `-j` (json), `-q` (query), `-s` (status), `-f` (from), `-n` (page-size), `-a` (all), `-i` (items-only)

Multi-letter flag aliases (hidden, for agents): `--jq` (query), `--bid` (beneficiary-id), `--tc` (transfer-currency), `--chid` (cardholder-id), `--cid` (card-id), `--fd`/`--td` (from-date/to-date), `--ev` (events), `--et` (entity-type)

Value shorthands: enum flags accept unambiguous prefixes, e.g. `--status pa` (PAID), `--status a` (ACTIVE), `--format c` (CSV)

### Card Transactions

```bash
# List recent transactions
awx issuing transactions list --page-size 20

# Filter by merchant name (case-insensitive)
awx issuing transactions list --jq \
  '[.[] | select(.merchant.name | test("COACH"; "i"))]'

# Last 10 transactions sorted by date
awx issuing transactions list --jq \
  'sort_by(.transaction_date) | reverse | .[0:10]'

# Top 5 highest spend transactions
awx issuing transactions list --output json --page-size 100 --query \
  'sort_by(.transaction_amount) | .[0:5]'

# Transactions over $500
awx issuing transactions list --jq \
  '[.[] | select(.transaction_amount < -500)]'

# Declined/failed transactions
awx issuing transactions list --jq \
  '[.[] | select(.status != "APPROVED")]'

# Spend by card (which cards are spending most)
awx issuing transactions list --output json --page-size 100 --query \
  'group_by(.card_nickname) | map({card: .[0].card_nickname, total: (map(.transaction_amount) | add)}) | sort_by(.total)'

# Top vendors by total spend
awx issuing transactions list --output json --page-size 100 --query \
  'group_by(.merchant.name) | map({vendor: .[0].merchant.name, total: (map(.transaction_amount) | add), count: length}) | sort_by(.total) | .[0:10]'

# Filter by date range
awx issuing transactions list --from 2025-01-01 --to 2025-01-31
```

### Transfers

```bash
# List recent transfers
awx transfers list --limit 20

# Filter by status
awx transfers list --status PAID

# Sort by amount (highest first)
awx transfers list --jq \
  'sort_by(.transfer_amount) | reverse | .[0:10]'

# Transfers over $1000
awx transfers list --jq \
  '[.[] | select(.transfer_amount > 1000)]'

# Failed/pending transfers
awx transfers list --jq \
  '[.[] | select(.status != "PAID")]'

# Total amount transferred
awx transfers list --jq \
  'map(.transfer_amount) | add'

# Total by currency
awx transfers list --jq \
  'group_by(.transfer_currency) | map({currency: .[0].transfer_currency, total: (map(.transfer_amount) | add)})'

# Filter by reference pattern
awx transfers list --jq \
  '[.[] | select(.reference | test("Invoice"; "i"))]'
```

### Common jq Patterns

| Query | Pattern |
|-------|---------|
| Filter by field | `[.[] \| select(.field \| test("value"; "i"))]` |
| Sort ascending | `sort_by(.field)` |
| Sort descending | `sort_by(.field) \| reverse` |
| Limit results | `.[0:N]` |
| Select fields | `.[] \| {field1, field2}` |
| Group and sum | `group_by(.field) \| map({key: .[0].field, sum: (map(.amount) \| add)})` |

### Transaction Fields

- `.transaction_id` - Unique transaction ID
- `.card_id` - Card used for transaction
- `.card_nickname` - Card nickname
- `.transaction_amount` - Amount (negative for debits)
- `.transaction_currency` - Currency code
- `.merchant.name` - Merchant name
- `.status` - Transaction status (APPROVED, DECLINED, etc.)
- `.transaction_date` - ISO timestamp

### Transfer Fields

- `.id` - Transfer ID
- `.transfer_amount` - Amount
- `.transfer_currency` - Currency
- `.status` - Status (PAID, PENDING, FAILED, etc.)
- `.reference` - Payment reference
- `.beneficiary_id` - Recipient ID
