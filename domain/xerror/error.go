package xerror

import (
	"errors"
	"fmt"
)

type Code string

const (
	CodeDuplicate     Code = "domain.duplicate.error"
	CodeForbidden     Code = "domain.forbidden.error"
	CodeInternal      Code = "domain.internal.error"
	CodeNotFound      Code = "domain.not_found.error"
	CodeUnknown       Code = "domain.unknown.error"
	CodeUnprocessable Code = "domain.unprocessable.error"
)

type Error struct {
	Code    Code          `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details"`
}

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func New(code Code, message string) Error {
	return Error{
		Code:    code,
		Message: message,
	}
}

func NewWithCause(code Code, message string, cause error) error {
	return fmt.Errorf("%w: %w", New(code, message), cause)
}

func (e Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// AddDetail add detail into error, appending from existing error
func AddDetail(err Error, field, message string) Error {
	if len(err.Details) == 0 {
		err.Details = make([]ErrorDetail, 0)
	}

	err.Details = append(err.Details, ErrorDetail{Field: field, Message: message})
	return err
}

// GetCode return Code from error, return CodeUnknown if there is no xerror.Error in the error chain
func GetCode(err error) Code {
	var target Error
	if errors.As(err, &target) {
		return target.Code
	}

	return CodeUnknown
}
