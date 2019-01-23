package templates

var (
	mainTemplate = `package main

import (
	"flag"
	"os"
	"path"

	"github.com/dolab/gogo"

	"{{.Namespace}}/{{.Application}}/app/controllers"
)

var (
	runMode string // app run mode, available values are [development|test|production], default to development
	srcPath string // app config path, e.g. /home/deploy/websites/helloapp
)

func main() {
	flag.StringVar(&runMode, "runMode", "development", "{{.Application}} -runMode=[development|test|production]")
	flag.StringVar(&srcPath, "srcPath", "", "{{.Application}} -srcPath=/path/to/[config/application.json]")
	flag.Parse()

	// verify run mode
	if mode := gogo.RunMode(runMode); !mode.IsValid() {
		flag.PrintDefaults()
		return
	}

	// adjust src path
	if srcPath == "" {
		var err error

		srcPath, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	} else {
		srcPath = path.Clean(srcPath)
	}

	gogo.New(runMode, srcPath).NewService(controllers.New()).Serve()
}
`
)
