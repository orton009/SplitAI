package expense

import "time"

type User struct {
	ID    int
	Name  string
	Email string
}

type Payer interface {
	isPayer() bool
}

type MultiPayer struct {
	Payers []User
}

func (m *MultiPayer) isPayer() bool {
	return true
}

func (u *User) isPayer() bool {
	return true
}

type Expense struct {
	ID          int
	Description string
	Amount      float64
	CreatedAt   time.Time
	Payee       Payer
	Split       Split
}
