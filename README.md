# Airwallex CLI

A CLI for managing Airwallex accounts, cards, transfers, and more from the command line.

## Features

- **Multiple account support** - manage multiple Airwallex accounts
- **Issuing** - create and manage cards, cardholders, and view transactions
- **Transfers & beneficiaries** - send payouts and manage recipients
- **Balances & accounts** - view account balances and global accounts
- **Secure credential storage** using OS keyring (Keychain on macOS, Secret Service on Linux, Credential Manager on Windows)
- **Auto-refreshing tokens** - authenticate once, use indefinitely

## Installation

### Homebrew

```bash
brew install salmonumbrella/tap/airwallex-cli
```

### From Source

```bash
go install github.com/salmonumbrella/airwallex-cli/cmd/airwallex@latest
```

## Quick Start

### 1. Authenticate

```bash
airwallex auth
```

This opens a browser for you to enter your Airwallex API credentials.

Or via command line:

```bash
airwallex auth add my-account --client-id <id> --api-key <key>
```

### 2. Test Authentication

```bash
airwallex auth test
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

### Credential Storage

Credentials are stored securely in your OS keyring:
- macOS: Keychain
- Linux: Secret Service (GNOME Keyring, KWallet)
- Windows: Credential Manager

## Commands

### Authentication

```bash
airwallex auth add <name> --client-id <id> [--api-key <key>]    # Add credentials
airwallex auth list                                               # List configured accounts
airwallex auth remove <name>                                      # Remove account
airwallex auth test [--account <name>]                            # Test credentials
```

### Balances & Accounts

```bash
airwallex balances                              # View current balances
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

### Transfers

```bash
airwallex transfers list [--status <status>]
airwallex transfers get <transferId>
airwallex transfers create --beneficiary-id <id> --amount <n> --currency <c> ...
airwallex transfers cancel <transferId>
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

Machine-readable output for scripting and automation:

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
  --amount 1000.00 \
  --currency USD \
  --reason "Vendor payment" \
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

## Global Flags

All commands support these flags:

- `--account <name>` - Account to use (overrides AWX_ACCOUNT)
- `--output <format>` - Output format: `text` or `json` (default: text)
- `--color <mode>` - Color mode: `auto`, `always`, or `never` (default: auto)
- `--help` - Show help for any command

## Development

After cloning, install git hooks:

```bash
make setup
```

This installs [lefthook](https://github.com/evilmartians/lefthook) pre-commit and pre-push hooks for linting and testing.

## License

MIT

## Links

- [Airwallex API Documentation](https://www.airwallex.com/docs/api)
- [GitHub Repository](https://github.com/salmonumbrella/airwallex-cli)
