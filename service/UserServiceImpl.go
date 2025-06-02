package service

import (
	"crypto"
	"encoding/base64"
	"errors"
	"splitExpense/config"
	expense "splitExpense/expense"
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
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user Already Exists")
	}

	hasher := crypto.SHA256.New()
	hasher.Write([]byte(password))
	passwordHash := hasher.Sum(nil)

	user, err := u.storage.CreateUser(expense.UserCreate{
		Name:     name,
		Password: base64.StdEncoding.EncodeToString(passwordHash),
		Email:    email,
	})

	return user, err
}

func (u *UserServiceImpl) JoinGroup(userId string, groupId string) (bool, error) {
	ok, err := u.storage.AddUserInGroup(userId, groupId)
	return ok, err
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
		if g.Name == name {
			return nil, errors.New("group with same name already exists")
		}
	}

	return u.storage.CreateOrUpdateGroup(expense.Group{
		Name:        name,
		Description: description,
		Admin:       userId,
	})
}

func (u *UserServiceImpl) GetAssociatedGroups(userId string) ([]expense.Group, error) {
	return u.storage.FetchGroupsByUser(userId)
}
