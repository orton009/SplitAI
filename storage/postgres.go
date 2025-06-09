package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"splitExpense/config"
	"splitExpense/db"
	"splitExpense/expense"
	models "splitExpense/expense"

	"strconv"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// NewPostgresDB establishes a new PostgreSQL connection using config values
func NewPostgresDB(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseName,
		cfg.DatabaseSSLMode,
	)
	return sql.Open("postgres", dsn)
}

type DBStorage struct {
	ctx     *context.Context
	db      *sql.DB
	queries *db.Queries
	config  *config.Config
}

func NewDBStorage(ctx *context.Context, config *config.Config) *DBStorage {
	pg, err := NewPostgresDB(config)
	if err != nil {
		log.Error().Err(err).Msg("error creating postgres connection")
		return nil
	}
	return &DBStorage{
		ctx:     ctx,
		db:      pg,
		queries: db.New(pg),
		config:  config,
	}
}

func (d *DBStorage) FetchUserByEmail(email string) (*models.User, error) {
	user, err := d.queries.FetchUserByEmail(*d.ctx, email)
	if err != nil {
		return nil, err
	}
	return &models.User{
		ID:         user.ID.String(),
		Name:       user.Name,
		Email:      user.Email,
		IsVerified: user.IsVerified,
		Password:   user.Password,
	}, nil
}

