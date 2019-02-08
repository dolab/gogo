package commands

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/golib/cli"
)

// Generate a new app
var (
	Application *_Application

	appDirs = [][]string{
		{"app", "controllers"},
		{"app", "middlewares"},
		{"app", "models"},
		{"app", "protos"},
		{"config"},
		{"gogo", "clients"},
		{"gogo", "errors"},
		{"gogo", "pbs"},
		{"gogo", "services"},
		{"log"},
		{"tmp", "cache"},
		{"tmp", "pids"},
		{"tmp", "sockes"},
	}
)

type _Application struct{}

func (*_Application) Command() cli.Command {
	return cli.Command{
		Name:    "new",
		Aliases: []string{"n"},
		Usage:   "Create a new gogo application. gogo new myapp creates a new application called myapp in ./myapp",
		Flags:   Application.Flags(),
		Action:  Application.Action(),
	}
}

func (*_Application) Flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "namespace",
			Value:  "github.com",
			Usage:  "specify application package import path, default to github.com",
			EnvVar: "GOGO_NAMESPACE",
		},
		cli.BoolFlag{
			Name:   "go-install",
			Usage:  "run `go mod tidy` for dependences resolving",
			EnvVar: "GOGO_GO_INSTALL",
		},
	}
}

func (*_Application) Action() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		root, err := os.Getwd()
		if err != nil {
			log.Error(err.Error())

			return err
		}

		name := path.Clean(ctx.Args().First())
		if name != "" {
			root = path.Join(root, name)
		}

		// is app root is empty?
		files, err := ioutil.ReadDir(root)
		if err != nil && !os.IsNotExist(err) {
			log.Error(err.Error())

			return err
		}
		if len(files) > 0 {
			log.Warn(ErrNoneEmptyDirectory.Error())

			return ErrNoneEmptyDirectory
		}

		// generate app struct
		for _, dir := range appDirs {
			dirname := path.Join(root, path.Join(dir...))

			err := os.MkdirAll(dirname, os.ModePerm)
			if err != nil {
				log.Error(err.Error())

				return err
			}

			err = ioutil.WriteFile(path.Join(dirname, ".keep"), []byte(""), os.ModePerm)
			if err != nil {
				log.Error(err.Error())

				return err
			}
		}

		namespace := ctx.String("namespace")

		// generate .gitignore
		Application.genGitIgnore(path.Join(root, ".gitignore"), name, namespace)

		// // generate env.sh
		// Application.genEnvfile(path.Join(root, "env.sh"), name, namespace)

		// generate readme.md
		Application.genReadme(path.Join(root, "README.md"), name, namespace)

		// generate Makefile
		Application.genMakefile(path.Join(root, "Makefile"), name, namespace)

		// generate go.mod
		Application.genModfile(path.Join(root, "go.mod"), name, namespace)

		// generate default controller dependences
		Application.genControllers(path.Join(root, "app", "controllers"), name, namespace)

		// generate default filters
		Application.genFilters(path.Join(root, "app", "middlewares"), name, namespace)

		// generate default models
		Application.genModels(path.Join(root, "app", "models"), name, namespace)

		// generate default application.yml
		Application.genConfigFile(path.Join(root, "config", "application.yml"), name, namespace)

		// generate main.go
		Application.genMainFile(path.Join(root, "app"), name, namespace)

		// generate errors.go
		Application.genErrorsFile(path.Join(root, "gogo", "errors", "errors.go"), name, namespace)

		// auto install dependences
		if ctx.Bool("go-install") {
			Application.runGoInstall(root)
		}
		return nil
	}
}

func (*_Application) genGitIgnore(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("gitignore").Execute(fd, AppData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (*_Application) genReadme(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("readme").Execute(fd, AppData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (*_Application) genMakefile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("makefile").Execute(fd, AppData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (*_Application) genEnvfile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("env.sh").Execute(fd, AppData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (*_Application) genModfile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("go.mod").Execute(fd, AppData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (*_Application) genControllers(root, app, namespace string) {
	data := AppData{
		Namespace:   namespace,
		Application: app,
	}

	// application.go
	fd, err := os.OpenFile(path.Join(root, "application.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("application.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
		return
	}

	// application_middlewares.go
	fd, err = os.OpenFile(path.Join(root, "application_middlewares.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("application_middlewares.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
		return
	}

	// testing_test.go
	fd, err = os.OpenFile(path.Join(root, "testing_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("application_testing.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
		return
	}

	// application_config.go
	fd, err = os.OpenFile(path.Join(root, "application_config.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("application_config.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
		return
	}

	// application_config_test.go
	fd, err = os.OpenFile(path.Join(root, "application_config_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("application_config_test.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
		return
	}

	// getting_start.go
	fd, err = os.OpenFile(path.Join(root, "getting_start.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("getting_start.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
		return
	}

	// getting_start_test.go
	fd, err = os.OpenFile(path.Join(root, "getting_start_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("getting_start_test.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
		return
	}
}

func (*_Application) genFilters(root, app, namespace string) {
	data := AppData{
		Namespace:   namespace,
		Application: app,
	}

	// testing_test.go
	fd, err := os.OpenFile(path.Join(root, "testing_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("filter_testing.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
	}

	// recovery.go
	fd, err = os.OpenFile(path.Join(root, "recovery.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("filter_recovery.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
	}

	// recovery_test.go
	fd, err = os.OpenFile(path.Join(root, "recovery_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("filter_recovery_test.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
	}
}

func (*_Application) genModels(root, app, namespace string) {
	data := AppData{
		Namespace:   namespace,
		Application: app,
	}

	// model.go
	fd, err := os.OpenFile(path.Join(root, "model.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("model.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
	}

	// model_test.go
	fd, err = os.OpenFile(path.Join(root, "model_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("model_test.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
	}
}

func (*_Application) genConfigFile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("application_config.yml").Execute(fd, AppData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (*_Application) genMainFile(root, app, namespace string) {
	// gen main.yml
	fd, err := os.OpenFile(path.Join(root, "main.yml"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("main.yml").Execute(fd, AppData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}

	// gen main.go
	fd, err = os.OpenFile(path.Join(root, "main.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("main.go").Execute(fd, AppData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (*_Application) genErrorsFile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("errors.go").Execute(fd, AppData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (*_Application) runGoInstall(root string) {
	// TODO: run go mod tidy auto
}
