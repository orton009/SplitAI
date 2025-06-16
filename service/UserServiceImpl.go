package service

import (
	"crypto"
	"database/sql"
	"encoding/base64"
	"errors"
	"splitExpense/config"
	expense "splitExpense/expense"
	"strings"

	"github.com/google/uuid"
)

type UserServiceImpl struct {
	config  *config.Config
	storage expense.Storage
}

func (u *UserServiceImpl) GetUser(id string) (*expense.User, error) {
	user, err := u.storage.FetchUserById(id)
	return user, err
}

func (u *UserServiceImpl) CreateUser(name string, email string, password string) (*expense.User, error) {

	existingUser, err := u.storage.FetchUserByEmail(email)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user Already Exists")
	}

	hasher := crypto.SHA256.New()
	hasher.Write([]byte(password))
	passwordHash := hasher.Sum(nil)

	user, err := u.storage.CreateUser(expense.User{
		ID:         uuid.New().String(),
		Name:       name,
		Password:   base64.StdEncoding.EncodeToString(passwordHash),
		Email:      email,
		IsVerified: false,
	})
	user.Password = ""

	return user, err
}

func (u *UserServiceImpl) JoinGroup(userId string, groupId string) (bool, error) {
	return u.storage.AddUserInGroup(userId, groupId)
}

func (u *UserServiceImpl) LeaveGroup(userId string, groupId string) (bool, error) {
	doesExist, err := u.storage.CheckUserExistsInGroup(userId, groupId)
	if err != nil {
		return false, err
	}
	if !doesExist {
		return false, errors.New("user does not exist in group")
	}

	return u.storage.RemoveUserFromGroup(userId, groupId)
}

func (u *UserServiceImpl) CreateGroup(userId string, name string, description string) (*expense.Group, error) {
	groups, err := u.storage.FetchGroupsByUser(userId)
	if err != nil {
		return nil, err
	}

	for _, g := range groups {
		if strings.EqualFold(g.Name, name) {
			return nil, errors.New("group with same name already exists")
		}
	}
	group := expense.Group{
		Id:          uuid.New().String(),
		Name:        name,
		Description: description,
		Admin:       userId,
	}
	updatedGroup, err := u.storage.CreateOrUpdateGroup(group)
	if err != nil {
		return nil, err
	}

	_, err = u.storage.AddUserInGroup(userId, updatedGroup.Id)
	if err != nil {
		return nil, err
	}

	return updatedGroup, nil
}

func (u *UserServiceImpl) GetAssociatedGroups(userId string) ([]expense.Group, error) {
	return u.storage.FetchGroupsByUser(userId)
}

func (u *UserServiceImpl) FetchUserCredentials(email string) (*expense.User, error) {
	return u.storage.FetchUserByEmail(email)
}

func NewUserServiceImpl(cfg *config.Config, storage expense.Storage) *UserServiceImpl {
	return &UserServiceImpl{
		config:  cfg,
		storage: storage,
	}
}

type AssociatedUsers struct {
	Group expense.Group
	Users []expense.User
}

func (u *UserServiceImpl) GetGroupById(groupId string) (*expense.Group, error) {
	group, err := u.storage.FetchGroupById(groupId)
	if err != nil {

		return nil, err
	}

	return group, nil
}

func (u *UserServiceImpl) GetAssociatedUsers(groupId string) (*AssociatedUsers, error) {
	group, err := u.storage.FetchGroupById(groupId)
	if err != nil {
		return nil, err
	}

	users, err := u.storage.FetchUsersInGroup(groupId)
	if err != nil {
		return nil, err
	}
	return &AssociatedUsers{
		Group: *group,
		Users: users,
	}, nil
}

// AddFriend implements the UserService interface
func (us *UserServiceImpl) AddFriend(userId string, friendId string) (bool, error) {

	return us.storage.AddFriend(userId, friendId)
}

// GetFriends implements the UserService interface
func (us *UserServiceImpl) GetFriends(userId string) ([]expense.User, error) {

	return us.storage.GetFriends(userId)
}

// GetFriend implements the UserService interface
func (us *UserServiceImpl) GetFriend(userId string, friendId string) (*expense.User, error) {

	return us.storage.GetFriend(userId, friendId)
}
