package orchestrator

import (
	"crypto"
	"encoding/base64"
	"fmt"
	"math"
	"splitExpense/config"
	"splitExpense/expense"
	"splitExpense/service"
	"splitExpense/storage"

	"context"

	lodash "github.com/samber/lo"
)

type ExpenseAppImpl struct {
	config         config.Config
	userService    service.UserService
	expenseService service.ExpenseService
}

// Implement service.Service interface
func (e *ExpenseAppImpl) GetUserService() service.UserService {
	return e.userService
}

func (e *ExpenseAppImpl) GetExpenseService() service.ExpenseService {
	return e.expenseService
}

func (e *ExpenseAppImpl) UserSignup(name, email, password string) (*expense.User, *expense.AppError) {
	validator := NewValidator().Email(email).Name(name).Password(password)
	if validator.Ok() {

		user, err := e.userService.CreateUser(name, email, password)
		if err != nil {
			return nil, expense.ErrService(err.Error())
		}
		return user, nil
	} else {
		return nil, validator.Err()
	}
}

func (e *ExpenseAppImpl) AddFriend(userId, friendEmail string) (bool, error) {
	validator := NewValidator().NonEmptyID(userId).Email(friendEmail)
	if !validator.Ok() {
		return false, validator.Err()
	}

	friend, err := e.userService.FetchUserCredentials(friendEmail)
	if err != nil {
		return false, expense.ErrService(err.Error())
	}

	ok, err := e.userService.AddFriend(userId, friend.ID)
	if err != nil {
		return false, expense.ErrService(err.Error())
	}
	return ok, nil
}

func (e *ExpenseAppImpl) JoinGroup(userId, newMemberId, groupId string) (bool, error) {
	validator := NewValidator().NonEmptyID(userId).NonEmptyID(newMemberId).NonEmptyID(groupId)
	if !validator.Ok() {
		return false, validator.Err()
	}

	_, err := e.userService.GetFriend(userId, newMemberId)
	if err != nil {
		return false, expense.ErrService(err.Error())
	}
	ok, err := e.userService.JoinGroup(newMemberId, groupId)
	if err != nil {
		return false, expense.ErrService(err.Error())
	}
	return ok, nil
}

func (e *ExpenseAppImpl) LeaveGroup(userId, groupId string) (bool, *expense.AppError) {
	validator := NewValidator().NonEmptyID(userId).NonEmptyID(groupId)
	if !validator.Ok() {
		return false, validator.Err()
	}

	ok, err := e.userService.LeaveGroup(userId, groupId)
	if err != nil {
		return false, expense.ErrService(err.Error())
	}

	return ok, nil
}

func (e *ExpenseAppImpl) CreateGroup(userId, name, description string) (*expense.Group, error) {
	validator := NewValidator().NonEmptyID(userId).Name(name)
	if !validator.Ok() {
		return nil, validator.Err()
	}
	return e.userService.CreateGroup(userId, name, description)
}

func (e *ExpenseAppImpl) verifyAmount(a1 float64, a2 float64) bool {
	return math.Round(a1*100)/100 == math.Round(a2*100)/100
}

func (e *ExpenseAppImpl) CreateExpense(userId string, exp expense.ExpenseCreate) (*expense.Expense, error) {
	validator := NewValidator().NonEmptyID(userId).LeastAmount(exp.Amount)
	if !validator.Ok() {
		return nil, validator.Err()
	}

	payers := lodash.Keys(exp.PayeeW.Payer.GetPayers())
	if len(payers) == 0 {
		return nil, expense.ErrValidation("payers are required")
	}

	friends, err := e.GetUserService().GetFriends(userId)
	if err != nil {
		return nil, err
	}
	friendNetwork := lodash.Union(lodash.Map(friends, func(u expense.User, _ int) string { return u.ID }), []string{userId})

	expenseMembers := lodash.Union(lodash.Keys(exp.PayeeW.Payer.GetPayers()), lodash.Keys(exp.SplitW.Split.GetPayeeSplit()))
	areValidFriends := lodash.ContainsBy(expenseMembers, func(uid string) bool {
		return lodash.Contains(friendNetwork, uid)
	})

	if !areValidFriends {
		return nil, expense.ErrValidation("all expense members should be friends of expense creator")
	}

	if !e.verifyAmount(exp.Amount, exp.SplitW.Split.ComputeTotal()) {
		fmt.Println("amount: ", exp.Amount, exp.SplitW.Split.ComputeTotal(), exp.SplitW.Split.GetPayeeSplit())
		return nil, expense.ErrValidation("split amount is not same as expense amount")
	}

	if !e.verifyAmount(exp.PayeeW.Payer.GetTotal(), exp.Amount) {
		return nil, expense.ErrValidation("payer contribution total is not same as expense amount")
	}

	createdExp, err := e.expenseService.CreateExpense(userId, exp)
	if err != nil {
		return nil, expense.ErrService(err.Error())
	}
	return createdExp, nil
}

