package named

import (
	"testing"

	"github.com/golib/assert"
)

func Test_ToSnakeCase(t *testing.T) {
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
			expected: "lowercase",
		},
		{
			in:       "Class",
			expected: "class",
		},
		{
			in:       "MyClass",
			expected: "my_class",
		},
		{
			in:       "MyC",
			expected: "my_c",
		},
		{
			in:       "HTML",
			expected: "html",
		},
		{
			in:       "PDFLoader",
			expected: "pdf_loader",
		},
		{
			in:       "AString",
			expected: "a_string",
		},
		{
			in:       "SimpleXMLParser",
			expected: "simple_xml_parser",
		},
		{
			in:       "vimRPCPlugin",
			expected: "vim_rpc_plugin",
		},
		{
			in:       "GO111Version",
			expected: "go_111_version",
		},
		{
			in:       "99Bottles",
			expected: "99_bottles",
		},
		{
			in:       "May5",
			expected: "may_5",
		},
		{
			in:       "BFG9000",
			expected: "bfg_9000",
		},
		{
			in:       "BöseÜberraschung",
			expected: "böse_überraschung",
		},
		{
			in:       "Two  spaces",
			expected: "two_spaces",
		},
		{
			in:       "snake_case",
			expected: "snake_case",
		},
		{
			in:       "_snakeCase",
			expected: "_snake_case",
		},
		{
			in:       "BadUTF8\xe2\xe2\xa1",
			expected: "BadUTF8\xe2\xe2\xa1",
		},
	}

	for i, testCase := range testCases {
		it.Equal(testCase.expected, ToSnakeCase(testCase.in), "Case(#%d)", i)
	}
}
