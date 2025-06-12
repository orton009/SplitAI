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
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.AbortWithError(400, err)
	}

	// orchestrator call
	user, err := h.orchestrator.Login(req.Email, req.Password)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	token, err := expense.GenerateToken(*user)
	if err != nil {
		c.AbortWithError(400, err)
	}
	c.SetCookie("token", token, 3600, "/", "", false, true)

	// response
	c.JSON(200, gin.H{"token": token, "user": user})
}

type AddFriendHandler struct {
	o orchestrator.ExpenseAppImpl
}

func (a *AddFriendHandler) Method() Method {
	return PUT
}

func (a *AddFriendHandler) Path() string {
	return Path("/friend/:id/add")
}

func (a *AddFriendHandler) Handle(c *gin.Context, cfg *config.Config) {
	friendId := c.Param("id")
	userId, err := CtxGetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
	}

	ok, err := a.o.AddFriend(userId, friendId)
	if err != nil || !ok {
		c.AbortWithError(500, err)
	}
	c.Status(201)

}
