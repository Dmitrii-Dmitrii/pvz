package custom_errors

import "fmt"

type UserError struct {
	Err     error
	Message string
}

func (e *UserError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *UserError) Unwrap() error {
	return e.Err
}

var (
	ErrNoOpenReception = &UserError{Message: "no open reception"}
	ErrNoReception     = &UserError{Message: "no reception"}
	ErrDateRange       = &UserError{Message: "end date cannot be before start date"}
	ErrLimitValue      = &UserError{Message: "limit must be between 1 and 30"}
	ErrPageValue       = &UserError{Message: "page must be greater than zero"}
	ErrUuidFormat      = &UserError{Message: "invalid UUID format"}
)
