package xerror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	err := New("some code", "some message")
	assert.Equal(t, Code("some code"), err.Code)
	assert.Equal(t, "some message", err.Message)
	assert.Len(t, err.Details, 0)
}

func Test_NewWithCause(t *testing.T) {
	err := NewWithCause(CodeNotFound, "some error", errors.New("inner cause"))
	assert.Error(t, err)
	assert.True(t, errors.As(err, &Error{}))
	assert.Equal(t, "[domain.not_found.error] some error: inner cause", err.Error())
}

func Test_AddDetail(t *testing.T) {
	err := New("some code", "some message")

	newErr := AddDetail(err, "name", "name is empty")
	assert.Equal(t, newErr.Details[0].Field, "name")
	assert.Equal(t, newErr.Details[0].Message, "name is empty")

	newErr2 := AddDetail(newErr, "address", "address is too long")
	assert.Equal(t, newErr2.Details[1].Field, "address")
	assert.Equal(t, newErr2.Details[1].Message, "address is too long")
	assert.Len(t, newErr2.Details, 2)
}

func Test_Error(t *testing.T) {
	err := New(CodeUnprocessable, "Some Message 123")
	assert.Equal(t, fmt.Sprintf("[%s] %s", err.Code, err.Message), err.Error())
}

func Test_GetCode(t *testing.T) {
	t.Run("error is xerror.Error", func(t *testing.T) {
		code := GetCode(New("custom_error", "custom message"))
		assert.Equal(t, Code("custom_error"), code)
	})

	t.Run("error is not xerror.Error", func(t *testing.T) {
		code := GetCode(errors.New("some error"))
		assert.Equal(t, CodeUnknown, code)
	})
}
