package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// findSubcmd walks a command path to find a nested subcommand.
func findSubcmd(root *cobra.Command, path []string) *cobra.Command {
	cmd := root
	for _, name := range path {
		found := false
		for _, sub := range cmd.Commands() {
			if sub.Name() == name {
				cmd = sub
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return cmd
}

func hasAlias(cmd *cobra.Command, alias string) bool {
	for _, a := range cmd.Aliases {
		if a == alias {
			return true
		}
	}
	return false
}

func TestParentCommandAliases(t *testing.T) {
	root := NewRootCmd()

	tests := []struct {
		command string
		aliases []string
	}{
		{"accounts", []string{"account", "acc", "ac"}},
		{"auth", []string{"au"}},
		{"balances", []string{"balance", "bal", "b"}},
		{"beneficiaries", []string{"benef", "beneficiary", "ben"}},
		{"billing", []string{"bill", "bi"}},
		{"deposits", []string{"deposit", "dep"}},
		{"issuing", []string{"iss", "is"}},
		{"cards", []string{"card", "cd"}},
		{"cardholders", []string{"cardholder", "ch"}},
		{"transactions", []string{"transaction", "tx"}},
		{"authorizations", []string{"authorization", "az"}},
		{"disputes", []string{"dispute", "di"}},
		{"payers", []string{"payer", "py"}},
		{"reports", []string{"report", "rp"}},
		{"schemas", []string{"schema", "sc"}},
		{"transfers", []string{"transfer", "tfr", "tr", "payout", "payouts"}},
		{"webhooks", []string{"webhook", "wh"}},
		{"api", []string{"ap"}},
		{"linked-accounts", []string{"la"}},
		{"payment-links", []string{"pl"}},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			cmd := findSubcmd(root, []string{tt.command})
			if cmd == nil {
				t.Fatalf("command %q not found", tt.command)
			}
			for _, expected := range tt.aliases {
				if !hasAlias(cmd, expected) {
					t.Errorf("command %q missing alias %q (has: %v)", tt.command, expected, cmd.Aliases)
				}
			}
		})
	}
}

func TestSubcommandAliases(t *testing.T) {
	root := NewRootCmd()

	tests := []struct {
		name    string
		path    []string
		aliases []string
	}{
		// transfers subcommands
		{"transfers list", []string{"transfers", "list"}, []string{"ls", "l"}},
		{"transfers get", []string{"transfers", "get"}, []string{"g"}},
		{"transfers create", []string{"transfers", "create"}, []string{"cr"}},
		{"transfers batch-create", []string{"transfers", "batch-create"}, []string{"bc"}},
		{"transfers cancel", []string{"transfers", "cancel"}, []string{"x"}},
		{"transfers confirmation", []string{"transfers", "confirmation"}, []string{"conf"}},

		// beneficiaries subcommands
		{"beneficiaries list", []string{"beneficiaries", "list"}, []string{"ls", "l"}},
		{"beneficiaries get", []string{"beneficiaries", "get"}, []string{"g"}},
		{"beneficiaries create", []string{"beneficiaries", "create"}, []string{"cr"}},
		{"beneficiaries update", []string{"beneficiaries", "update"}, []string{"up", "u"}},
		{"beneficiaries delete", []string{"beneficiaries", "delete"}, []string{"del", "rm"}},
		{"beneficiaries validate", []string{"beneficiaries", "validate"}, []string{"val", "v"}},

		// deposits subcommands
		{"deposits list", []string{"deposits", "list"}, []string{"ls", "l"}},
		{"deposits get", []string{"deposits", "get"}, []string{"g"}},

		// accounts subcommands
		{"accounts list", []string{"accounts", "list"}, []string{"ls", "l"}},
		{"accounts get", []string{"accounts", "get"}, []string{"g"}},

		// cards (top-level desire-path) subcommands
		{"cards list", []string{"cards", "list"}, []string{"ls", "l"}},
		{"cards get", []string{"cards", "get"}, []string{"g"}},
		{"cards create", []string{"cards", "create"}, []string{"cr"}},
		{"cards update", []string{"cards", "update"}, []string{"up", "u"}},
		{"cards activate", []string{"cards", "activate"}, []string{"act"}},
		{"cards details", []string{"cards", "details"}, []string{"det"}},
		{"cards limits", []string{"cards", "limits"}, []string{"lim"}},

		// cardholders subcommands
		{"cardholders list", []string{"cardholders", "list"}, []string{"ls", "l"}},
		{"cardholders get", []string{"cardholders", "get"}, []string{"g"}},
		{"cardholders create", []string{"cardholders", "create"}, []string{"cr"}},
		{"cardholders update", []string{"cardholders", "update"}, []string{"up", "u"}},

		// auth subcommands
		{"auth list", []string{"auth", "list"}, []string{"ls", "l"}},
		{"auth remove", []string{"auth", "remove"}, []string{"rm", "del"}},
		{"auth rename", []string{"auth", "rename"}, []string{"mv"}},
		{"auth login", []string{"auth", "login"}, []string{"li"}},
		{"auth add", []string{"auth", "add"}, []string{"a"}},
		{"auth test", []string{"auth", "test"}, []string{"t"}},

		// webhooks subcommands
		{"webhooks list", []string{"webhooks", "list"}, []string{"ls", "l"}},
		{"webhooks get", []string{"webhooks", "get"}, []string{"g"}},
		{"webhooks create", []string{"webhooks", "create"}, []string{"cr"}},
		{"webhooks delete", []string{"webhooks", "delete"}, []string{"del", "rm"}},

		// reports subcommands
		{"reports list", []string{"reports", "list"}, []string{"ls", "l"}},
		{"reports get", []string{"reports", "get"}, []string{"g"}},

		// balances subcommands
		{"balances history", []string{"balances", "history"}, []string{"hist", "h"}},

		// fx sub-groups (parent aliases)
		{"fx conversions", []string{"fx", "conversions"}, []string{"conversion", "conv", "cv"}},
		{"fx quotes", []string{"fx", "quotes"}, []string{"quote", "q"}},

		// fx conversions subcommands
		{"fx conversions list", []string{"fx", "conversions", "list"}, []string{"ls", "l"}},
		{"fx conversions get", []string{"fx", "conversions", "get"}, []string{"g"}},
		{"fx conversions create", []string{"fx", "conversions", "create"}, []string{"cr"}},

		// billing sub-groups
		{"billing customers", []string{"billing", "customers"}, []string{"cust", "cu"}},
		{"billing products", []string{"billing", "products"}, []string{"prod", "pr"}},
		{"billing prices", []string{"billing", "prices"}, []string{"price", "pc"}},
		{"billing invoices", []string{"billing", "invoices"}, []string{"inv"}},
		{"billing subscriptions", []string{"billing", "subscriptions"}, []string{"sub", "su"}},

		// billing customers subcommands
		{"billing customers list", []string{"billing", "customers", "list"}, []string{"ls", "l"}},
		{"billing customers get", []string{"billing", "customers", "get"}, []string{"g"}},
		{"billing customers create", []string{"billing", "customers", "create"}, []string{"cr"}},
		{"billing customers update", []string{"billing", "customers", "update"}, []string{"up", "u"}},

		// billing products subcommands
		{"billing products list", []string{"billing", "products", "list"}, []string{"ls", "l"}},
		{"billing products get", []string{"billing", "products", "get"}, []string{"g"}},
		{"billing products create", []string{"billing", "products", "create"}, []string{"cr"}},
		{"billing products update", []string{"billing", "products", "update"}, []string{"up", "u"}},

		// billing prices subcommands
		{"billing prices list", []string{"billing", "prices", "list"}, []string{"ls", "l"}},
		{"billing prices get", []string{"billing", "prices", "get"}, []string{"g"}},
		{"billing prices create", []string{"billing", "prices", "create"}, []string{"cr"}},
		{"billing prices update", []string{"billing", "prices", "update"}, []string{"up", "u"}},

		// billing invoices subcommands
		{"billing invoices list", []string{"billing", "invoices", "list"}, []string{"ls", "l"}},
		{"billing invoices get", []string{"billing", "invoices", "get"}, []string{"g"}},
		{"billing invoices create", []string{"billing", "invoices", "create"}, []string{"cr"}},

		// billing subscriptions subcommands
		{"billing subscriptions list", []string{"billing", "subscriptions", "list"}, []string{"ls", "l"}},
		{"billing subscriptions get", []string{"billing", "subscriptions", "get"}, []string{"g"}},
		{"billing subscriptions create", []string{"billing", "subscriptions", "create"}, []string{"cr"}},
		{"billing subscriptions update", []string{"billing", "subscriptions", "update"}, []string{"up", "u"}},
		{"billing subscriptions cancel", []string{"billing", "subscriptions", "cancel"}, []string{"x"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := findSubcmd(root, tt.path)
			if cmd == nil {
				t.Fatalf("command path %v not found", tt.path)
			}
			for _, expected := range tt.aliases {
				if !hasAlias(cmd, expected) {
					t.Errorf("command %v missing alias %q (has: %v)", tt.path, expected, cmd.Aliases)
				}
			}
		})
	}
}

func TestGlobalFlagShortcodes(t *testing.T) {
	root := NewRootCmd()

	tests := []struct {
		flag      string
		shorthand string
	}{
		{"output", "o"},
		{"json", "j"},
		{"debug", "d"},
		{"query", "q"},
		{"template", "t"},
		{"yes", "y"},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			f := root.PersistentFlags().Lookup(tt.flag)
			if f == nil {
				t.Fatalf("global flag --%s not found", tt.flag)
			}
			if f.Shorthand != tt.shorthand {
				t.Errorf("flag --%s shorthand = %q, want %q", tt.flag, f.Shorthand, tt.shorthand)
			}
		})
	}
}

