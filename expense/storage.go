package expense

type UserHome struct {
	ExpenseWithUsers []Expense
	AssociatedGroups []Group
	User             User
}

type UserCreate struct {
	Name     string
	Email    string
	Password string
}

type Storage interface {
	FetchUserByEmail(email string) (*User, error)
	CreateUser(User) (*User, error)
	UpdateUser(User) (*User, error)
	FetchGroupsByUser(userId string) ([]Group, error)
	FetchUserById(id string) (*User, error)

	AddFriend(userId string, friendId string) (bool, error)
	GetFriend(userId string, friendId string) (*User, error)
	GetFriends(userId string) ([]User, error)
	RemoveFriend(userId string, friendId string) (bool, error)

	DeleteGroup(groupId string) (bool, error)

	FetchGroupMembers(groupId string) ([]User, error)
	FetchGroupById(id string) (*Group, error)
	FetchGroupExpenses(groupId string, pageNumber int) (*StoredGroupExpenseHistory, error)
	CreateOrUpdateGroup(group Group) (*Group, error)
	AddUserInGroup(userId string, groupId string) (bool, error)
	RemoveUserFromGroup(userId string, groupId string) (bool, error)

	AddExpenseMapping(expenseId string, userId string) (bool, error)
	FetchExpenseCountByGroup(groupId string) (int, error)
	CreateOrUpdateExpense(expense Expense) (*Expense, error)
	FetchExpense(id string) (*Expense, error)
	CheckUserExistsInGroup(userId string, groupId string) (bool, error)
	RemoveUsersFromExpense(expenseId string, usersToRemove []string) (bool, error)
	DeleteExpense(id string) (bool, error)

	FetchExpenseByUserAndStatus(userId string, status ExpenseStatus, pageNumber int, limit int32) (*StoredGroupExpenseHistory, error)
	FetchGroupExpensesByStatus(groupId string, status ExpenseStatus, pageNumber int) (*StoredGroupExpenseHistory, error)
}
