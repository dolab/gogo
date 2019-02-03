package commands

import (
	"os"
	"path"
	"strings"
)

func ensureAppRoot(root string) string {
	root = strings.TrimSpace(root)
	if len(root) == 0 {
		panic("Cannot resolve application root path")
	}

	// cd /path/to/<myapp>
	root = strings.TrimSuffix(root, "/")
	root = strings.TrimSuffix(root, "/app")
	root = strings.TrimSuffix(root, "/app/controllers")
	root = strings.TrimSuffix(root, "/app/middlewares")
	root = strings.TrimSuffix(root, "/app/models")
	root = strings.TrimSuffix(root, "/app/protos")
	root = strings.TrimSuffix(root, "/config")
	root = strings.TrimSuffix(root, "/gogo")
	root = strings.TrimSuffix(root, "/gogo/clients")
	root = strings.TrimSuffix(root, "/gogo/errors")
	root = strings.TrimSuffix(root, "/gogo/pbs")
	root = strings.TrimSuffix(root, "/gogo/services")
	root = strings.TrimSuffix(root, "/pkgs")

	// is this a gogo app project?
	appRoot := path.Join(root, "app", "main.yml")
	if stat, err := os.Stat(appRoot); err != nil || !stat.IsDir() {
		panic("It seems there is no app/main.yml file of gogo project within " + root)
	}

	gogoRoot := path.Join(root, "gogo")
	if stat, err := os.Stat(gogoRoot); err != nil || !stat.IsDir() {
		panic("It seems there is no gogo folder of gogo project within " + root)
	}

	return root
}