func TestListFlagShortcodes(t *testing.T) {
	root := NewRootCmd()
	// Use transfers list as representative list command (page-based pagination)
	listCmd := findSubcmd(root, []string{"transfers", "list"})
	if listCmd == nil {
		t.Fatal("transfers list command not found")
	}

	tests := []struct {
		flag      string
		shorthand string
	}{
		{"page", "p"},
		{"page-size", "n"},
		{"all", "a"},
		{"items-only", "i"},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			f := listCmd.Flags().Lookup(tt.flag)
			if f == nil {
				t.Fatalf("list flag --%s not found", tt.flag)
			}
			if f.Shorthand != tt.shorthand {
				t.Errorf("flag --%s shorthand = %q, want %q", tt.flag, f.Shorthand, tt.shorthand)
			}
		})
	}
}

func TestPerCommandFlagShortcodes(t *testing.T) {
	root := NewRootCmd()

	tests := []struct {
		name      string
		path      []string
		flag      string
		shorthand string
	}{
		// transfers
		{"transfers list --status", []string{"transfers", "list"}, "status", "s"},
		{"transfers create --beneficiary-id", []string{"transfers", "create"}, "beneficiary-id", "b"},
		{"transfers create --method", []string{"transfers", "create"}, "method", "m"},
		{"transfers create --reference", []string{"transfers", "create"}, "reference", "r"},
		{"transfers create --wait", []string{"transfers", "create"}, "wait", "w"},

		// deposits
		{"deposits list --status", []string{"deposits", "list"}, "status", "s"},
		{"deposits list --from", []string{"deposits", "list"}, "from", "f"},

		// balances
		{"balances history --currency", []string{"balances", "history"}, "currency", "c"},
		{"balances history --from", []string{"balances", "history"}, "from", "f"},

		// cards
		{"cards list --status", []string{"cards", "list"}, "status", "s"},

		// webhooks
		{"webhooks create --url", []string{"webhooks", "create"}, "url", "u"},
		{"webhooks create --events", []string{"webhooks", "create"}, "events", "e"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := findSubcmd(root, tt.path)
			if cmd == nil {
				t.Fatalf("command path %v not found", tt.path)
			}
			f := cmd.Flags().Lookup(tt.flag)
			if f == nil {
				t.Fatalf("flag --%s not found on command %v", tt.flag, tt.path)
			}
			if f.Shorthand != tt.shorthand {
				t.Errorf("flag --%s shorthand = %q, want %q", tt.flag, f.Shorthand, tt.shorthand)
			}
		})
	}
}
