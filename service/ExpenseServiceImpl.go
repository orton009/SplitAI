package service

import (
	"errors"
	"splitExpense/expense"
	"time"
)

type ExpenseServiceImpl struct {
	// config  *config.Config
	storage expense.Storage
}

// TODO: Record expense update in expense history

func (e *ExpenseServiceImpl) CreateExpense(userId string, expenseCreate expense.ExpenseCreate) (*expense.Expense, error) {
	_, err := e.storage.FetchUserById(userId)
	if err != nil {
		return nil, err
	}

	if expenseCreate.IsGroupExpense && expenseCreate.GroupId == "" {
		return nil, errors.New("invalid Group ID")
	}

	_, err = e.storage.FetchGroupById(expenseCreate.GroupId)
	if err != nil {
		return nil, err
	}

	splitW := expense.SplitWrapper{Split: expenseCreate.Split}
	splitJson, err := splitW.MarshalJSON()
	if err != nil {
		return nil, err
	}

	payerW := expense.PayerWrapper{Payer: expenseCreate.Payee}
	payeeJson, err := payerW.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return e.storage.CreateOrUpdateExpense(expense.ExpenseData{
		Description: expenseCreate.Description,
		Split:       string(splitJson),
		CreatedAt:   time.Now(),
		Payee:       string(payeeJson),
		Amount:      expenseCreate.Amount,
		Status:      expense.ExpenseDraft,
	})
}

func (e *ExpenseServiceImpl) UpdateExpense(userId string, expense expense.Expense) (*expense.Expense, error) {
	// TODO: IMPL, add/remove users from expense
	return nil, nil
}

func (e *ExpenseServiceImpl) DeleteExpense(userId string, expenseId string) (bool, error) {
	return e.storage.DeleteExpense(expenseId)
}

func (e *ExpenseServiceImpl) SettleExpense(userId string, expenseId string) (*expense.Expense, error) {
	expenseData, err := e.storage.FetchExpense(expenseId)
	if err != nil {
		return nil, err
	}
	expenseData.Status = expense.ExpenseSettled
	return e.storage.CreateOrUpdateExpense(*expenseData)
}
