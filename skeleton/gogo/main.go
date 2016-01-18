package main

import (
	"flag"
	"os"
	"path"

	"github.com/dolab/gogo"

	"github.com/skeleton/app/controllers"
)

var (
	runMode string // app run mode, available values are [development|test|production], default to development
	srcPath string // app source path, e.g. /home/deploy/websites/helloapp
)

func main() {
	flag.StringVar(&runMode, "runMode", "development", "example -runMode=[development|test|production]")
	flag.StringVar(&srcPath, "srcPath", "", "example -srcPath=/path/to/source")
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

	controllers.New(runMode, srcPath).Run()
}
