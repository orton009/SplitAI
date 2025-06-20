package service

import (
	"errors"
	"splitExpense/expense"
	"time"

	"github.com/google/uuid"
	lodash "github.com/samber/lo"
)

type ExpenseServiceImpl struct {
	// config  *config.Config
	storage expense.Storage
}

// TODO: Record expense update in expense history

func (e *ExpenseServiceImpl) CreateExpense(userId string, expenseCreate expense.ExpenseCreate) (*expense.Expense, error) {
	_, err := e.storage.FetchUserById(userId)
	if err != nil {
		return nil, errors.Join(errors.New("user trying to create expense does not exist"))
	}

	// validate group expense, check if group exists
	if expenseCreate.IsGroupExpense {
		_, err := e.storage.FetchGroupById(expenseCreate.GroupId)
		if err != nil {
			return nil, err
		}
	}
	payeeMap := expenseCreate.SplitW.Split.GetPayeeSplit()

	exp := expense.Expense{
		ID:             uuid.New().String(),
		Description:    expenseCreate.Description,
		SplitW:         expenseCreate.SplitW,
		CreatedAt:      time.Now(),
		PayeeW:         expenseCreate.PayeeW,
		Amount:         expenseCreate.Amount,
		Status:         expense.ExpenseDraft,
		CreatedBy:      userId,
		IsGroupExpense: expenseCreate.IsGroupExpense,
		GroupId:        expenseCreate.GroupId,
	}

	// TODO: Add transaction LOCK
	expData, err := e.storage.CreateOrUpdateExpense(exp)
	if err != nil {
		return nil, err
	}

	userIds := lodash.Union([]string{userId}, lodash.Keys(payeeMap), lodash.Keys(expenseCreate.PayeeW.Payer.GetPayers()))

	if expenseCreate.IsGroupExpense {
		for _, userId := range userIds {

			_, err := e.storage.AddExpenseMapping(expData.ID, userId)
			if err != nil {
				return nil, err
			}
		}
	}

	// TODO: CLOSE LOCK

	return expData, nil
}

func (e *ExpenseServiceImpl) UpdateExpense(userId string, exp expense.Expense) (*expense.Expense, error) {
	existingExp, err := e.storage.FetchExpense(exp.ID)
	if err != nil {
		return nil, err
	}

	if existingExp.CreatedAt != exp.CreatedAt || existingExp.CreatedBy != exp.CreatedBy {
		return nil, errors.New("protected field change")
	}

	existingPayers := lodash.Keys(existingExp.PayeeW.Payer.GetPayers())
	newPayers := lodash.Keys(exp.PayeeW.Payer.GetPayers())
	payersToRemove, payersToAdd := lodash.Difference(existingPayers, newPayers)

	// borrowers
	existingB := lodash.Keys(existingExp.SplitW.Split.GetPayeeSplit())
	newB := lodash.Keys(exp.SplitW.Split.GetPayeeSplit())
	removeB, addB := lodash.Difference(existingB, newB)

	usersToAdd := lodash.Union(addB, payersToAdd)
	// only remove users who are not borrower or payer
	usersToRemove, _ := lodash.Difference(lodash.Union(removeB, payersToRemove), lodash.Union(newPayers, newB))

	// TODO: ADD LOCK
	for _, userId := range usersToAdd {
		_, err := e.storage.AddExpenseMapping(exp.ID, userId)
		if err != nil {
			return nil, err
		}
	}

	_, err = e.storage.RemoveUsersFromExpense(exp.ID, usersToRemove)
	if err != nil {
		return nil, err
	}

	updatedExp, err := e.storage.CreateOrUpdateExpense(exp)
	if err != nil {
		return nil, err
	}

	// REMOVE LOCK

	return updatedExp, err
}

func (e *ExpenseServiceImpl) DeleteExpense(userId string, expenseId string) (bool, error) {
	exp, err := e.storage.FetchExpense(expenseId)
	if err != nil {
		return false, err
	}

	userIds := lodash.Union(lodash.Keys(exp.PayeeW.Payer.GetPayers()), lodash.Keys(exp.SplitW.Split.GetPayeeSplit()))

	// TODO: ADD LOCK
	_, err = e.storage.RemoveUsersFromExpense(exp.ID, userIds)
	if err != nil {
		return false, err
	}
	return e.storage.DeleteExpense(expenseId)
}

func (e *ExpenseServiceImpl) SettleExpense(userId string, expenseId string) (*expense.Expense, error) {
	Expense, err := e.storage.FetchExpense(expenseId)
	if err != nil {
		return nil, err
	}
	Expense.Status = expense.ExpenseSettled
	Expense.SettledBy = userId
	updatedData, err := e.storage.CreateOrUpdateExpense(*Expense)
	if err != nil {
		return nil, err
	}
	return updatedData, nil
}

