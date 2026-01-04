package api

import (
	"encoding/json"
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

func TestCardholder_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		json          string
		wantFirstName string
		wantLastName  string
		wantEmail     string
	}{
		{
			name: "nested individual.name structure",
			json: `{
				"cardholder_id": "ch_123",
				"type": "INDIVIDUAL",
				"email": "john@example.com",
				"mobile_number": "+1234567890",
				"individual": {
					"name": {
						"first_name": "John",
						"last_name": "Doe"
					}
				},
				"status": "ACTIVE",
				"created_at": "2024-01-01T00:00:00Z"
			}`,
			wantFirstName: "John",
			wantLastName:  "Doe",
			wantEmail:     "john@example.com",
		},
		{
			name: "flat first_name/last_name structure",
			json: `{
				"cardholder_id": "ch_456",
				"type": "INDIVIDUAL",
				"email": "jane@example.com",
				"mobile_number": "+0987654321",
				"first_name": "Jane",
				"last_name": "Smith",
				"status": "ACTIVE",
				"created_at": "2024-01-02T00:00:00Z"
			}`,
			wantFirstName: "Jane",
			wantLastName:  "Smith",
			wantEmail:     "jane@example.com",
		},
		{
			name: "nested takes precedence over flat",
			json: `{
				"cardholder_id": "ch_789",
				"type": "INDIVIDUAL",
				"email": "bob@example.com",
				"individual": {
					"name": {
						"first_name": "Bob",
						"last_name": "Nested"
					}
				},
				"first_name": "Robert",
				"last_name": "Flat",
				"status": "ACTIVE",
				"created_at": "2024-01-03T00:00:00Z"
			}`,
			wantFirstName: "Bob",
			wantLastName:  "Nested",
			wantEmail:     "bob@example.com",
		},
		{
			name: "empty nested falls back to flat",
			json: `{
				"cardholder_id": "ch_abc",
				"type": "INDIVIDUAL",
				"email": "alice@example.com",
				"individual": {
					"name": {
						"first_name": "",
						"last_name": ""
					}
				},
				"first_name": "Alice",
				"last_name": "Fallback",
				"status": "ACTIVE",
				"created_at": "2024-01-04T00:00:00Z"
			}`,
			wantFirstName: "Alice",
			wantLastName:  "Fallback",
			wantEmail:     "alice@example.com",
		},
		{
			name: "missing individual object uses flat",
			json: `{
				"cardholder_id": "ch_def",
				"type": "INDIVIDUAL",
				"email": "charlie@example.com",
				"first_name": "Charlie",
				"last_name": "NoNested",
				"status": "ACTIVE",
				"created_at": "2024-01-05T00:00:00Z"
			}`,
			wantFirstName: "Charlie",
			wantLastName:  "NoNested",
			wantEmail:     "charlie@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ch Cardholder
			if err := json.Unmarshal([]byte(tt.json), &ch); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if ch.FirstName != tt.wantFirstName {
				t.Errorf("FirstName = %q, want %q", ch.FirstName, tt.wantFirstName)
			}
			if ch.LastName != tt.wantLastName {
				t.Errorf("LastName = %q, want %q", ch.LastName, tt.wantLastName)
			}
			if ch.Email != tt.wantEmail {
				t.Errorf("Email = %q, want %q", ch.Email, tt.wantEmail)
			}
		})
	}
}
