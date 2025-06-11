package apiServer

import (
	"splitExpense/config"
	"splitExpense/orchestrator"

	"github.com/gin-gonic/gin"
)

// RouteHandler implementation for JoinGroup
type JoinGroupRouteHandler struct {
	orchestrator orchestrator.ExpenseAppImpl
}

func (h *JoinGroupRouteHandler) Method() Method {
	return PUT
}

func (h *JoinGroupRouteHandler) Path() string {
	return Path("/group/:id/invite")
}

func (h *JoinGroupRouteHandler) Handle(c *gin.Context, cfg *config.Config) {
	// decode and validation
	type JoinGroupRequest struct {
		MemberId string `json:"new_member_id" binding:"required"`
	}
	var req JoinGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
	}

	userId, err := CtxGetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
	}

	// orchestrator call
	ok, err := h.orchestrator.JoinGroup(userId, req.MemberId, c.Param("id"))
	if err != nil || !ok {
		c.AbortWithError(400, err)
	}

	// response
	c.Status(201)
}

// RouteHandler implementation for CreateGroup
type CreateGroupRouteHandler struct {
	orchestrator orchestrator.ExpenseAppImpl
}

func (h *CreateGroupRouteHandler) Method() Method {
	return POST
}

func (h *CreateGroupRouteHandler) Path() string {
	return Path("/group")
}

func (h *CreateGroupRouteHandler) Handle(c *gin.Context, cfg *config.Config) {
	// decode and validation
	type CreateGroupRequest struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
	}
	userId, err := CtxGetUserId(c)
	if err != nil {
		c.Status(500)
		return
	}

	// orchestrator call
	group, err := h.orchestrator.CreateGroup(userId, req.Name, req.Description)
	if err != nil {
		c.AbortWithError(400, err)
	}
	// response
	c.JSON(201, group)
}

// RouteHandler implementation for CreateGroup
type LeaveGroupRouteHandler struct {
	orchestrator orchestrator.ExpenseAppImpl
}

func (h *LeaveGroupRouteHandler) Method() Method {
	return PUT
}

func (h *LeaveGroupRouteHandler) Path() string {
	return Path("/group/:id/leave")
}

func (h *LeaveGroupRouteHandler) Handle(c *gin.Context, cfg *config.Config) {
	userId, err := CtxGetUserId(c)
	if err != nil {
		c.Status(500)
		return
	}
	// orchestrator call
	_, err = h.orchestrator.LeaveGroup(userId, c.Param("id"))

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	// response
	c.Status(201)
}