func (e *ExpenseServiceImpl) FetchExpense(id string) (*expense.Expense, error) {
	return e.storage.FetchExpense(id)
}

func NewExpenseServiceImpl(storage expense.Storage) *ExpenseServiceImpl {
	return &ExpenseServiceImpl{
		storage: storage,
	}
}

// TODO: have a thread to fetch in background, also use streams alternative for data processing
func (e *ExpenseServiceImpl) CalculateUserRunningExpensesInGroup(userId string, group *expense.Group) (float64, float64, error) {
	pageNumber := 1
	totalPayed, totalBorrowed := 0.0, 0.0

	for {
		stored, err := e.storage.FetchGroupExpensesByStatus(group.Id, expense.ExpenseDraft, pageNumber)
		if err != nil {
			return 0, 0, err
		}

		for _, exp := range stored.Expenses {
			payed := exp.PayeeW.Payer.GetPayers()[userId]
			borrowed := exp.SplitW.Split.GetPayeeSplit()[userId]
			if payed > borrowed {
				totalPayed += payed - borrowed
			} else if payed < borrowed {
				totalBorrowed += borrowed - payed
			}
		}

		if pageNumber >= stored.TotalPages {
			break
		}
		pageNumber++
	}

	return totalPayed, totalBorrowed, nil
}

func (e *ExpenseServiceImpl) FetchExpenseByGroup(userId string, groupId string, pageNumber int) (*expense.GroupExpenseHistory, error) {
	if pageNumber == 0 {
		pageNumber = 1
	}

	stored, err := e.storage.FetchGroupExpenses(groupId, pageNumber)
	if err != nil {
		return nil, err
	}

	result := &expense.GroupExpenseHistory{Expenses: []expense.DetailedExpense{}, PageNumber: stored.PageNumber, TotalPages: stored.TotalPages}

	totalPayed, totalBorrowed := 0.0, 0.0
	for _, exp := range stored.Expenses {
		payed := exp.PayeeW.Payer.GetPayers()[userId]
		borrowed := exp.SplitW.Split.GetPayeeSplit()[userId]
		if payed > borrowed {
			result.Expenses = append(result.Expenses, expense.DetailedExpense{Expense: exp, TotalOwed: payed - borrowed})
		} else if payed < borrowed {
			result.Expenses = append(result.Expenses, expense.DetailedExpense{Expense: exp, TotalBorrowed: borrowed - payed})
		}

		totalPayed += payed
		totalBorrowed += borrowed
	}
	return result, err
}

func (e *ExpenseServiceImpl) FetchExpenseCountByGroup(groupId string) (int, error) {
	_, err := e.storage.FetchGroupById(groupId)
	if err != nil {
		return 0, err
	}
	return e.storage.FetchExpenseCountByGroup(groupId)
}

func (e *ExpenseServiceImpl) CalculateAllUserRunningExpenses(userId string) (float64, float64, error) {
	pageNumber := 1
	totalPayed, totalBorrowed := 0.0, 0.0

	for {
		stored, err := e.storage.FetchExpenseByUserAndStatus(userId, expense.ExpenseDraft, pageNumber, 100)
		if err != nil {
			return 0, 0, err
		}

		for _, exp := range stored.Expenses {
			payed := exp.PayeeW.Payer.GetPayers()[userId]
			borrowed := exp.SplitW.Split.GetPayeeSplit()[userId]
			if payed > borrowed {
				totalPayed += payed - borrowed
			} else if payed < borrowed {
				totalBorrowed += borrowed - payed
			}
		}

		if pageNumber >= stored.TotalPages {
			break
		}
		pageNumber++
	}

	return totalPayed, totalBorrowed, nil
}

func (e *ExpenseServiceImpl) FetchActiveUserExpenses(userId string, pageNumber int) (*expense.GroupExpenseHistory, error) {
	if pageNumber == 0 {
		pageNumber = 1
	}

	stored, err := e.storage.FetchExpenseByUserAndStatus(userId, expense.ExpenseDraft, pageNumber, 100)
	if err != nil {
		return nil, err
	}

	result := &expense.GroupExpenseHistory{Expenses: []expense.DetailedExpense{}, PageNumber: stored.PageNumber, TotalPages: stored.TotalPages}

	for _, exp := range stored.Expenses {
		payed := exp.PayeeW.Payer.GetPayers()[userId]
		borrowed := exp.SplitW.Split.GetPayeeSplit()[userId]
		if payed > borrowed {
			result.Expenses = append(result.Expenses, expense.DetailedExpense{Expense: exp, TotalOwed: payed - borrowed})
		} else if payed < borrowed {
			result.Expenses = append(result.Expenses, expense.DetailedExpense{Expense: exp, TotalBorrowed: borrowed - payed})
		}
	}

	return result, nil
}
