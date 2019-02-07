package named

import (
	"testing"

	"github.com/golib/assert"
)

func Test_parse(t *testing.T) {
	it := assert.New(t)

	testCases := []struct {
		in       string
		expected []string
	}{
		{
			in:       "",
			expected: []string{""},
		},
		{
			in:       "lowercase",
			expected: []string{"lowercase"},
		},
		{
			in:       "snake_case",
			expected: []string{"snake", "_", "case"},
		},
		{
			in:       "Class",
			expected: []string{"Class"},
		},
		{
			in:       "snake_caseClass",
			expected: []string{"snake", "_", "case", "Class"},
		},
		{
			in:       "MyClass",
			expected: []string{"My", "Class"},
		},
		{
			in:       "MyC",
			expected: []string{"My", "C"},
		},
		{
			in:       "HTML",
			expected: []string{"HTML"},
		},
		{
			in:       "PDFLoader",
			expected: []string{"PDF", "Loader"},
		},
		{
			in:       "AString",
			expected: []string{"A", "String"},
		},
		{
			in:       "SimpleXMLParser",
			expected: []string{"Simple", "XML", "Parser"},
		},
		{
			in:       "vimRPCPlugin",
			expected: []string{"vim", "RPC", "Plugin"},
		},
		{
			in:       "GL11Version",
			expected: []string{"GL", "11", "Version"},
		},
		{
			in:       "99Bottles",
			expected: []string{"99", "Bottles"},
		},
		{
			in:       "May5",
			expected: []string{"May", "5"},
		},
		{
			in:       "BFG9000",
			expected: []string{"BFG", "9000"},
		},
		{
			in:       "BöseÜberraschung",
			expected: []string{"Böse", "Überraschung"},
		},
		{
			in:       "Two  spaces",
			expected: []string{"Two", "  ", "spaces"},
		},
		{
			in:       "BadUTF8\xe2\xe2\xa1",
			expected: []string{"BadUTF8\xe2\xe2\xa1"},
		},
	}

	for i, testCase := range testCases {
		it.Equal(testCase.expected, parse(testCase.in), "Case(#%d)", i)
	}
}

func Benchmark_parse(b *testing.B) {
	testCases := []struct {
		in       string
		expected []string
	}{
		{
			in:       "",
			expected: []string{""},
		},
		{
			in:       "lowercase",
			expected: []string{"lowercase"},
		},
		{
			in:       "snake_case",
			expected: []string{"snake", "_", "case"},
		},
		{
			in:       "Class",
			expected: []string{"Class"},
		},
		{
			in:       "snake_caseClass",
			expected: []string{"snake", "_", "case", "Class"},
		},
		{
			in:       "MyClass",
			expected: []string{"My", "Class"},
		},
		{
			in:       "MyC",
			expected: []string{"My", "C"},
		},
		{
			in:       "HTML",
			expected: []string{"HTML"},
		},
		{
			in:       "PDFLoader",
			expected: []string{"PDF", "Loader"},
		},
		{
			in:       "AString",
			expected: []string{"A", "String"},
		},
		{
			in:       "SimpleXMLParser",
			expected: []string{"Simple", "XML", "Parser"},
		},
		{
			in:       "vimRPCPlugin",
			expected: []string{"vim", "RPC", "Plugin"},
		},
		{
			in:       "GL11Version",
			expected: []string{"GL", "11", "Version"},
		},
		{
			in:       "99Bottles",
			expected: []string{"99", "Bottles"},
		},
		{
			in:       "May5",
			expected: []string{"May", "5"},
		},
		{
			in:       "BFG9000",
			expected: []string{"BFG", "9000"},
		},
		{
			in:       "BöseÜberraschung",
			expected: []string{"Böse", "Überraschung"},
		},
		{
			in:       "Two  spaces",
			expected: []string{"Two", "  ", "spaces"},
		},
		{
			in:       "BadUTF8\xe2\xe2\xa1",
			expected: []string{"BadUTF8\xe2\xe2\xa1"},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			parse(tc.in)
		}
	}
}
