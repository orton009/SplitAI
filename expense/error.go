package expense

import "fmt"

type AppError struct {
	Type    string
	Message string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func ErrValidation(message string) *AppError {
	return &AppError{Type: "ValidationError", Message: message}
}

func ErrInvalidQuery(message string) *AppError {
	return &AppError{Type: "InvalidQueryError", Message: message}
}

func ErrService(message string) *AppError {
	return &AppError{Type: "ServiceError", Message: message}
}
