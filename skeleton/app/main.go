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
	cfgPath string // app source path, e.g. /home/deploy/websites/helloapp
)

func main() {
	flag.StringVar(&runMode, "runMode", "development", "example -runMode=[development|test|production]")
	flag.StringVar(&cfgPath, "cfgPath", "", "example -cfgPath=/path/to/[config/application.json]")
	flag.Parse()

	// verify run mode
	if mode := gogo.RunMode(runMode); !mode.IsValid() {
		flag.PrintDefaults()
		return
	}

	// adjust src path
	if cfgPath == "" {
		var err error

		cfgPath, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	} else {
		cfgPath = path.Clean(cfgPath)
	}

	app := gogo.New(runMode, cfgPath)
	app.NewService(controllers.New())
	app.Serve()
}
