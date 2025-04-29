package expense

type ExpenseSummary struct {
	// list of users, for each user -> list of all users he has borrowed the amount from
	BorrowedFrom map[int](map[int]float64)
	// list of users, for each user -> list of all users who have borrowed amount from him
	Owed map[int](map[int]float64)
}

type Group struct {
	Users          []User
	ExpenseHistory []Expense
}

func (g *Group) getExpenseSummary() ExpenseSummary {
	// TODO:
	return ExpenseSummary{}
}
