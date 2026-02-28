package xhttp

import (
	"errors"
	"net/http"
	"testing"

	"github.com/raymondwongso/goerp/domain/xerror"
	"github.com/stretchr/testify/assert"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "CodeConflict",
			err:      xerror.New(xerror.CodeConflict, "conflict"),
			expected: http.StatusConflict,
		},
		{
			name:     "CodeDuplicate",
			err:      xerror.New(xerror.CodeDuplicate, "duplicate"),
			expected: http.StatusConflict,
		},
		{
			name:     "CodeForbidden",
			err:      xerror.New(xerror.CodeForbidden, "forbidden"),
			expected: http.StatusForbidden,
		},
		{
			name:     "CodeInternal",
			err:      xerror.New(xerror.CodeInternal, "internal"),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "CodeInvalidParameter",
			err:      xerror.New(xerror.CodeInvalidParameter, "invalid parameter"),
			expected: http.StatusBadRequest,
		},
		{
			name:     "CodeNotFound",
			err:      xerror.New(xerror.CodeNotFound, "not found"),
			expected: http.StatusNotFound,
		},
		{
			name:     "CodeNotImplemented",
			err:      xerror.New(xerror.CodeNotImplemented, "not implemented"),
			expected: http.StatusNotImplemented,
		},
		{
			name:     "CodeServiceUnavailable",
			err:      xerror.New(xerror.CodeServiceUnavailable, "service unavailable"),
			expected: http.StatusServiceUnavailable,
		},
		{
			name:     "CodeTimeout",
			err:      xerror.New(xerror.CodeTimeout, "timeout"),
			expected: http.StatusRequestTimeout,
		},
		{
			name:     "CodeUnauthorized",
			err:      xerror.New(xerror.CodeUnauthorized, "unauthorized"),
			expected: http.StatusUnauthorized,
		},
		{
			name:     "CodeUnprocessable",
			err:      xerror.New(xerror.CodeUnprocessable, "unprocessable"),
			expected: http.StatusUnprocessableEntity,
		},
		{
			name:     "CodeUnknown returns 0",
			err:      xerror.New(xerror.CodeUnknown, "unknown"),
			expected: 0,
		},
		{
			name:     "non-xerror returns 0",
			err:      errors.New("some generic error"),
			expected: 0,
		},
		{
			name:     "wrapped xerror preserves code",
			err:      xerror.NewWithCause(xerror.CodeNotFound, "not found", errors.New("cause")),
			expected: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, MapError(tc.err))
		})
	}
}