func (e *ExpenseAppImpl) UpdateExpense(userId string, exp expense.Expense) (*expense.Expense, error) {

	// fetch existing exp from db
	existingExp, err := e.GetExpenseService().FetchExpense(exp.ID)
	if err != nil {
		return nil, expense.ErrValidation("expense not found")
	}

	if existingExp.Status != expense.ExpenseDraft {
		return nil, expense.ErrValidation("expense is not in draft state, cannot update")
	}

	if existingExp.GroupId != exp.GroupId {
		return nil, expense.ErrValidation("expense group id cannot be changed")
	}

	// check is user is allowed to update expense
	// only group members or expense members or expense creator is allowed
	userAllowed := existingExp.CreatedBy == userId
	if existingExp.IsGroupExpense {
		groupMembers, err := e.GetUserService().GetAssociatedUsers(existingExp.GroupId)
		if err != nil {
			return nil, err
		}

		userAllowed = userAllowed || lodash.ContainsBy(groupMembers.Users, func(u expense.User) bool { return u.ID == userId })
	}
	existingExpenseUsers := lodash.Union(lodash.Keys(existingExp.PayeeW.Payer.GetPayers()), lodash.Keys(existingExp.SplitW.Split.GetPayeeSplit()))
	userAllowed = userAllowed || lodash.ContainsBy(existingExpenseUsers, func(uid string) bool { return uid == userId })
	if !userAllowed {
		return nil, expense.ErrValidation("user is not authorised to update expense, only members of this expense or members of group can edit this")
	}

	// validate amount, payee total and split total
	validator := NewValidator().LeastAmount(exp.Amount)
	if !validator.Ok() {
		return nil, validator.Err()
	}
	if !e.verifyAmount(exp.Amount, exp.PayeeW.Payer.GetTotal()) || !e.verifyAmount(exp.Amount, exp.SplitW.Split.ComputeTotal()) {
		return nil, expense.ErrValidation("amount mismatch between payer or split when compared with expense amount")
	}

	friends, err := e.GetUserService().GetFriends(userId)
	if err != nil {
		return nil, err
	}

	// validate is all expense members are friends of expense creator
	expUsers := lodash.Union(lodash.Keys(exp.PayeeW.Payer.GetPayers()), lodash.Keys(exp.SplitW.Split.GetPayeeSplit()))
	areValidFriends := lodash.ContainsBy(expUsers, func(uid string) bool {
		return lodash.ContainsBy(friends, func(f expense.User) bool { return f.ID == uid })
	})
	if !areValidFriends {
		return nil, expense.ErrValidation("users part of expense should be friends of current user")
	}

	// updating selected fields from existing expense
	expenseUpdate := *existingExp
	expenseUpdate.PayeeW = exp.PayeeW
	expenseUpdate.SplitW = exp.SplitW
	expenseUpdate.Description = exp.Description
	expenseUpdate.Amount = exp.Amount

	return e.expenseService.UpdateExpense(userId, expenseUpdate)
}

func (e *ExpenseAppImpl) DeleteExpense(userId string, expenseId string) (bool, error) {
	// check if user is part of expense or not, user should be part of group to delete the expense. user can be just the member of splitDetails but should be part of group if its a group expense.
	return e.expenseService.DeleteExpense(userId, expenseId)
}

func (e *ExpenseAppImpl) DeleteGroup(userId string, groupId string) (bool, error) {
	// user should be part of group to delete the group, user can be just the member of splitDetails but should be part of group if its a group expense.
	validator := NewValidator().NonEmptyID(userId).NonEmptyID(groupId)
	if !validator.Ok() {
		return false, validator.Err()
	}
	group, err := e.userService.GetGroupById(groupId)
	if group.Admin != userId {
		return false, expense.ErrValidation("user is not admin of the group, cannot delete group")
	}

	ok, err := e.userService.DeleteGroup(groupId)
	if err != nil {
		return false, expense.ErrService(err.Error())
	}
	return ok, nil
}

func (e *ExpenseAppImpl) SettleExpense(userId string, expenseId string) (*expense.Expense, error) {
	// user should be part of group if group expense, user should always be part of expense or creator of expense.
	return e.expenseService.SettleExpense(userId, expenseId)
}

