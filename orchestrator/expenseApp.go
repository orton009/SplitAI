package orchestrator

import (
	"splitExpense/config"
	"splitExpense/expense"
	"splitExpense/service"
	"splitExpense/storage"

	"context"
)

type ExpenseAppImpl struct {
	config         config.Config
	userService    service.UserService
	expenseService service.ExpenseService
}

// Implement service.Service interface
func (e *ExpenseAppImpl) GetUserService() service.UserService {
	return e.userService
}

func (e *ExpenseAppImpl) GetExpenseService() service.ExpenseService {
	return e.expenseService
}

func (e *ExpenseAppImpl) UserSignup(name, email, password string) (*expense.User, *expense.AppError) {
	validator := NewValidator().Email(email).Name(name).Password(password)
	if validator.Ok() {

		user, err := e.userService.CreateUser(name, email, password)
		if err != nil {
			return nil, expense.ErrService(err.Error())
		}
		return user, nil
	} else {
		return nil, validator.Err()
	}
}

func (e *ExpenseAppImpl) JoinGroup(userId, newMemberId, groupId string) (bool, *expense.AppError) {
	_, err := e.userService.GetFriend(userId, newMemberId)
	if err != nil {
		return false, expense.ErrService(err.Error())
	}
	ok, err := e.userService.JoinGroup(userId, groupId)
	if err != nil {
		return false, expense.ErrService(err.Error())
	}
	return ok, nil
}

func (e *ExpenseAppImpl) LeaveGroup(userId, groupId string) (bool, *expense.AppError) {
	ok, err := e.userService.LeaveGroup(userId, groupId)
	if err != nil {
		return false, expense.ErrService(err.Error())
	}

	return ok, nil
}

func (e *ExpenseAppImpl) CreateGroup(userId, name, description string) (*expense.Group, error) {
	return e.userService.CreateGroup(userId, name, description)
}

// func (e *ExpenseAppImpl) GetAssociatedGroups(userId string) ([]expense.Group, error) {
// 	return e.userService.GetAssociatedGroups(userId)
// }

func (e *ExpenseAppImpl) CreateExpense(userId string, exp expense.ExpenseCreate) (*expense.Expense, error) {
	// TODO: check split details, validate that computeTotal() vs amount, check if all userId exist in DB
	// TODO: check if all users are friends of current user.
	return e.expenseService.CreateExpense(userId, exp)
}

func (e *ExpenseAppImpl) UpdateExpense(userId string, exp expense.Expense) (*expense.Expense, error) {
	// TODO: check split details, validate that computeTotal() vs amount, check if all userId exist in DB
	return e.expenseService.UpdateExpense(userId, exp)
}

func (e *ExpenseAppImpl) DeleteExpense(userId string, expenseId string) (bool, error) {
	// check if user is part of expense or not, user should be part of group to delete the expense. user can be just the member of splitDetails but should be part of group if its a group expense.
	return e.expenseService.DeleteExpense(userId, expenseId)
}

func (e *ExpenseAppImpl) SettleExpense(userId string, expenseId string) (*expense.Expense, error) {
	// user should be part of group if group expense, user should always be part of expense or creator of expense.
	return e.expenseService.SettleExpense(userId, expenseId)
}

// Add Login method for orchestrator
func (e *ExpenseAppImpl) Login(email, password string) (*expense.User, error) {
	return e.userService.Login(email, password)
}

// Implement GetUserHome to satisfy ExpenseApp interface
func (e *ExpenseAppImpl) GetUserHome() service.UserHome {
	return service.UserHome{}
}

// func (e *ExpenseAppImpl) GerAssociatedUsers()

// NewExpenseApp creates an ExpenseAppImpl and mocks or creates service dependencies internally
func NewExpenseApp(ctx context.Context, cfg *config.Config) ExpenseAppImpl {
	// For now, create real storage and services, but this can be mocked for tests
	storageImpl := storage.NewDBStorage(&ctx, cfg)
	userService := service.NewUserServiceImpl(cfg, storageImpl)
	expenseService := service.NewExpenseServiceImpl(storageImpl)
	return ExpenseAppImpl{
		userService:    userService,
		expenseService: expenseService,
		config:         *cfg,
	}
}
