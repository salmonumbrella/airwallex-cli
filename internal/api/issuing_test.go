package api

import (
	"testing"
)

func TestCardDetails_MaskedPAN(t *testing.T) {
	tests := []struct {
		name       string
		cardNumber string
		want       string
	}{
		{
			name:       "standard 16-digit card",
			cardNumber: "4532015112830366",
			want:       "************0366",
		},
		{
			name:       "amex 15-digit card",
			cardNumber: "378282246310005",
			want:       "***********0005",
		},
		{
			name:       "short card number",
			cardNumber: "123",
			want:       "****",
		},
		{
			name:       "exactly 4 digits",
			cardNumber: "1234",
			want:       "****",
		},
		{
			name:       "empty card number",
			cardNumber: "",
			want:       "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := &CardDetails{
				CardNumber: tt.cardNumber,
			}
			if got := cd.MaskedPAN(); got != tt.want {
				t.Errorf("CardDetails.MaskedPAN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCardDetails_Zeroize(t *testing.T) {
	cd := &CardDetails{
		CardID:      "card_123",
		CardNumber:  "4532015112830366",
		Cvv:         "123",
		ExpiryMonth: 12,
		ExpiryYear:  2025,
	}

	cd.Zeroize()

	if cd.CardNumber != "" {
		t.Errorf("CardNumber not zeroed: got %q", cd.CardNumber)
	}
	if cd.Cvv != "" {
		t.Errorf("Cvv not zeroed: got %q", cd.Cvv)
	}
	if cd.ExpiryMonth != 0 {
		t.Errorf("ExpiryMonth not zeroed: got %d", cd.ExpiryMonth)
	}
	if cd.ExpiryYear != 0 {
		t.Errorf("ExpiryYear not zeroed: got %d", cd.ExpiryYear)
	}
	// CardID should NOT be zeroed as it's not sensitive
	if cd.CardID != "card_123" {
		t.Errorf("CardID should not be zeroed: got %q", cd.CardID)
	}
}