// Add Login method for orchestrator
func (e *ExpenseAppImpl) Login(email, password string) (*expense.User, error) {

	validator := NewValidator().Email(email).Password(password)
	if !validator.Ok() {
		return nil, validator.Err()
	}

	user, err := e.GetUserService().FetchUserCredentials(email)
	if err != nil {
		return nil, err
	}

	hasher := crypto.SHA256.New()
	hasher.Write([]byte(password))
	passwordHash := hasher.Sum(nil)
	if user.Password != base64.StdEncoding.EncodeToString(passwordHash) {
		return nil, expense.ErrValidation("invalid password")
	}
	return user, nil
}

// Implement GetUserHome to satisfy ExpenseApp interface
func (e *ExpenseAppImpl) GetUserHome(userId string) (service.UserHome, error) {
	var home service.UserHome
	user, err := e.userService.GetUser(userId)
	if err != nil {
		return home, err
	}

	groups, err := e.userService.GetAssociatedGroups(userId)
	if err != nil {
		return home, err
	}

	expGroups := []service.GroupWithExpense{}

	for _, group := range groups {
		// TODO: add go routines
		exp, err := e.expenseService.FetchExpenseByGroup(userId, group.Id, 0)
		if err != nil {
			return home, err
		}

		// TODO: move to different API, lot of db calls
		totalOwed, totalBorrowed, err := e.expenseService.CalculateUserRunningExpensesInGroup(userId, &group)
		if err != nil {
			return home, err
		}

		expGroups = append(expGroups, service.GroupWithExpense{
			Group:          group,
			ExpenseHistory: *exp,
			TotalOwed:      totalOwed,
			TotalBorrowed:  totalBorrowed,
		})

	}

	return service.UserHome{AssociatedGroups: expGroups, User: *user}, nil
}

func (e *ExpenseAppImpl) GetUserExpenseHistory(userId string, pageNumber int) (*service.UserExpenses, error) {

	// Fetch active user expenses
	expHistory, err := e.expenseService.FetchActiveUserExpenses(userId, pageNumber)
	if err != nil {
		return nil, err
	}

	totalOwed, totalBorrowed, err := e.expenseService.CalculateAllUserRunningExpenses(userId)
	if err != nil {
		return nil, expense.ErrService(err.Error())
	}

	history := service.UserExpenses{
		Expenses:      expHistory.Expenses,
		TotalOwed:     totalOwed,
		TotalBorrowed: totalBorrowed,
		PageNumber:    expHistory.PageNumber,
		TotalPages:    expHistory.TotalPages,
	}

	return &history, nil
}

func (e *ExpenseAppImpl) GetGroupDetail(userId string, groupId string) (service.GroupDetail, error) {
	var detail service.GroupDetail

	group, err := e.userService.GetGroupById(groupId)
	if err != nil {
		return detail, err
	}

	users, err := e.userService.GetAssociatedUsers(groupId)
	if err != nil {
		return detail, err
	}

	expHistory, err := e.expenseService.FetchExpenseByGroup(userId, groupId, 0)
	if err != nil {
		return detail, err
	}

	totalOwed, totalBorrowed, err := e.expenseService.CalculateUserRunningExpensesInGroup(userId, group)
	if err != nil {
		return detail, err
	}

	detail = service.GroupDetail{
		Group: service.GroupWithExpense{
			Group:          *group,
			ExpenseHistory: *expHistory,
			TotalOwed:      totalOwed,
			TotalBorrowed:  totalBorrowed,
		},
		GroupMembers: users.Users,
	}

	return detail, nil
}

// NewExpenseApp creates an ExpenseAppImpl and mocks or creates service dependencies internally
func NewExpenseApp(ctx context.Context, cfg *config.Config) ExpenseAppImpl {
	// For now, create real storage and services, but this can be mocked for tests
	storageImpl := storage.NewDBStorage(&ctx, cfg)
	userService := service.NewUserServiceImpl(cfg, storageImpl)
	expenseService := service.NewExpenseServiceImpl(storageImpl)
	return ExpenseAppImpl{
		userService:    userService,
		expenseService: expenseService,
		config:         *cfg,
	}
}

func (e *ExpenseAppImpl) GetFriends(userId string) ([]expense.User, error) {
	return e.userService.GetFriends(userId)
}

func calculateUserLiability(g *expense.GroupExpenseHistory) (totalOwed, totalBorrowed float64) {
	totalOwed, totalBorrowed = 0.0, 0.0

	for _, exp := range g.Expenses {
		totalBorrowed += exp.TotalBorrowed
		totalOwed += exp.TotalOwed
	}

	if totalOwed > totalBorrowed {
		totalOwed -= totalBorrowed
		totalBorrowed = 0.0
	} else {
		totalBorrowed -= totalOwed
		totalOwed = 0.0
	}

	return
}
