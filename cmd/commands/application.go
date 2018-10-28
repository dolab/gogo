package commands

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/golib/cli"
)

var (
	Application *_Application

	appDirs = [][]string{
		{"app", "controllers"},
		{"app", "middlewares"},
		{"app", "models"},
		{"config"},
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
		cli.StringFlag{
			Name:   "project",
			Value:  "gogo-project",
			Usage:  "specify application project name, default to gogo-project",
			EnvVar: "GOGO_PROJECT",
		},
		cli.BoolFlag{
			Name:   "skip-install",
			Usage:  "skip `go get` dependences installation",
			EnvVar: "GOGO_SKIP_INSTALL",
		},
	}
}

func (_ *_Application) Action() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		root, err := os.Getwd()
		if err != nil {
			stderr.Error(err.Error())

			return err
		}

		appName := path.Clean(ctx.Args().First())
		if appName != "" {
			root += "/" + strings.TrimPrefix(appName, "/")
		}

		// is app root is empty?
		files, err := ioutil.ReadDir(root)
		if err != nil && !os.IsNotExist(err) {
			stderr.Error(err.Error())

			return err
		}

		if len(files) > 0 {
			stderr.Warn(ErrNoneEmptyDirectory.Error())

			return ErrNoneEmptyDirectory
		}

		// create app root dir
		appRoot := path.Join(root, "gogo")

		err = os.MkdirAll(appRoot, os.ModePerm)
		if err != nil {
			stderr.Error(err.Error())

			return err
		}

		// generate app struct
		for _, dir := range appDirs {
			absPath := path.Join(appRoot, path.Join(dir...))

			err := os.MkdirAll(absPath, os.ModePerm)
			if err != nil {
				stderr.Error(err.Error())

				return err
			}

			err = ioutil.WriteFile(path.Join(absPath, ".keep"), []byte(""), os.ModePerm)
			if err != nil {
				stderr.Error(err.Error())

				return err
			}
		}

		appNamespace := ctx.String("namespace")

		// generate .gitignore
		Application.genGitIgnore(path.Join(root, ".gitignore"), appName, appNamespace)

		// generate env.sh
		Application.genEnvFile(path.Join(root, "env.sh"), appName, appNamespace)

		// generate Makefile
		Application.genMakefile(path.Join(root, "Makefile"), appName, appNamespace)

		// generate readme.md
		Application.genReadme(path.Join(root, "README.md"), appName, appNamespace)

		// generate default controller dependences
		Application.genControllers(path.Join(appRoot, "app", "controllers"), appName, appNamespace)

		// generate default middlewares
		Application.genMiddlewares(path.Join(appRoot, "app", "middlewares"), appName, appNamespace)

		// generate default models
		Application.genModels(path.Join(appRoot, "app", "models"), appName, appNamespace)

		// generate default application.json
		Application.genConfigFile(path.Join(appRoot, "config", "application.json"), appName, appNamespace)

		// generate main.go
		Application.genMainFile(path.Join(appRoot, "main.go"), appName, appNamespace)

		// // auto install dependences
		// if !ctx.Bool("skip-install") {
		// 	Application.getDependences()
		// }
		return nil
	}
}

func (_ *_Application) genEnvFile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("env").Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		stderr.Error(err.Error())
	}
}

func (_ *_Application) genMakefile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("makefile").Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		stderr.Error(err.Error())
	}
}

func (_ *_Application) genGitIgnore(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("gitignore").Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		stderr.Error(err.Error())
	}
}

func (_ *_Application) genReadme(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("readme").Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		stderr.Error(err.Error())
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
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("application").Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// testing_test.go
	fd, err = os.OpenFile(path.Join(root, "testing_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("application_testing").Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// application_config.go
	fd, err = os.OpenFile(path.Join(root, "application_config.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("application_config").Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// application_config_test.go
	fd, err = os.OpenFile(path.Join(root, "application_config_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("application_config_test").Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// getting_start.go
	fd, err = os.OpenFile(path.Join(root, "getting_start.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("getting_start").Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// getting_start_test.go
	fd, err = os.OpenFile(path.Join(root, "getting_start_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("getting_start_test").Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
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
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("middleware_testing").Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// recovery.go
	fd, err = os.OpenFile(path.Join(root, "recovery.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("middleware_recovery").Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// recovery_test.go
	fd, err = os.OpenFile(path.Join(root, "recovery_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("middleware_recovery_test").Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
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
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("model").Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// model_test.go
	fd, err = os.OpenFile(path.Join(root, "model_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("model_test").Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}
}

func (_ *_Application) genConfigFile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("application_config_json").Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		stderr.Error(err.Error())
	}
}

func (_ *_Application) genMainFile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = apptpl.Lookup("main").Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		stderr.Error(err.Error())
	}
}
