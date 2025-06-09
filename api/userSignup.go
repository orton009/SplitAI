package apiServer

import (
	"splitExpense/config"
	"splitExpense/orchestrator"

	"github.com/gin-gonic/gin"
)

type signUpRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userSignUpHandler struct {
	orchestrator *orchestrator.ExpenseAppImpl
}

func (u *userSignUpHandler) Method() Method {
	return POST
}

func (u *userSignUpHandler) Path() string {
	return Path("/user/signup")
}

func (u *userSignUpHandler) Handle(ctx *gin.Context, cfg *config.Config) {

}
