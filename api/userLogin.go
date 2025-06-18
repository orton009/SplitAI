package apiServer

import (
	"errors"
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
	return Path("/friend/add")
}

func (a *AddFriendHandler) Handle(c *gin.Context, cfg *config.Config) {
	type AddFriendRequest struct {
		Email string `json:"email" binding:"required"`
	}
	var req AddFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
		return
	}
	userId, err := CtxGetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
	}

	ok, err := a.o.AddFriend(userId, req.Email)
	if err != nil || !ok {
		c.AbortWithError(500, err)
	}
	c.JSON(201, gin.H{})

}

type UserHomeHandler struct {
	o orchestrator.ExpenseAppImpl
}

func (a *UserHomeHandler) Method() Method {
	return GET
}

func (a *UserHomeHandler) Path() string {
	return Path("/user/home")
}

func (a *UserHomeHandler) Handle(c *gin.Context, cfg *config.Config) {
	userId, err := CtxGetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
	}

	userHome, err := a.o.GetUserHome(userId)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.JSON(200, userHome)
}

type FetchGroupDetailsHandler struct {
	o orchestrator.ExpenseAppImpl
}

func (a *FetchGroupDetailsHandler) Method() Method {
	return GET
}
func (a *FetchGroupDetailsHandler) Path() string {
	return Path("/group/:id")
}
func (a *FetchGroupDetailsHandler) Handle(c *gin.Context, cfg *config.Config) {
	groupId := c.Param("id")
	if groupId == "" {
		c.AbortWithError(400, errors.New("invalid group id"))
		return
	}

	userId, err := CtxGetUserId(c)
	if err != nil {
		c.AbortWithError(400, err)
	}

	group, err := a.o.GetGroupDetail(userId, groupId)
	if err != nil {
		c.AbortWithStatus(400)
		return
	}
	c.JSON(200, group)
}

type GetFriendsHandler struct {
	o orchestrator.ExpenseAppImpl
}

func (h *GetFriendsHandler) Method() Method {
	return GET
}

func (h *GetFriendsHandler) Path() string {
	return Path("/user/friends")
}

func (h *GetFriendsHandler) Handle(c *gin.Context, cfg *config.Config) {
	userId, err := CtxGetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	friends, err := h.o.GetFriends(userId)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.JSON(200, friends)
}
