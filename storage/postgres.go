package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"splitExpense/config"
	"splitExpense/db"
	"splitExpense/expense"
	models "splitExpense/expense"

	"strconv"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"

	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
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
	url := os.Getenv("DATABASE_URL")
	fmt.Println("DATABASE_URL: ", url)
	if url == "" {
		fmt.Println("using local postgres connection....")
		url = dsn
	}
	fmt.Println("Connecting to Postgres DB with URL: ", url)
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatal("failed to establish db connection: ", err)
		return nil, err
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout})

	loggerAdapter := zerologadapter.New(logger)
	db = sqldblogger.OpenDriver(url, db.Driver(), loggerAdapter /*, using_default_options*/) // db is STILL *sql.DB
	return db, err
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
		log.Fatal("error creating postgres connectiong", err)
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

func (d *DBStorage) FetchGroupMembers(groupId string) ([]models.User, error) {
	gid, _ := uuid.Parse(groupId)
	users, err := d.queries.FetchGroupMembers(*d.ctx, gid)
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

func (d *DBStorage) FetchGroupExpenses(groupId string, pageNumber int) (*models.StoredGroupExpenseHistory, error) {
	gid, _ := uuid.Parse(groupId)
	rows, err := d.queries.FetchGroupExpenses(*d.ctx, db.FetchGroupExpensesParams{
		GroupID: uuid.NullUUID{UUID: gid, Valid: true},
		Column2: pageNumber,
		Limit:   20,
	})

	if err != nil {
		return nil, err
	}

	// Calculate total pages
	var totalCount int
	pageSize := 20
	totalPages := (totalCount + pageSize - 1) / pageSize

	return d.GetStoredGroupExpenseFromRows(rows, pageNumber, totalPages)

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
	_, err := d.queries.AddUserInGroup(*d.ctx, db.AddUserInGroupParams{
		UserID:  uid,
		GroupID: gid,
	})
	if err == nil || err == sql.ErrNoRows {
		return true, nil
	} else {
		return false, err
	}
}

func (d *DBStorage) RemoveUserFromGroup(userId string, groupId string) (bool, error) {
	uid, _ := uuid.Parse(userId)
	gid, _ := uuid.Parse(groupId)
	_, err := d.queries.RemoveUserFromGroup(*d.ctx, db.RemoveUserFromGroupParams{
		UserID:  uid,
		GroupID: gid,
	})
	if err == nil || err == sql.ErrNoRows {
		return true, nil
	}
	return false, err
}

func (d *DBStorage) CreateOrUpdateExpense(expense models.Expense) (*models.Expense, error) {

	var err error

	var settledBy uuid.UUID
	createdBy, err := uuid.Parse(expense.CreatedBy)
	if expense.SettledBy != "" {
		settledBy, err = uuid.Parse(expense.SettledBy)
	}
	if err != nil {
		return nil, err
	}

	groupId := uuid.NullUUID{}
	if expense.IsGroupExpense {
		groupId_, err := uuid.Parse(expense.GroupId)
		if err != nil {
			return nil, err
		}

		groupId = uuid.NullUUID{UUID: groupId_, Valid: true}
	}

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

	splitJson, err := json.Marshal(expense.SplitW)
	if err != nil {
		return nil, err
	}

	payeeJson, err := json.Marshal(expense.PayeeW)
	if err != nil {
		return nil, err
	}

	e, err := d.queries.CreateOrUpdateExpense(*d.ctx, db.CreateOrUpdateExpenseParams{
		ID:          parsed,
		Description: sql.NullString{String: expense.Description, Valid: true},
		Amount:      amountStr,
		Split:       json.RawMessage(splitJson),
		Status:      string(expense.Status),
		SettledBy:   uuid.NullUUID{UUID: settledBy, Valid: expense.SettledBy != ""},
		CreatedBy:   createdBy,
		Payee:       json.RawMessage(payeeJson),
		CreatedAt:   sql.NullTime{Time: createdAt, Valid: true},
		UpdatedAt:   sql.NullTime{Time: now, Valid: true},
		GroupID:     groupId,
	})
	if err != nil {
		return nil, err
	}
	amount, _ := strconv.ParseFloat(e.Amount, 64)

	var payeeW models.PayerWrapper
	err = json.Unmarshal(e.Payee, &payeeW)
	if err != nil {
		return nil, err
	}

	var splitW models.SplitWrapper
	err = json.Unmarshal(e.Split, &splitW)
	if err != nil {
		return nil, err
	}

	var settledBy_ string
	if e.SettledBy.Valid {
		settledBy_ = e.SettledBy.UUID.String()
	}

	return &models.Expense{
		ID:             e.ID.String(),
		Description:    e.Description.String,
		Amount:         amount,
		Status:         models.ExpenseStatus(e.Status),
		CreatedBy:      e.CreatedBy.String(),
		SettledBy:      settledBy_,
		CreatedAt:      e.CreatedAt.Time,
		PayeeW:         payeeW,
		SplitW:         splitW,
		IsGroupExpense: e.GroupID.Valid,
		GroupId:        e.GroupID.UUID.String(),
	}, nil
}

func (d *DBStorage) FetchExpense(id string) (*models.Expense, error) {
	expenseUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	e, err := d.queries.FetchExpense(*d.ctx, expenseUUID)
	if err != nil {
		return nil, err
	}

	var payeeW models.PayerWrapper
	err = json.Unmarshal(e.Payee, &payeeW)
	if err != nil {
		return nil, err
	}

	var splitW models.SplitWrapper
	err = json.Unmarshal(e.Split, &splitW)
	if err != nil {
		return nil, err
	}
	amount, _ := strconv.ParseFloat(e.Amount, 64)

	return &models.Expense{
		ID:             e.ID.String(),
		Description:    e.Description.String,
		Amount:         amount,
		Status:         models.ExpenseStatus(e.Status),
		CreatedBy:      e.CreatedBy.String(),
		SettledBy:      e.SettledBy.UUID.String(),
		CreatedAt:      e.CreatedAt.Time,
		PayeeW:         payeeW,
		SplitW:         splitW,
		IsGroupExpense: e.GroupID.Valid,
		GroupId:        e.GroupID.UUID.String(),
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

func (d *DBStorage) RemoveUsersFromExpense(expenseId string, usersToRemove []string) (bool, error) {
	eid, _ := uuid.Parse(expenseId)
	var userUUIDs []uuid.UUID
	for _, u := range usersToRemove {
		uid, _ := uuid.Parse(u)
		userUUIDs = append(userUUIDs, uid)
	}
	return d.queries.RemoveUsersFromExpenseMapping(*d.ctx, db.RemoveUsersFromExpenseMappingParams{
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

func (d *DBStorage) GetFriends(userId string) ([]models.User, error) {
	uid, err := uuid.Parse(userId)
	if err != nil {
		return nil, err
	}
	friends, err := d.queries.GetFriends(*d.ctx, uid)
	if err != nil {
		return nil, err
	}
	var result []models.User
	for _, u := range friends {
		result = append(result, models.User{
			ID:         u.ID.String(),
			Name:       u.Name,
			Email:      u.Email,
			IsVerified: u.IsVerified,
			Password:   "",
		})
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

func (d *DBStorage) FetchExpenseCountByGroup(groupId string) (int, error) {
	gid, _ := uuid.Parse(groupId)
	count, err := d.queries.FetchExpenseCountByGroup(*d.ctx, uuid.NullUUID{UUID: gid, Valid: true})
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return int(count), nil
}

func (d *DBStorage) FetchExpenseByUserAndStatus(userId string, status models.ExpenseStatus, pageNumber int, limit int32) (*models.StoredGroupExpenseHistory, error) {
	if pageNumber == 0 {
		pageNumber = 1
	}

	uid, _ := uuid.Parse(userId)
	rows, err := d.queries.FetchExpenseByUserAndStatus(*d.ctx, db.FetchExpenseByUserAndStatusParams{
		UserID:  uid,
		Status:  string(status),
		Column4: pageNumber,
		Limit:   limit,
	})
	if err != nil {
		return nil, err
	}

	totalCount, err := d.queries.FetchExpenseCountByUserAndStatus(*d.ctx, db.FetchExpenseCountByUserAndStatusParams{
		UserID: uid,
		Status: string(status),
	})
	totalPages := (int(totalCount) + int(limit) - 1) / int(limit)
	if err != nil {
		return nil, err
	}
	return d.GetStoredGroupExpenseFromRows(rows, pageNumber, int(totalPages))
}

func (d *DBStorage) GetStoredGroupExpenseFromRows(rows []db.Expense, pageNumber int, totalPages int) (*models.StoredGroupExpenseHistory, error) {
	result := models.StoredGroupExpenseHistory{Expenses: []models.Expense{}, PageNumber: pageNumber, TotalPages: totalPages}
	for _, row := range rows {
		amount, _ := strconv.ParseFloat(row.Amount, 64)
		var splitW models.SplitWrapper
		_ = json.Unmarshal(row.Split, &splitW)
		var payeeW models.PayerWrapper
		_ = json.Unmarshal(row.Payee, &payeeW)
		result.Expenses = append(result.Expenses, models.Expense{
			ID:             row.ID.String(),
			Description:    row.Description.String,
			Amount:         amount,
			Status:         models.ExpenseStatus(row.Status),
			CreatedBy:      row.CreatedBy.String(),
			SettledBy:      row.SettledBy.UUID.String(),
			CreatedAt:      row.CreatedAt.Time,
			SplitW:         splitW,
			PayeeW:         payeeW,
			IsGroupExpense: row.GroupID.Valid,
			GroupId:        row.GroupID.UUID.String(),
		})
	}
	return &result, nil
}

func (d *DBStorage) FetchGroupExpensesByStatus(groupId string, status models.ExpenseStatus, pageNumber int) (*models.StoredGroupExpenseHistory, error) {
	if pageNumber == 0 {
		pageNumber = 1
	}
	limit := 20

	gid_, _ := uuid.Parse(groupId)
	gid := uuid.NullUUID{UUID: gid_, Valid: true}

	totalCount, err := d.queries.FetchExpenseCountByGroupAndStatus(*d.ctx, db.FetchExpenseCountByGroupAndStatusParams{
		GroupID: gid,
		Status:  string(status)},
	)
	totalPages := (int(totalCount) + limit - 1) / limit // Assuming page size is 20
	if err != nil {
		return nil, err
	}

	rows, err := d.queries.FetchGroupExpensesByStatus(*d.ctx, db.FetchGroupExpensesByStatusParams{
		GroupID: gid,
		Status:  string(status),
		Column3: pageNumber,
		Limit:   int32(limit),
	})
	if err != nil {
		return nil, err
	}

	return d.GetStoredGroupExpenseFromRows(rows, pageNumber, int(totalPages))

}

func (d *DBStorage) DeleteGroup(groupId string) (bool, error) {
	gid, err := uuid.Parse(groupId)
	if err != nil {
		return false, err
	}
	deleted, err := d.queries.DeleteGroup(*d.ctx, gid)
	if err == nil || err == sql.ErrNoRows {
		return deleted, nil
	}
	return false, nil
}
