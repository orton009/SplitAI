package apiServer

import (
	"path/filepath"
	"splitExpense/config"
	"splitExpense/expense"
	"splitExpense/orchestrator"

	"github.com/gin-gonic/gin"
)

func Path(path string) string {
	return filepath.Join("v1", path)
}

type LoginRouteHandler struct {
	orchestrator orchestrator.ExpenseAppImpl
}

func (h *LoginRouteHandler) Method() Method {
	return POST
}

func (h *LoginRouteHandler) Path() string {
	return Path("/user/login")
}

func (h *LoginRouteHandler) Handle(c *gin.Context, cfg *config.Config) {

	// decode and validation

	type LoginRequest struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	// orchestrator call

	user, err := h.orchestrator.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	token, err := expense.GenerateToken(*user)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not generate token"})
		return
	}
	c.SetCookie("token", token, 3600, "/", "", false, true)

	// response
	c.JSON(200, gin.H{"token": token, "user": user})
}
