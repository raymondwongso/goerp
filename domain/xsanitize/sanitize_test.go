package xsanitize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeEscapeCharacters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean string is unchanged",
			input:    "/api/users",
			expected: "/api/users",
		},
		{
			name:     "newline is escaped",
			input:    "foo\nbar",
			expected: `foo\nbar`,
		},
		{
			name:     "carriage return is escaped",
			input:    "foo\rbar",
			expected: `foo\rbar`,
		},
		{
			name:     "tab is escaped",
			input:    "foo\tbar",
			expected: `foo\tbar`,
		},
		{
			name:     "mixed escape characters are all escaped",
			input:    "foo\nbar\r\tbaz",
			expected: `foo\nbar\r\tbaz`,
		},
		{
			name:     "empty string is unchanged",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, SanitizeEscapeCharacters(tc.input))
		})
	}
}
