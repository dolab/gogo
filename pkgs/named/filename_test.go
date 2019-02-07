package named

import (
	"testing"

	"github.com/golib/assert"
)

func Test_Basename(t *testing.T) {
	it := assert.New(t)

	testCases := []struct {
		in       string
		expected string
	}{
		{"", ""},
		{"filename", "filename"},
		{"filename.ext", "filename"},
		{"file.name.ext", "file.name"},
		{"/path/to/filename", "filename"},
		{"/path/to/filename.ext", "filename"},
		{"/path/to/file.name.ext", "file.name"},
		{"/path/to.ext/filename.ext", "filename"},
		{"/path/to.ext/file.name.ext", "file.name"},
	}
	for i, testCase := range testCases {
		it.Equal(testCase.expected, Basename(testCase.in), "Case(#%d)", i)
	}
}
