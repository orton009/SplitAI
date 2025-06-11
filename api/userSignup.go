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
	orchestrator orchestrator.ExpenseAppImpl
}

func (u *userSignUpHandler) Method() Method {
	return POST
}

func (u *userSignUpHandler) Path() string {
	return Path("/user/signup")
}

func (u *userSignUpHandler) Handle(c *gin.Context, cfg *config.Config) {
	// decode and validation
	var req signUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	// orchestrator call
	user, err := u.orchestrator.UserSignup(req.Name, req.Email, req.Password)
	if err != nil {
		c.AbortWithError(500, err)
		c.JSON(500, gin.H{"error": "could not create user"})
		return
	}

	// response
	c.JSON(201, user)
}
