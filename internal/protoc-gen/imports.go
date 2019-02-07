package gen

import (
	"io/ioutil"
	"path"
	"regexp"
)

var (
	rimport = regexp.MustCompile(`(?m)^import\W+"([\w_]+\.proto)"\W*;$`)
)

// AllProtoImports resolves all proto files imported with generated
func AllProtoImports(filename string) (imps []string) {
	root := path.Base(filename)
	imps = parseProtoImports(filename)
	for _, imp := range imps {
		imps = append(imps, AllProtoImports(path.Join(root, imp))...)
	}

	return
}

func parseProtoImports(filename string) (imps []string) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	matches := rimport.FindAllStringSubmatch(string(b), -1)
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}

		imps = append(imps, match[1])
	}

	return
}
