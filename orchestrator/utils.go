package orchestrator

import (
	"regexp"
	"splitExpense/expense"

	"github.com/google/uuid"
)

type validator struct {
	errors [](*expense.AppError)
}

func NewValidator() *validator {
	return &validator{errors: []*expense.AppError{}}
}

func (v *validator) Ok() bool {
	return len(v.errors) == 0
}

func (v *validator) Err() *expense.AppError {
	if len(v.errors) == 0 {
		return nil
	} else {
		return v.errors[0]
	}
}

func (v *validator) Email(e string) *validator {

	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	ok := re.MatchString(e)
	if !ok {
		v.errors = append(v.errors, expense.ErrValidation("Invalid email format: "+e+". Expected format: localpart@domain.tld"))
	}
	return v
}

func (v *validator) Password(password string) *validator {

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#\$%\^&\*\(\)_\+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	ok := len(password) > 8 && hasLower && hasUpper && hasNumber && hasSpecial

	if !ok {
		v.errors = append(v.errors, expense.ErrValidation("Invalid Password, password should contain atleast one uppercase, one lowercase, one special character and one digit"))
	}
	return v
}

func (v *validator) Name(n string) *validator {
	ok := len(n) >= 2
	if !ok {
		v.errors = append(v.errors, expense.ErrValidation("Invalid Name, name should contain atleast two characters"))
	}
	return v
}

func (v *validator) UUID(id string) (*validator, *uuid.UUID) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		v.errors = append(v.errors, expense.ErrValidation("Invalid UUID"))
		return v, nil
	}
	return v, &parsed
}

func (v *validator) NonEmptyID(id string) *validator {
	ok := len(id) > 0
	if !ok {
		v.errors = append(v.errors, expense.ErrValidation("Emmpty ID"))
	}
	return v
}

func (v *validator) LeastAmount(amount float64) *validator {
	ok := amount >= 1.0
	if !ok {
		v.errors = append(v.errors, expense.ErrValidation("Amount less than 1"))
	}
	return v
}
