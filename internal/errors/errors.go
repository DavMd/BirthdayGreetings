package errors

import (
	"fmt"
)

type CurstomError struct {
	Code    int
	Message string
}

func (e *CurstomError) Error() string {
	return fmt.Sprintf("Code: %d, Message: %s", e.Code, e.Message)
}

func New(code int, message string) error {
	return &CurstomError{
		Code:    code,
		Message: message,
	}
}

func IsCustomError(err error) bool {
	_, ok := err.(*CurstomError)
	return ok
}
