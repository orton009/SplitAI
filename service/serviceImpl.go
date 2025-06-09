package service

import (
	"splitExpense/config"
	"splitExpense/expense"
)

const ErrNilStorage = ("Nil Storage")

type ServiceImpl struct {
	userService    UserService
	expenseService ExpenseService
}

func NewServiceImpl(config *config.Config, storage expense.Storage) *ServiceImpl {
	return &ServiceImpl{
		userService: &UserServiceImpl{
			config:  config,
			storage: storage,
		},
		expenseService: &ExpenseServiceImpl{
			// config:  config,
			storage: storage,
		},
	}
}
