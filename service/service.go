package service

import (
	expense "splitExpense/expense"
)

type UserExpenses struct {
	Expenses      []expense.DetailedExpense `json:"expenses"`
	TotalOwed     float64                   `json:"totalOwed"`
	TotalBorrowed float64                   `json:"totalBorrowed"`
	PageNumber    int                       `json:"pageNumber"`
	TotalPages    int                       `json:"totalPages"`
}

type GroupWithExpense struct {
	Group          expense.Group               `json:"group"`
	ExpenseHistory expense.GroupExpenseHistory `json:"expenseHistory"`
	TotalOwed      float64                     `json:"totalOwed"`
	TotalBorrowed  float64                     `json:"totalBorrowed"`
}

type UserHome struct {
	AssociatedGroups  []GroupWithExpense `json:"associatedGroups"`
	User              expense.User       `json:"user"`
	UserTotalOwed     float64            `json:"userTotalOwed"`
	UserTotalBorrowed float64            `json:"userTotalBorrowed"`
}

type GroupDetail struct {
	Group        GroupWithExpense `json:"group"`
	GroupMembers []expense.User   `json:"groupMembers"`
}

type Service interface {
	GetUserService() *UserService
	GetExpenseService() *ExpenseService
}

type UserService interface {
	GetUser(id string) (*expense.User, error)
	CreateUser(name string, email string, password string) (*expense.User, error)
	AddFriend(userId string, friendId string) (bool, error)
	GetFriends(userId string) ([]expense.User, error)
	GetFriend(userId string, friendId string) (*expense.User, error)
	JoinGroup(userId string, groupId string) (bool, error)
	LeaveGroup(userId string, groupId string) (bool, error)
	CreateGroup(userId string, name string, description string) (*expense.Group, error)
	GetAssociatedGroups(userId string) ([]expense.Group, error)
	FetchUserCredentials(email string) (*expense.User, error)
	GetAssociatedUsers(groupId string) (*AssociatedUsers, error)
	GetGroupById(groupId string) (*expense.Group, error)
}

type ExpenseService interface {
	FetchExpense(id string) (*expense.Expense, error)
	CreateExpense(userId string, expense expense.ExpenseCreate) (*expense.Expense, error)
	UpdateExpense(userId string, expense expense.Expense) (*expense.Expense, error)
	DeleteExpense(userId string, expenseId string) (bool, error)
	SettleExpense(userId string, expenseId string) (*expense.Expense, error)
	FetchExpenseByGroup(userId string, groupId string, pageNumber int) (*expense.GroupExpenseHistory, error)
	FetchExpenseCountByGroup(groupId string) (int, error)
	FetchActiveUserExpenses(userId string, pageNumber int) (*expense.GroupExpenseHistory, error)
	// GetExpenseHistory(id string) (*expense.ExpenseHistory, error)
}
