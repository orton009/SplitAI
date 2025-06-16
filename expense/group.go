package expense

type ExpenseSummary struct {
	// list of users, for each user -> list of all users he has borrowed the amount from
	BorrowedFrom map[int](map[int]float64)
	// list of users, for each user -> list of all users who have borrowed amount from him
	Owed map[int](map[int]float64)
}

type DetailedExpense struct {
	Expense       Expense `json:"expense"`
	TotalOwed     float64 `json:"totalOwed"`
	TotalBorrowed float64 `json:"totalBorrowed"`
}

type StoredGroupExpenseHistory struct {
	Expenses   []Expense `json:"expenses"`
	PageNumber int       `json:"pageNumber"`
	TotalPages int       `json:"totalPages"`
}

type GroupExpenseHistory struct {
	Expenses   []DetailedExpense `json:"expenses"`
	PageNumber int               `json:"pageNumber"`
	TotalPages int               `json:"totalPages"`
}

type Group struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Admin       string `json:"admin"`
}

func (g *Group) getExpenseSummary() ExpenseSummary {
	// TODO:
	return ExpenseSummary{}
}
