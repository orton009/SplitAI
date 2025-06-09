package service

import (
	"errors"
	"splitExpense/expense"
	"time"

	"github.com/google/uuid"
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

	exp := expense.Expense{
		ID:          uuid.New().String(),
		Description: expenseCreate.Description,
		Split:       expenseCreate.Split,
		CreatedAt:   time.Now(),
		Payee:       expenseCreate.Payee,
		Amount:      expenseCreate.Amount,
		Status:      expense.ExpenseDraft,
	}

	expData, err := expense.ConvertExpenseToExpenseData(&exp)
	if err != nil {
		return nil, err
	}
	expData, err = e.storage.CreateOrUpdateExpense(*expData)
	if err != nil {
		return nil, err
	}
	return expense.ConvertExpenseDataToExpense(expData)
}

func (e *ExpenseServiceImpl) UpdateExpense(userId string, exp expense.Expense) (*expense.Expense, error) {
	expense_, err := e.storage.FetchExpense(exp.ID)
	if err != nil {
		return nil, errors.Join(errors.New("expense doesn't exist"), err)
	}
	if exp.CreatedAt.IsZero() {
		exp.CreatedAt = time.Now()
	}
	if expense_.CreatedAt != exp.CreatedAt {
		return nil, errors.New("cannot change createdAt field")
	}
	expData, err := expense.ConvertExpenseToExpenseData(&exp)
	if err != nil {
		return nil, err
	}
	expenseData, err := e.storage.CreateOrUpdateExpense(*expData)
	if err != nil {
		return nil, err
	}
	return expense.ConvertExpenseDataToExpense(expenseData)
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
	updatedData, err := e.storage.CreateOrUpdateExpense(*expenseData)
	if err != nil {
		return nil, err
	}
	return expense.ConvertExpenseDataToExpense(updatedData)
}

func NewExpenseServiceImpl(storage expense.Storage) *ExpenseServiceImpl {
	return &ExpenseServiceImpl{
		storage: storage,
	}
}