func (d *DBStorage) CreateUser(u models.User) (*models.User, error) {
	userUUID, err := uuid.Parse(u.ID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	user, err := d.queries.InsertUser(*d.ctx, db.InsertUserParams{
		ID:         userUUID,
		Name:       u.Name,
		Email:      u.Email,
		IsVerified: u.IsVerified,
		Password:   u.Password,
		CreatedAt:  sql.NullTime{Time: now, Valid: true},
		UpdatedAt:  sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &models.User{
		ID:         user.ID.String(),
		Name:       user.Name,
		Email:      user.Email,
		IsVerified: user.IsVerified,
		Password:   user.Password,
	}, nil
}

func (d *DBStorage) UpdateUser(u models.User) (*models.User, error) {
	id, _ := uuid.Parse(u.ID)
	user, err := d.queries.UpdateUser(*d.ctx, db.UpdateUserParams{
		ID:         id,
		Name:       u.Name,
		Email:      u.Email,
		IsVerified: u.IsVerified,
		Password:   u.Password,
	})
	if err != nil {
		return nil, err
	}
	return &models.User{
		ID:         user.ID.String(),
		Name:       user.Name,
		Email:      user.Email,
		IsVerified: user.IsVerified,
		Password:   user.Password,
	}, nil
}

func (d *DBStorage) FetchGroupsByUser(userId string) ([]models.Group, error) {
	uid, _ := uuid.Parse(userId)
	groups, err := d.queries.FetchGroupsByUser(*d.ctx, uid)
	if err != nil {
		return nil, err
	}
	var result []models.Group
	for _, g := range groups {
		result = append(result, models.Group{
			Id:          g.ID.String(),
			Name:        g.Name,
			Description: g.Description,
			Admin:       g.AdminID.String(),
		})
	}
	return result, nil
}

func (d *DBStorage) FetchUserById(id string) (*models.User, error) {
	uid, _ := uuid.Parse(id)
	user, err := d.queries.FetchUserById(*d.ctx, uid)
	if err != nil {
		return nil, err
	}
	return &models.User{
		ID:         user.ID.String(),
		Name:       user.Name,
		Email:      user.Email,
		IsVerified: user.IsVerified,
		Password:   user.Password,
	}, nil
}

func (d *DBStorage) FetchUsersInGroup(groupId string) ([]models.User, error) {
	gid, _ := uuid.Parse(groupId)
	users, err := d.queries.FetchUsersInGroup(*d.ctx, gid)
	if err != nil {
		return nil, err
	}
	var result []models.User
	for _, u := range users {
		result = append(result, models.User{
			ID:         u.ID.String(),
			Name:       u.Name,
			Email:      u.Email,
			IsVerified: u.IsVerified,
			Password:   u.Password,
		})
	}
	return result, nil
}

func (d *DBStorage) FetchGroupById(id string) (*models.Group, error) {
	gid, _ := uuid.Parse(id)
	group, err := d.queries.FetchGroupById(*d.ctx, gid)
	if err != nil {
		return nil, err
	}
	return &models.Group{
		Id:          group.ID.String(),
		Name:        group.Name,
		Description: group.Description,
		Admin:       group.AdminID.String(),
	}, nil
}

func (d *DBStorage) FetchGroupExpenses(groupId string, pageNumber int) ([]models.GroupExpenseHistory, error) {
	gid, _ := uuid.Parse(groupId)
	rows, err := d.queries.FetchGroupExpenses(*d.ctx, db.FetchGroupExpensesParams{
		GroupID: uuid.NullUUID{UUID: gid, Valid: true},
		Column2: pageNumber,
	})
	if err != nil {
		return nil, err
	}

	// Calculate total pages
	var totalCount int
	err = d.db.QueryRowContext(*d.ctx, "SELECT COUNT(DISTINCT e.id) FROM expense e JOIN expense_mapping em ON e.id = em.expense_id WHERE em.group_id = $1", gid).Scan(&totalCount)
	if err != nil {
		return nil, err
	}
	pageSize := 20
	totalPages := (totalCount + pageSize - 1) / pageSize

	var result []models.GroupExpenseHistory
	for _, row := range rows {
		var participants []string
		_ = json.Unmarshal(row.ParticipantIds, &participants)
		amount, _ := strconv.ParseFloat(row.Amount, 64)
		var splitW models.SplitWrapper
		_ = json.Unmarshal(row.Split, &splitW)
		var payeeW models.PayerWrapper
		_ = json.Unmarshal(row.Payee, &payeeW)
		result = append(result, models.GroupExpenseHistory{
			Expenses: []models.Expense{
				{
					ID:             row.ID.String(),
					Description:    row.Description.String,
					Amount:         amount,
					Status:         models.ExpenseStatus(row.Status),
					CreatedBy:      row.CreatedBy.UUID.String(),
					SettledBy:      row.SettledBy.UUID.String(),
					CreatedAt:      row.CreatedAt.Time,
					Split:          splitW.Split,
					Payee:          payeeW.Payer,
					IsGroupExpense: true,
					GroupId:        groupId,
				},
			},
			PageNumber: pageNumber,
			TotalPages: totalPages,
		})
	}
	return result, nil
}

func (d *DBStorage) CreateOrUpdateGroup(group models.Group) (*models.Group, error) {
	id, _ := uuid.Parse(group.Id)
	admin, _ := uuid.Parse(group.Admin)
	g, err := d.queries.CreateOrUpdateGroup(*d.ctx, db.CreateOrUpdateGroupParams{
		ID:          id,
		Name:        group.Name,
		Description: group.Description,
		AdminID:     admin,
	})
	if err != nil {
		return nil, err
	}
	return &models.Group{
		Id:          g.ID.String(),
		Name:        g.Name,
		Description: g.Description,
		Admin:       g.AdminID.String(),
	}, nil
}

func (d *DBStorage) AddUserInGroup(userId string, groupId string) (bool, error) {
	uid, _ := uuid.Parse(userId)
	gid, _ := uuid.Parse(groupId)
	return d.queries.AddUserInGroup(*d.ctx, db.AddUserInGroupParams{
		UserID:  uid,
		GroupID: gid,
	})
}

func (d *DBStorage) RemoveUserFromGroup(userId string, groupId string) (bool, error) {
	uid, _ := uuid.Parse(userId)
	gid, _ := uuid.Parse(groupId)
	return d.queries.RemoveUserFromGroup(*d.ctx, db.RemoveUserFromGroupParams{
		UserID:  uid,
		GroupID: gid,
	})
}

func (d *DBStorage) CreateOrUpdateExpense(expense models.ExpenseData) (*models.ExpenseData, error) {
	createdBy, _ := uuid.Parse(expense.CreatedBy)
	settledBy, _ := uuid.Parse(expense.SettledBy)
	now := time.Now()
	amountStr := fmt.Sprintf("%f", expense.Amount)
	var createdAt time.Time
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	parsed, err := uuid.Parse(expense.ID)
	if err != nil {
		return nil, err
	}
	e, err := d.queries.CreateOrUpdateExpense(*d.ctx, db.CreateOrUpdateExpenseParams{
		ID:          parsed,
		Description: sql.NullString{String: expense.Description, Valid: true},
		Amount:      amountStr,
		Split:       json.RawMessage(expense.Split),
		Status:      string(expense.Status),
		SettledBy:   uuid.NullUUID{UUID: settledBy, Valid: expense.SettledBy != ""},
		CreatedBy:   uuid.NullUUID{UUID: createdBy, Valid: expense.CreatedBy != ""},
		Payee:       json.RawMessage(expense.Payee),
		CreatedAt:   sql.NullTime{Time: createdAt, Valid: true},
		UpdatedAt:   sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	amount, _ := strconv.ParseFloat(e.Amount, 64)
	payeeW, err := e.Payee.MarshalJSON()
	if err != nil {
		return nil, err
	}
	splitW, err := e.Split.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return &models.ExpenseData{
		ID:          e.ID.String(),
		Description: e.Description.String,
		Amount:      amount,
		Status:      models.ExpenseStatus(e.Status),
		CreatedBy:   e.CreatedBy.UUID.String(),
		SettledBy:   e.SettledBy.UUID.String(),
		CreatedAt:   e.CreatedAt.Time,
		Payee:       string(payeeW),
		Split:       string(splitW),
	}, nil
}

func (d *DBStorage) FetchExpense(id string) (*models.ExpenseData, error) {
	expenseUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	e, err := d.queries.FetchExpense(*d.ctx, expenseUUID)
	if err != nil {
		return nil, err
	}
	amount, _ := strconv.ParseFloat(e.Amount, 64)
	return &models.ExpenseData{
		ID:          e.ID.String(),
		Description: e.Description.String,
		Amount:      amount,
		Status:      models.ExpenseStatus(e.Status),
		CreatedBy:   e.CreatedBy.UUID.String(),
		SettledBy:   e.SettledBy.UUID.String(),
		CreatedAt:   e.CreatedAt.Time,
		Payee:       string(e.Payee),
		Split:       string(e.Split),
	}, nil
}

func (d *DBStorage) CheckUserExistsInGroup(userId string, groupId string) (bool, error) {
	uid, _ := uuid.Parse(userId)
	gid, _ := uuid.Parse(groupId)
	return d.queries.CheckUserExistsInGroup(*d.ctx, db.CheckUserExistsInGroupParams{
		UserID:  uid,
		GroupID: gid,
	})
}

func (d *DBStorage) AddExpenseMapping(expenseId string, userId string) (bool, error) {
	uid, _ := uuid.Parse(userId)
	eid, _ := uuid.Parse(expenseId)
	return d.queries.AddUserExpenseMapping(*d.ctx, db.AddUserExpenseMappingParams{ExpenseID: eid, UserID: uid})
}

func (d *DBStorage) AttachExpenseToGroup(expenseId string, groupId string, users []string) (bool, error) {
	eid, _ := uuid.Parse(expenseId)
	gid, _ := uuid.Parse(groupId)
	var userUUIDs []uuid.UUID
	for _, u := range users {
		uid, _ := uuid.Parse(u)
		userUUIDs = append(userUUIDs, uid)
	}
	return d.queries.AttachExpenseToGroup(*d.ctx, db.AttachExpenseToGroupParams{
		ExpenseID: eid,
		GroupID:   uuid.NullUUID{UUID: gid, Valid: true},
		Column3:   userUUIDs,
	})
}

func (d *DBStorage) RemoveUserFromExpense(expenseId string, usersToRemove []string) (bool, error) {
	eid, _ := uuid.Parse(expenseId)
	var userUUIDs []uuid.UUID
	for _, u := range usersToRemove {
		uid, _ := uuid.Parse(u)
		userUUIDs = append(userUUIDs, uid)
	}
	return d.queries.RemoveUserFromExpense(*d.ctx, db.RemoveUserFromExpenseParams{
		ExpenseID: eid,
		Column2:   userUUIDs,
	})
}

func (d *DBStorage) DeleteExpense(id string) (bool, error) {
	expenseId, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}
	return d.queries.DeleteExpense(*d.ctx, expenseId)
}

func (d *DBStorage) GetFriend(userId string, friendId string) (*expense.User, error) {
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return nil, err
	}
	friendUUID, err := uuid.Parse(friendId)
	if err != nil {
		return nil, err
	}
	row, err := d.queries.GetFriend(*d.ctx, db.GetFriendParams{UserID: userUUID, FriendID: friendUUID})
	if err != nil {
		return nil, err
	}
	return &expense.User{Name: row.Name, Email: row.Email, ID: row.ID.String()}, nil

}

func (d *DBStorage) GetFriends(userId string) ([]expense.User, error) {
	parsed, err := uuid.Parse(userId)
	if err != nil {
		return []expense.User{}, err
	}
	rows, err := d.queries.GetFriends(*d.ctx, parsed)
	if err != nil {
		return []expense.User{}, err
	}
	result := []expense.User{}
	for _, row := range rows {
		result = append(result, expense.User{Name: row.Name, Email: row.Email, ID: row.ID.String()})
	}

	return result, nil
}

func (d *DBStorage) RemoveFriend(userId string, friendId string) (bool, error) {
	parsedUserID, err := uuid.Parse(userId)
	if err != nil {
		return false, err
	}
	friendUUID, err := uuid.Parse(friendId)
	if err != nil {
		return false, err
	}
	return d.queries.RemoveFriend(*d.ctx, db.RemoveFriendParams{UserID: parsedUserID, FriendID: friendUUID})
}

func (d *DBStorage) AddFriend(userId string, friendId string) (bool, error) {
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return false, err
	}
	friendUUID, err := uuid.Parse(friendId)
	if err != nil {
		return false, err
	}
	_, err = d.queries.AddFriend(*d.ctx, db.AddFriendParams{UserID: userUUID, FriendID: friendUUID})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *DBStorage) FetchExpenseAssociatedGroup(expenseId string) (bool, string, error) {
	expenseUUID, err := uuid.Parse(expenseId)
	if err != nil {
		return false, "", err
	}
	gID, err := d.queries.FetchExpenseAssociatedGroup(*d.ctx, expenseUUID)
	if err != nil && err != sql.ErrNoRows {
		return false, "", err
	}
	return err != sql.ErrNoRows, gID.UUID.String(), nil
}
