package expense

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"
)

type User struct {
	ID         string
	Name       string
	Email      string
	IsVerified bool
	Password   string
}

type Payer interface {
	GetPayers() map[string]float64
	GetTotal() float64
}

type SinglePayer struct {
	Payer  string  `json:"payer"`
	Amount float64 `json:"amount"`
}

type MultiPayer struct {
	Payers map[string]float64 `json:"payers"`
}

func (m *MultiPayer) GetPayers() map[string]float64 {
	return m.Payers
}

func (m *MultiPayer) GetTotal() float64 {
	var total float64
	for _, amount := range m.Payers {
		total += amount
	}
	return total
}

func (u *SinglePayer) GetPayers() map[string]float64 {
	return map[string]float64{u.Payer: u.Amount}
}

func (u *SinglePayer) GetTotal() float64 {
	return u.Amount
}

// PayerWrapper handles JSON marshaling/unmarshaling of Payer interface
type PayerWrapper struct {
	Payer Payer  `json:"-"`
	Type  string `json:"type"`
}

type PayerJson struct {
	Type       string             `json:"type"`
	PayerSplit map[string]float64 `json:"payerSplit"`
}

// MarshalJSON custom marshaling for Payer interface
func (pw PayerWrapper) MarshalJSON() ([]byte, error) {
	// Marshal the underlying data

	var p PayerJson
	// Determine the type
	switch pw.Payer.(type) {
	case *SinglePayer:
		p.Type = "single"
		payer := pw.Payer.(*SinglePayer).Payer
		p.PayerSplit = map[string]float64{payer: pw.Payer.(*SinglePayer).Amount}

	case *MultiPayer:
		p.Type = "multi"
		p.PayerSplit = pw.Payer.(*MultiPayer).Payers
	default:
		return nil, fmt.Errorf("unknown payer type: %T", pw.Payer)
	}
	return json.Marshal(p)
}

// UnmarshalJSON custom unmarshaling for Payer interface
func (pw *PayerWrapper) UnmarshalJSON(data []byte) error {

	var temp PayerJson

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	switch temp.Type {
	case "single":
		var sp SinglePayer
		payers := lo.Keys(temp.PayerSplit)
		if len(payers) != 1 {
			return errors.New("payers not found")
		}
		sp.Payer = payers[0]
		sp.Amount = temp.PayerSplit[sp.Payer]
		pw.Payer = &sp

	case "multi":
		var mp MultiPayer
		mp.Payers = temp.PayerSplit
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
	SplitW         SplitWrapper
	PayeeW         PayerWrapper
	IsGroupExpense bool
	GroupId        string
}

type Expense struct {
	ID             string
	Description    string
	Amount         float64
	CreatedAt      time.Time
	PayeeW         PayerWrapper
	SplitW         SplitWrapper
	Status         ExpenseStatus
	IsGroupExpense bool
	GroupId        string
	SettledBy      string
	CreatedBy      string
}

// type Expense struct {
// 	ID          string
// 	Description string
// 	Amount      float64
// 	CreatedAt   time.Time
// 	Payee       string
// 	Split       string
// 	Status      ExpenseStatus
// 	SettledBy   string
// 	CreatedBy   string
// }

type SplitJson struct {
	Type            string             `json:"type"`
	TotalAmount     float64            `json:"totalAmount"`
	EqualSplit      []string           `json:"equalSplit"`
	PercentageSplit map[string]float64 `json:"percentageSplit"`
	ShareSplit      map[string]int     `json:"shareSplit"`
	UnitSplit       map[string]float64 `json:"unitSplit"`
}

// func ConvertExpenseToExpense(e *Expense) (*Expense, error) {
// 	// pw := PayerWrapper{Payer: e.Payee}
// 	payeeW, err := json.Marshal(e.Payee)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// sw := SplitWrapper{Split: e.Split}
// 	splitW, err := json.Marshal(e.Split)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &Expense{
// 		ID:          e.ID,
// 		Description: e.Description,
// 		Amount:      e.Amount,
// 		CreatedAt:   e.CreatedAt,
// 		Payee:       string(payeeW),
// 		Split:       string(splitW),
// 		CreatedBy:   e.CreatedBy,
// 		SettledBy:   e.SettledBy,
// 		Status:      e.Status,
// 	}, nil
// }

// func ConvertExpenseToExpense(e *Expense) (*Expense, error) {
// 	// Unmarshal Payee
// 	var pw PayerWrapper
// 	if err := json.Unmarshal([]byte(e.Payee), &pw); err != nil {
// 		return nil, err
// 	}
// 	// Unmarshal Split
// 	var sw SplitWrapper
// 	if err := json.Unmarshal([]byte(e.Split), &sw); err != nil {
// 		return nil, err
// 	}
// 	return &Expense{
// 		ID:          e.ID,
// 		Description: e.Description,
// 		Amount:      e.Amount,
// 		CreatedAt:   e.CreatedAt,
// 		Payee:       pw.Payer,
// 		Split:       sw.Split,
// 		Status:      e.Status,
// 		CreatedBy:   e.CreatedBy,
// 		SettledBy:   e.SettledBy,
// 	}, nil
// }
