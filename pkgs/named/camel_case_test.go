package named

import (
	"testing"

	"github.com/golib/assert"
)

func Test_ToCamelCase(t *testing.T) {
	it := assert.New(t)

	testCases := []struct {
		in       string
		expected string
	}{
		{
			in:       "",
			expected: "",
		},
		{
			in:       "lowercase",
			expected: "Lowercase",
		},
		{
			in:       "Class",
			expected: "Class",
		},
		{
			in:       "MyClass",
			expected: "MyClass",
		},
		{
			in:       "MyC",
			expected: "MyC",
		},
		{
			in:       "HTML",
			expected: "HTML",
		},
		{
			in:       "PDFLoader",
			expected: "PDFLoader",
		},
		{
			in:       "AString",
			expected: "AString",
		},
		{
			in:       "SimpleXMLParser",
			expected: "SimpleXMLParser",
		},
		{
			in:       "vimRPCPlugin",
			expected: "VimRPCPlugin",
		},
		{
			in:       "GO111Version",
			expected: "GO111Version",
		},
		{
			in:       "99Bottles",
			expected: "99Bottles",
		},
		{
			in:       "May5",
			expected: "May5",
		},
		{
			in:       "BFG9000",
			expected: "BFG9000",
		},
		{
			in:       "BöseÜberraschung",
			expected: "BöseÜberraschung",
		},
		{
			in:       "Two  spaces",
			expected: "TwoSpaces",
		},
		{
			in:       "snake_case",
			expected: "SnakeCase",
		},
		{
			in:       "_snakeCase",
			expected: "SnakeCase",
		},
		{
			in:       "BadUTF8\xe2\xe2\xa1",
			expected: "BadUTF8\xe2\xe2\xa1",
		},
	}

	for i, testCase := range testCases {
		it.Equal(testCase.expected, ToCamelCase(testCase.in), "Case(#%d)", i)
	}
}
