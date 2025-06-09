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

	FetchUsersInGroup(groupId string) ([]User, error)
	FetchGroupById(id string) (*Group, error)
	FetchGroupExpenses(groupId string, pageNumber int) ([]GroupExpenseHistory, error)
	CreateOrUpdateGroup(group Group) (*Group, error)
	AddUserInGroup(userId string, groupId string) (bool, error)
	RemoveUserFromGroup(userId string, groupId string) (bool, error)

	CreateOrUpdateExpense(expense ExpenseData) (*ExpenseData, error)
	FetchExpense(id string) (*ExpenseData, error)
	FetchExpenseAssociatedGroup(expenseId string) (ok bool, groupId string, err error)
	CheckUserExistsInGroup(userId string, groupId string) (bool, error)
	AttachExpenseToGroup(expenseId string, groupId string, users []string) (bool, error)
	RemoveUserFromExpense(expenseId string, usersToRemove []string) (bool, error)
	DeleteExpense(id string) (bool, error)
}
