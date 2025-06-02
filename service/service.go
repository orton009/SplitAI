package service

import (
	expense "splitExpense/expense"
)

type UserHome struct {
	AssociatedGroups []expense.Group
	User             expense.User
	TotalOwed        float64
	TotalBorrowed    float64
	// Friends          []expense.User
}

type Service interface {
	GetUserService() *UserService
	GetExpenseService() *ExpenseService
}

type UserService interface {
	GetUser(id string) (*expense.User, error)
	CreateUser(name string, email string, password string) (*expense.User, error)
	JoinGroup(userId string, groupId string) (bool, error)
	LeaveGroup(userId string, groupId string) (bool, error)
	CreateGroup(userId string, name string, description string) (*expense.Group, error)
	GetAssociatedGroups(userId string) ([]expense.Group, error)
}

type ExpenseService interface {
	CreateExpense(userId string, expense expense.ExpenseCreate) (*expense.Expense, error)
	UpdateExpense(userId string, expense expense.Expense) (*expense.Expense, error)
	DeleteExpense(userId string, expenseId string) (bool, error)
	SettleExpense(userId string, expenseId string) (bool, error)
	// GetExpenseHistory(id string) (*expense.ExpenseHistory, error)
}
