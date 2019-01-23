package commands

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/golib/cli"
)

var (
	Application *_Application

	appDirs = [][]string{
		{"app", "controllers"},
		{"app", "middlewares"},
		{"app", "models"},
		{"app", "protos"},
		{"config"},
		{"gogo", "service"},
		{"lib"},
		{"log"},
		{"tmp", "cache"},
		{"tmp", "pids"},
		{"tmp", "sockes"},
	}
)

type _Application struct{}

func (_ *_Application) Command() cli.Command {
	return cli.Command{
		Name:    "new",
		Aliases: []string{"n"},
		Usage:   "Create a new gogo application. gogo new myapp creates a new application called myapp in ./myapp",
		Flags:   Application.Flags(),
		Action:  Application.Action(),
	}
}

func (_ *_Application) Flags() []cli.Flag {
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

func (_ *_Application) Action() cli.ActionFunc {
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

		// generate default middlewares
		Application.genMiddlewares(path.Join(root, "app", "middlewares"), name, namespace)

		// generate default models
		Application.genModels(path.Join(root, "app", "models"), name, namespace)

		// generate default application.json
		Application.genConfigFile(path.Join(root, "config", "application.json"), name, namespace)

		// generate main.go
		Application.genMainFile(path.Join(root, "app", "main.go"), name, namespace)

		// auto install dependences
		if ctx.Bool("go-install") {
			Application.runGoInstall(root)
		}
		return nil
	}
}

func (_ *_Application) genGitIgnore(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("gitignore").Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (_ *_Application) genReadme(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("readme").Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (_ *_Application) genMakefile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("makefile").Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (_ *_Application) genEnvfile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("env.sh").Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (_ *_Application) genModfile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("go.mod").Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (_ *_Application) genControllers(root, app, namespace string) {
	data := templateData{
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

func (_ *_Application) genMiddlewares(root, app, namespace string) {
	data := templateData{
		Namespace:   namespace,
		Application: app,
	}

	// testing_test.go
	fd, err := os.OpenFile(path.Join(root, "testing_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("middleware_testing.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
	}

	// recovery.go
	fd, err = os.OpenFile(path.Join(root, "recovery.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("middleware_recovery.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
	}

	// recovery_test.go
	fd, err = os.OpenFile(path.Join(root, "recovery_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := box.Lookup("middleware_recovery_test.go").Execute(fd, data); err != nil {
		log.Error(err.Error())
	}
}

func (_ *_Application) genModels(root, app, namespace string) {
	data := templateData{
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

func (_ *_Application) genConfigFile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("application_config.json").Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (_ *_Application) genMainFile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = box.Lookup("main.go").Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		log.Error(err.Error())
	}
}

func (_ *_Application) runGoInstall(root string) {
	// TODO: run go mod tidy auto
}
