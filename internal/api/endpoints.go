package api

import "net/http"

// Endpoint defines metadata for an API endpoint
type Endpoint struct {
	Path           string
	Method         string
	RequiresIdem   bool // Needs idempotency key
	ExpectedStatus int  // Expected success status (200, 201, etc.)
}

// NeedsIdempotencyKey returns whether this endpoint requires an idempotency key
func (e Endpoint) NeedsIdempotencyKey() bool {
	return e.RequiresIdem
}

// Registry of all API endpoints
var Endpoints = struct {
	// Authentication
	Login Endpoint

	// Transfers
	TransfersList   Endpoint
	TransfersGet    Endpoint
	TransfersCreate Endpoint
	TransfersCancel Endpoint

	// Beneficiaries
	BeneficiariesList     Endpoint
	BeneficiariesGet      Endpoint
	BeneficiariesCreate   Endpoint
	BeneficiariesUpdate   Endpoint
	BeneficiariesDelete   Endpoint
	BeneficiariesValidate Endpoint

	// Payers
	PayersList     Endpoint
	PayersGet      Endpoint
	PayersCreate   Endpoint
	PayersUpdate   Endpoint
	PayersDelete   Endpoint
	PayersValidate Endpoint

	// Confirmation Letters
	ConfirmationLettersCreate Endpoint

	// Balances
	BalancesCurrent Endpoint
	BalancesHistory Endpoint

	// Cards
	CardsList       Endpoint
	CardsGet        Endpoint
	CardsGetDetails Endpoint
	CardsGetLimits  Endpoint
	CardsCreate     Endpoint
	CardsUpdate     Endpoint
	CardsActivate   Endpoint

	// Cardholders
	CardholdersList   Endpoint
	CardholdersGet    Endpoint
	CardholdersCreate Endpoint
	CardholdersUpdate Endpoint

	// Card Transactions
	CardTransactionsList Endpoint
	CardTransactionsGet  Endpoint

	// Issuing Authorizations
	AuthorizationsList Endpoint
	AuthorizationsGet  Endpoint

	// Issuing Transaction Disputes
	TransactionDisputesList   Endpoint
	TransactionDisputesGet    Endpoint
	TransactionDisputesCreate Endpoint
	TransactionDisputesUpdate Endpoint
	TransactionDisputesSubmit Endpoint
	TransactionDisputesCancel Endpoint

	// FX
	FXRatesCurrent      Endpoint
	FXQuotesCreate      Endpoint
	FXQuotesGet         Endpoint
	FXConversionsList   Endpoint
	FXConversionsGet    Endpoint
	FXConversionsCreate Endpoint

	// Deposits
	DepositsList Endpoint
	DepositsGet  Endpoint

	// Reports
	ReportsCreate     Endpoint
	ReportsList       Endpoint
	ReportsGet        Endpoint
	ReportsGetContent Endpoint

	// Webhooks
	WebhooksList   Endpoint
	WebhooksGet    Endpoint
	WebhooksCreate Endpoint
	WebhooksDelete Endpoint

	// Schemas
	BeneficiarySchemaGenerate Endpoint
	TransferSchemaGenerate    Endpoint

	// Global Accounts
	AccountsList Endpoint
	AccountsGet  Endpoint

	// Linked Accounts
	LinkedAccountsList            Endpoint
	LinkedAccountsGet             Endpoint
	LinkedAccountsCreate          Endpoint
	LinkedAccountsInitiateDeposit Endpoint

	// Payment Links
	PaymentLinksList   Endpoint
	PaymentLinksGet    Endpoint
	PaymentLinksCreate Endpoint

	// Billing Customers
	BillingCustomersList   Endpoint
	BillingCustomersGet    Endpoint
	BillingCustomersCreate Endpoint
	BillingCustomersUpdate Endpoint

	// Billing Products
	BillingProductsList   Endpoint
	BillingProductsGet    Endpoint
	BillingProductsCreate Endpoint
	BillingProductsUpdate Endpoint

	// Billing Prices
	BillingPricesList   Endpoint
	BillingPricesGet    Endpoint
	BillingPricesCreate Endpoint
	BillingPricesUpdate Endpoint

	// Billing Invoices
	BillingInvoicesList   Endpoint
	BillingInvoicesGet    Endpoint
	BillingInvoicesCreate Endpoint

	// Billing Subscriptions
	BillingSubscriptionsList   Endpoint
	BillingSubscriptionsGet    Endpoint
	BillingSubscriptionsCreate Endpoint
}{
	// Authentication
	Login: Endpoint{
		Path:           "/api/v1/authentication/login",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusCreated,
	},

	// Transfers
	TransfersList: Endpoint{
		Path:           "/api/v1/transfers",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	TransfersGet: Endpoint{
		Path:           "/api/v1/transfers/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	TransfersCreate: Endpoint{
		Path:           "/api/v1/transfers/create",
		Method:         http.MethodPost,
		RequiresIdem:   true,
		ExpectedStatus: http.StatusCreated,
	},
	TransfersCancel: Endpoint{
		Path:           "/api/v1/transfers/{id}/cancel",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Beneficiaries
	BeneficiariesList: Endpoint{
		Path:           "/api/v1/beneficiaries",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BeneficiariesGet: Endpoint{
		Path:           "/api/v1/beneficiaries/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BeneficiariesCreate: Endpoint{
		Path:           "/api/v1/beneficiaries/create",
		Method:         http.MethodPost,
		RequiresIdem:   true,
		ExpectedStatus: http.StatusCreated,
	},
	BeneficiariesUpdate: Endpoint{
		Path:           "/api/v1/beneficiaries/{id}/update",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BeneficiariesDelete: Endpoint{
		Path:           "/api/v1/beneficiaries/{id}/delete",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BeneficiariesValidate: Endpoint{
		Path:           "/api/v1/beneficiaries/validate",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Payers
	PayersList: Endpoint{
		Path:           "/api/v1/payers",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	PayersGet: Endpoint{
		Path:           "/api/v1/payers/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	PayersCreate: Endpoint{
		Path:           "/api/v1/payers/create",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusCreated,
	},
	PayersUpdate: Endpoint{
		Path:           "/api/v1/payers/update/{id}",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	PayersDelete: Endpoint{
		Path:           "/api/v1/payers/delete/{id}",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	PayersValidate: Endpoint{
		Path:           "/api/v1/payers/validate",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Confirmation Letters
	ConfirmationLettersCreate: Endpoint{
		Path:           "/api/v1/confirmation_letters/create",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusCreated,
	},

	// Balances
	BalancesCurrent: Endpoint{
		Path:           "/api/v1/balances/current",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BalancesHistory: Endpoint{
		Path:           "/api/v1/balances/history",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Cards
	CardsList: Endpoint{
		Path:           "/api/v1/issuing/cards",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	CardsGet: Endpoint{
		Path:           "/api/v1/issuing/cards/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	CardsGetDetails: Endpoint{
		Path:           "/api/v1/issuing/cards/{id}/details",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	CardsGetLimits: Endpoint{
		Path:           "/api/v1/issuing/cards/{id}/limits",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	CardsCreate: Endpoint{
		Path:           "/api/v1/issuing/cards/create",
		Method:         http.MethodPost,
		RequiresIdem:   true,
		ExpectedStatus: http.StatusCreated,
	},
	CardsUpdate: Endpoint{
		Path:           "/api/v1/issuing/cards/{id}/update",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	CardsActivate: Endpoint{
		Path:           "/api/v1/issuing/cards/{id}/activate",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Cardholders
	CardholdersList: Endpoint{
		Path:           "/api/v1/issuing/cardholders",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	CardholdersGet: Endpoint{
		Path:           "/api/v1/issuing/cardholders/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	CardholdersCreate: Endpoint{
		Path:           "/api/v1/issuing/cardholders/create",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusCreated,
	},
	CardholdersUpdate: Endpoint{
		Path:           "/api/v1/issuing/cardholders/{id}/update",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Card Transactions
	CardTransactionsList: Endpoint{
		Path:           "/api/v1/issuing/transactions",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	CardTransactionsGet: Endpoint{
		Path:           "/api/v1/issuing/transactions/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Issuing Authorizations
	AuthorizationsList: Endpoint{
		Path:           "/api/v1/issuing/authorizations",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	AuthorizationsGet: Endpoint{
		Path:           "/api/v1/issuing/authorizations/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Issuing Transaction Disputes
	TransactionDisputesList: Endpoint{
		Path:           "/api/v1/issuing/transaction_disputes",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	TransactionDisputesGet: Endpoint{
		Path:           "/api/v1/issuing/transaction_disputes/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	TransactionDisputesCreate: Endpoint{
		Path:           "/api/v1/issuing/transaction_disputes/create",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusCreated,
	},
	TransactionDisputesUpdate: Endpoint{
		Path:           "/api/v1/issuing/transaction_disputes/{id}/update",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	TransactionDisputesSubmit: Endpoint{
		Path:           "/api/v1/issuing/transaction_disputes/{id}/submit",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	TransactionDisputesCancel: Endpoint{
		Path:           "/api/v1/issuing/transaction_disputes/{id}/cancel",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// FX
	FXRatesCurrent: Endpoint{
		Path:           "/api/v1/fx/rates/current",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	FXQuotesCreate: Endpoint{
		Path:           "/api/v1/fx/quotes/create",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusCreated,
	},
	FXQuotesGet: Endpoint{
		Path:           "/api/v1/fx/quotes/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	FXConversionsList: Endpoint{
		Path:           "/api/v1/fx/conversions",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	FXConversionsGet: Endpoint{
		Path:           "/api/v1/fx/conversions/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	FXConversionsCreate: Endpoint{
		Path:           "/api/v1/fx/conversions/create",
		Method:         http.MethodPost,
		RequiresIdem:   true,
		ExpectedStatus: http.StatusCreated,
	},

	// Deposits
	DepositsList: Endpoint{
		Path:           "/api/v1/deposits",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	DepositsGet: Endpoint{
		Path:           "/api/v1/deposits/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Reports
	ReportsCreate: Endpoint{
		Path:           "/api/v1/finance/financial_reports/create",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusCreated,
	},
	ReportsList: Endpoint{
		Path:           "/api/v1/finance/financial_reports",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	ReportsGet: Endpoint{
		Path:           "/api/v1/finance/financial_reports/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	ReportsGetContent: Endpoint{
		Path:           "/api/v1/finance/financial_reports/{id}/content",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Webhooks
	WebhooksList: Endpoint{
		Path:           "/api/v1/webhooks",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	WebhooksGet: Endpoint{
		Path:           "/api/v1/webhooks/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	WebhooksCreate: Endpoint{
		Path:           "/api/v1/webhooks/create",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusCreated,
	},
	WebhooksDelete: Endpoint{
		Path:           "/api/v1/webhooks/{id}/delete",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Schemas
	BeneficiarySchemaGenerate: Endpoint{
		Path:           "/api/v1/beneficiary_api_schemas/generate",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	TransferSchemaGenerate: Endpoint{
		Path:           "/api/v1/transfer_api_schemas/generate",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Global Accounts
	AccountsList: Endpoint{
		Path:           "/api/v1/global_accounts",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	AccountsGet: Endpoint{
		Path:           "/api/v1/global_accounts/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Linked Accounts
	LinkedAccountsList: Endpoint{
		Path:           "/api/v1/linked_accounts",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	LinkedAccountsGet: Endpoint{
		Path:           "/api/v1/linked_accounts/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	LinkedAccountsCreate: Endpoint{
		Path:           "/api/v1/linked_accounts/create",
		Method:         http.MethodPost,
		RequiresIdem:   true,
		ExpectedStatus: http.StatusCreated,
	},
	LinkedAccountsInitiateDeposit: Endpoint{
		Path:           "/api/v1/linked_accounts/{id}/initiate_deposit",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Payment Links
	PaymentLinksList: Endpoint{
		Path:           "/api/v1/pa/payment_links",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	PaymentLinksGet: Endpoint{
		Path:           "/api/v1/pa/payment_links/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	PaymentLinksCreate: Endpoint{
		Path:           "/api/v1/pa/payment_links/create",
		Method:         http.MethodPost,
		RequiresIdem:   true,
		ExpectedStatus: http.StatusCreated,
	},

	// Billing Customers
	BillingCustomersList: Endpoint{
		Path:           "/api/v1/billing_customers",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BillingCustomersGet: Endpoint{
		Path:           "/api/v1/billing_customers/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BillingCustomersCreate: Endpoint{
		Path:           "/api/v1/billing_customers/create",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusCreated,
	},
	BillingCustomersUpdate: Endpoint{
		Path:           "/api/v1/billing_customers/{id}/update",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Billing Products
	BillingProductsList: Endpoint{
		Path:           "/api/v1/products",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BillingProductsGet: Endpoint{
		Path:           "/api/v1/products/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BillingProductsCreate: Endpoint{
		Path:           "/api/v1/products/create",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusCreated,
	},
	BillingProductsUpdate: Endpoint{
		Path:           "/api/v1/products/{id}/update",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Billing Prices
	BillingPricesList: Endpoint{
		Path:           "/api/v1/prices",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BillingPricesGet: Endpoint{
		Path:           "/api/v1/prices/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BillingPricesCreate: Endpoint{
		Path:           "/api/v1/prices/create",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusCreated,
	},
	BillingPricesUpdate: Endpoint{
		Path:           "/api/v1/prices/{id}/update",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},

	// Billing Invoices
	BillingInvoicesList: Endpoint{
		Path:           "/api/v1/invoices",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BillingInvoicesGet: Endpoint{
		Path:           "/api/v1/invoices/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BillingInvoicesCreate: Endpoint{
		Path:           "/api/v1/invoices/create",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusCreated,
	},

	// Billing Subscriptions
	BillingSubscriptionsList: Endpoint{
		Path:           "/api/v1/subscriptions",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BillingSubscriptionsGet: Endpoint{
		Path:           "/api/v1/subscriptions/{id}",
		Method:         http.MethodGet,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusOK,
	},
	BillingSubscriptionsCreate: Endpoint{
		Path:           "/api/v1/subscriptions/create",
		Method:         http.MethodPost,
		RequiresIdem:   false,
		ExpectedStatus: http.StatusCreated,
	},
}
