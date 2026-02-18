package cmd

import (
	"encoding/json"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
)

// Light structs for commands that produce >10 lines of JSON per item.
// Each keeps 3-7 essential fields so agent context stays small.

type lightTransfer struct {
	ID       string      `json:"id"`
	Amount   json.Number `json:"amount"`
	Currency string      `json:"currency"`
	Status   string      `json:"status"`
	BenID    string      `json:"beneficiary_id"`
	Created  string      `json:"created_at"`
}

func toLightTransfer(t api.Transfer) lightTransfer {
	return lightTransfer{
		ID:       t.TransferID,
		Amount:   t.TransferAmount,
		Currency: t.TransferCurrency,
		Status:   t.Status,
		BenID:    t.BeneficiaryID,
		Created:  t.CreatedAt,
	}
}

type lightBeneficiary struct {
	ID         string `json:"id"`
	Nickname   string `json:"nickname"`
	EntityType string `json:"entity_type"`
	Name       string `json:"name"`
	BankName   string `json:"bank_name"`
	Country    string `json:"country"`
}

func toLightBeneficiary(b api.Beneficiary) lightBeneficiary {
	name := b.Beneficiary.CompanyName
	if name == "" {
		name = b.Beneficiary.FirstName + " " + b.Beneficiary.LastName
	}
	return lightBeneficiary{
		ID:         b.BeneficiaryID,
		Nickname:   b.Nickname,
		EntityType: b.Beneficiary.EntityType,
		Name:       name,
		BankName:   b.Beneficiary.BankDetails.BankName,
		Country:    b.Beneficiary.BankDetails.BankCountryCode,
	}
}

type lightTransaction struct {
	ID       string      `json:"id"`
	CardID   string      `json:"card_id"`
	Type     string      `json:"type"`
	Amount   json.Number `json:"amount"`
	Currency string      `json:"currency"`
	Merchant string      `json:"merchant"`
	Status   string      `json:"status"`
}

func toLightTransaction(t api.Transaction) lightTransaction {
	return lightTransaction{
		ID:       t.TransactionID,
		CardID:   t.CardID,
		Type:     t.TransactionType,
		Amount:   t.Amount,
		Currency: t.Currency,
		Merchant: t.Merchant.Name,
		Status:   t.Status,
	}
}

type lightAuthorization struct {
	ID       string      `json:"id"`
	CardID   string      `json:"card_id"`
	Status   string      `json:"status"`
	Amount   json.Number `json:"amount"`
	Currency string      `json:"currency"`
	Merchant string      `json:"merchant"`
	Created  string      `json:"created_at"`
}

func toLightAuthorization(a api.Authorization) lightAuthorization {
	id := a.AuthorizationID
	if id == "" {
		id = a.ID
	}
	return lightAuthorization{
		ID:       id,
		CardID:   a.CardID,
		Status:   a.Status,
		Amount:   a.Amount,
		Currency: a.Currency,
		Merchant: a.Merchant.Name,
		Created:  a.CreatedAt,
	}
}

type lightConversion struct {
	ID           string      `json:"id"`
	SellCurrency string      `json:"sell_currency"`
	BuyCurrency  string      `json:"buy_currency"`
	SellAmount   json.Number `json:"sell_amount"`
	BuyAmount    json.Number `json:"buy_amount"`
	Status       string      `json:"status"`
	Created      string      `json:"created_at"`
}

func toLightConversion(c api.Conversion) lightConversion {
	return lightConversion{
		ID:           c.ID,
		SellCurrency: c.SellCurrency,
		BuyCurrency:  c.BuyCurrency,
		SellAmount:   c.SellAmount,
		BuyAmount:    c.BuyAmount,
		Status:       c.Status,
		Created:      c.CreatedAt,
	}
}

type lightInvoice struct {
	ID         string      `json:"id"`
	CustomerID string      `json:"customer_id"`
	Status     string      `json:"status"`
	Currency   string      `json:"currency"`
	Total      json.Number `json:"total_amount"`
	Created    string      `json:"created_at"`
}

func toLightInvoice(inv api.BillingInvoice) lightInvoice {
	return lightInvoice{
		ID:         inv.ID,
		CustomerID: inv.CustomerID,
		Status:     inv.Status,
		Currency:   inv.Currency,
		Total:      inv.TotalAmount,
		Created:    inv.CreatedAt,
	}
}

type lightSubscription struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Status     string `json:"status"`
	NextBill   string `json:"next_billing_at"`
	Created    string `json:"created_at"`
}

func toLightSubscription(s api.BillingSubscription) lightSubscription {
	return lightSubscription{
		ID:         s.ID,
		CustomerID: s.CustomerID,
		Status:     s.Status,
		NextBill:   s.NextBillingAt,
		Created:    s.CreatedAt,
	}
}

type lightPrice struct {
	ID        string      `json:"id"`
	ProductID string      `json:"product_id"`
	Currency  string      `json:"currency"`
	Amount    json.Number `json:"unit_amount"`
	Type      string      `json:"type"`
	Active    bool        `json:"active"`
}

func toLightPrice(p api.BillingPrice) lightPrice {
	return lightPrice{
		ID:        p.ID,
		ProductID: p.ProductID,
		Currency:  p.Currency,
		Amount:    p.UnitAmount,
		Type:      p.Type,
		Active:    p.Active,
	}
}

type lightReport struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	FileFormat string `json:"file_format"`
	FromDate   string `json:"from_date"`
	ToDate     string `json:"to_date"`
}

func toLightReport(r api.FinancialReport) lightReport {
	return lightReport{
		ID:         r.ID,
		Type:       r.Type,
		Status:     r.Status,
		FileFormat: r.FileFormat,
		FromDate:   r.FromDate,
		ToDate:     r.ToDate,
	}
}
