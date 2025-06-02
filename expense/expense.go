package expense

import (
	"encoding/json"
	"fmt"
	"time"
)

type User struct {
	ID         string
	Name       string
	Email      string
	IsVerified bool
	Password   string
}

type Payer interface {
	isPayer() bool
}

type SinglePayer struct {
	Payer string `json:"payer"`
}

type MultiPayer struct {
	Payers []string `json:"payers"`
}

func (m *MultiPayer) isPayer() bool {
	return true
}

func (u *SinglePayer) isPayer() bool {
	return true
}

// PayerWrapper handles JSON marshaling/unmarshaling of Payer interface
type PayerWrapper struct {
	Payer Payer           `json:"-"`
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data"`
}

// MarshalJSON custom marshaling for Payer interface
func (pw PayerWrapper) MarshalJSON() ([]byte, error) {
	// Marshal the underlying data
	data, err := json.Marshal(pw.Payer)
	if err != nil {
		return nil, err
	}

	// Determine the type
	var typeName string
	switch pw.Payer.(type) {
	case *SinglePayer:
		typeName = "single"
	case *MultiPayer:
		typeName = "multi"
	default:
		return nil, fmt.Errorf("unknown payer type: %T", pw.Payer)
	}

	return json.Marshal(struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}{
		Type: typeName,
		Data: data,
	})
}

// UnmarshalJSON custom unmarshaling for Payer interface
func (pw *PayerWrapper) UnmarshalJSON(data []byte) error {
	var temp struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	switch temp.Type {
	case "single":
		var sp SinglePayer
		if err := json.Unmarshal(temp.Data, &sp); err != nil {
			return err
		}
		pw.Payer = &sp
	case "multi":
		var mp MultiPayer
		if err := json.Unmarshal(temp.Data, &mp); err != nil {
			return err
		}
		pw.Payer = &mp
	default:
		return fmt.Errorf("unknown payer type: %s", temp.Type)
	}

	pw.Type = temp.Type
	return nil
}

type ExpenseStatus string

const (
	ExpenseDraft    ExpenseStatus = "DRAFT"
	ExpenseSettled  ExpenseStatus = "SETTLED"
	ExpenseReopened ExpenseStatus = "REOPENED"
)

type ExpenseHistory struct {
	Field       string
	OldValue    string
	NewValue    string
	ModifiedBy  string
	UpdatedAt   time.Time
	Description string
}

type ExpenseCreate struct {
	Description    string
	Amount         float64
	Split          Split
	Payee          Payer
	IsGroupExpense bool
	GroupId        string
}

type Expense struct {
	ID             string
	Description    string
	Amount         float64
	CreatedAt      time.Time
	Payee          Payer
	Split          Split
	Status         ExpenseStatus
	IsGroupExpense bool
	GroupId        string
	SettledBy      string
	CreatedBy      string
}

type ExpenseData struct {
	ID          string
	Description string
	Amount      float64
	CreatedAt   time.Time
	Payee       string
	Split       string
	Status      ExpenseStatus
	SettledBy   string
	CreatedBy   string
}
