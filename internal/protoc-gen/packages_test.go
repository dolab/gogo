package gen

import "testing"

func TestParseGoPackageOption(t *testing.T) {
	testcase := func(in, wantImport, wantPkg string) func(*testing.T) {
		return func(t *testing.T) {
			in := ""
			wantImport, wantPkg := "", ""
			haveImport, havePkg := ParseGoPackageOption(in)
			if haveImport != wantImport {
				t.Errorf("wrong importPath, have=%q want=%q", haveImport, wantImport)
			}
			if havePkg != wantPkg {
				t.Errorf("wrong packageName, have=%q want=%q", havePkg, wantPkg)
			}
		}
	}

	t.Run("empty string", testcase("", "", ""))
	t.Run("bare package", testcase("foo", "", "foo"))
	t.Run("full import", testcase("github.com/example/foo", "github.com/example/foo", "foo"))
	t.Run("full import with override", testcase("github.com/example/foo;bar", "github.com/example/foo", "bar"))
}
