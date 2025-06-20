package apiServer

import (
	"errors"
	"splitExpense/config"
	"splitExpense/expense"
	"splitExpense/orchestrator"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreateOrUpdateExpenseRouteHandler struct {
	orchestrator orchestrator.ExpenseAppImpl
}

func (h *CreateOrUpdateExpenseRouteHandler) Method() Method {
	return POST
}

func (h *CreateOrUpdateExpenseRouteHandler) Path() string {
	return Path("/expense")
}

func (h *CreateOrUpdateExpenseRouteHandler) Handle(c *gin.Context, cfg *config.Config) {

	userId, err := CtxGetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
	}

	type CreateOrUpdateExpenseRequest struct {
		ID          string               `json:"id"`
		Description string               `json:"description"`
		Amount      float64              `json:"amount"`
		Split       expense.SplitWrapper `json:"split"`
		Payee       expense.PayerWrapper `json:"payee"`
		GroupId     string               `json:"groupId"`
	}
	var req CreateOrUpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(400, err)
	}

	var updatedExpense *expense.Expense
	if req.ID != "" {
		// update request
		updatedExpense, err = h.orchestrator.UpdateExpense(userId, expense.Expense{
			ID:          req.ID,
			Description: req.Description,
			Amount:      req.Amount,
			SplitW:      req.Split,
			PayeeW:      req.Payee,
		})

	} else {

		// create request
		updatedExpense, err = h.orchestrator.CreateExpense(userId, expense.ExpenseCreate{
			Description:    req.Description,
			Amount:         req.Amount,
			SplitW:         req.Split,
			PayeeW:         req.Payee,
			IsGroupExpense: len(req.GroupId) > 0,
			GroupId:        req.GroupId,
		})

	}

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(201, *updatedExpense)

}

type DeleteExpenseRouteHandler struct {
	orchestrator orchestrator.ExpenseAppImpl
}

func (d *DeleteExpenseRouteHandler) Method() Method {
	return DELETE
}

func (d *DeleteExpenseRouteHandler) Path() string {
	return Path("/expense/:id")
}

func (d *DeleteExpenseRouteHandler) Handle(c *gin.Context, cfg *config.Config) {
	userId, err := CtxGetUserId(c)
	if err != nil {
		c.Status(500)
		return
	}

	_, err = d.orchestrator.DeleteExpense(userId, c.Param("id"))
	if err != nil {
		c.AbortWithError(500, errors.Join(errors.New("could not delete expense"), err))
		return
	}

	c.JSON(201, gin.H{})
}

type SettleExpenseHandler struct {
	orchestrator orchestrator.ExpenseAppImpl
}

func (h *SettleExpenseHandler) Method() Method {
	return PUT
}

func (h *SettleExpenseHandler) Path() string {
	return Path("/expense/:id/settle")
}

func (h *SettleExpenseHandler) Handle(c *gin.Context, cfg *config.Config) {
	userId, err := CtxGetUserId(c)
	if err != nil {
		c.Status(500)
		return
	}

	_, err = h.orchestrator.SettleExpense(userId, c.Param("id"))
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(201, gin.H{})
}

type UserExpensesHandler struct {
	o orchestrator.ExpenseAppImpl
}

func (h *UserExpensesHandler) Method() Method {
	return GET
}

func (h *UserExpensesHandler) Path() string {
	return Path("/expenses")
}

func (h *UserExpensesHandler) Handle(c *gin.Context, cfg *config.Config) {
	userId, err := CtxGetUserId(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	pageNumber := c.Query("pageNumber")
	page, _ := strconv.Atoi(pageNumber)

	history, err := h.o.GetUserExpenseHistory(userId, page)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, history)
}
