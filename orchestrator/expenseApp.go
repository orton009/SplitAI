package orchestrator

import (
	"splitExpense/config"
	"splitExpense/service"
)

type ExpenseApp interface {
	service.UserService
	service.ExpenseService
	GetUserHome() service.UserHome
}

type ExpenseAppImpl struct {
	config config.Config
}
