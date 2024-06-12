package errors

import (
	"fmt"
)

type CustomError struct {
	Code    int
	Message string
}

func (e *CustomError) ErrorWithCode() string {
	return fmt.Sprintf("Code: %d, Message: %s", e.Code, e.Message)
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("Message: %s", e.Message)
}

func New(code int, message string) error {
	return &CustomError{
		Code:    code,
		Message: message,
	}
}
