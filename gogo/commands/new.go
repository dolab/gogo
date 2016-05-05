package commands

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/golib/cli"
)

var (
	New *_New

	appDirs = [][]string{
		[]string{"app", "controllers"},
		[]string{"app", "middlewares"},
		[]string{"app", "models"},
		[]string{"config"},
		[]string{"lib"},
		[]string{"log"},
		[]string{"tmp", "cache"},
		[]string{"tmp", "pids"},
		[]string{"tmp", "sockes"},
	}

	appEnv       = template.Must(template.New("gogo").Parse(envTemplate))
	appMakefile  = template.Must(template.New("gogo").Parse(strings.Replace(makefileTemplate, "    ", "\t", -1)))
	appGitIgnore = template.Must(template.New("gogo").Parse(gitIgnoreTemplate))

	appController       = template.Must(template.New("gogo").Parse(applicationTemplate[0]))
	appControllerTest   = template.Must(template.New("gogo").Parse(applicationTemplate[1]))
	appConfig           = template.Must(template.New("gogo").Parse(configTemplate[0]))
	appConfigTest       = template.Must(template.New("gogo").Parse(configTemplate[1]))
	appGettingStart     = template.Must(template.New("gogo").Parse(gettingStartTemplate[0]))
	appGettingStartTest = template.Must(template.New("gogo").Parse(gettingStartTemplate[1]))
	appMiddleware       = template.Must(template.New("gogo").Parse(middlewareTemplate[0]))
	appMiddlewareTest   = template.Must(template.New("gogo").Parse(middlewareTemplate[1]))
	appMiddlewareInit   = template.Must(template.New("gogo").Parse(middlewareTemplate[2]))
	appJSON             = template.Must(template.New("gogo").Parse(jsonTemplate))
	appMain             = template.Must(template.New("gogo").Parse(mainTemplate))
)

type _New struct{}

func (_ *_New) Command() cli.Command {
	return cli.Command{
		Name:   "new",
		Usage:  "Create a new gogo application. gogo new myapp creates a new application called myapp in ./myapp",
		Flags:  New.Flags(),
		Action: New.Action(),
	}
}

func (_ *_New) Flags() []cli.Flag {
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

func (_ *_New) Action() func(*cli.Context) {
	return func(ctx *cli.Context) {
		root, err := os.Getwd()
		if err != nil {
			stderr.Error(err.Error())
			return
		}

		appName := path.Clean(ctx.Args().First())
		if appName != "" {
			root += "/" + strings.TrimPrefix(appName, "/")
		}

		// is app root is empty?
		files, err := ioutil.ReadDir(root)
		if err != nil && !os.IsNotExist(err) {
			stderr.Error(err.Error())
			return
		}

		if len(files) > 0 {
			stderr.Warn("Can't initialize a new gogo application within the unempty directory, please change to an empty directory first.")
			return
		}

		appRoot := path.Join(root, "gogo")

		// create app root dir
		err = os.MkdirAll(appRoot, os.ModePerm)
		if err != nil {
			stderr.Error(err.Error())
			return
		}

		// generate app struct
		for _, dir := range appDirs {
			absPath := path.Join(appRoot, path.Join(dir...))

			err := os.MkdirAll(absPath, os.ModePerm)
			if err != nil {
				println(err.Error())
				return
			}

			err = ioutil.WriteFile(path.Join(absPath, ".keep"), []byte(""), os.ModePerm)
			if err != nil {
				println(err.Error())
			}
		}

		appNamespace := ctx.String("namespace")

		// generate env.sh
		New.genEnvFile(path.Join(root, "env.sh"), appName, appNamespace)

		// generate Makefile
		New.genMakefile(path.Join(root, "Makefile"), appName, appNamespace)

		// generate .gitignore
		New.genGitIgnore(path.Join(root, ".gitignore"), appName, appNamespace)

		// generate default controller dependences
		New.genControllers(path.Join(appRoot, "app", "controllers"), appName, appNamespace)

		// generate default middlewares
		New.genMiddlewares(path.Join(appRoot, "app", "middlewares"), appName, appNamespace)

		// generate default application.json
		New.genConfigFile(path.Join(appRoot, "config", "application.json"), appName, appNamespace)

		// generate main.go
		New.genMainFile(path.Join(appRoot, "main.go"), appName, appNamespace)

		// // auto install dependences
		// if !ctx.Bool("skip-install") {
		// 	New.getDependences()
		// }
	}
}

func (_ *_New) genEnvFile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = appEnv.Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		stderr.Error(err.Error())
	}
}

func (_ *_New) genMakefile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = appMakefile.Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		stderr.Error(err.Error())
	}
}

func (_ *_New) genGitIgnore(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = appGitIgnore.Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		stderr.Error(err.Error())
	}
}

func (_ *_New) genControllers(root, app, namespace string) {
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

	err = appController.Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// application_test.go
	fd, err = os.OpenFile(path.Join(root, "application_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = appControllerTest.Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// application_config.go
	fd, err = os.OpenFile(path.Join(root, "application_config.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = appConfig.Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// application_config_test.go
	fd, err = os.OpenFile(path.Join(root, "application_config_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = appConfigTest.Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// getting_start.go
	fd, err = os.OpenFile(path.Join(root, "getting_start.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = appGettingStart.Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// getting_start_test.go
	fd, err = os.OpenFile(path.Join(root, "getting_start_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = appGettingStartTest.Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}
}

func (_ *_New) genMiddlewares(root, app, namespace string) {
	data := templateData{
		Namespace:   namespace,
		Application: app,
	}

	// init_test.go
	fd, err := os.OpenFile(path.Join(root, "init_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = appMiddlewareInit.Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// recovery.go
	fd, err = os.OpenFile(path.Join(root, "recovery.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = appMiddleware.Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}

	// recovery_test.go
	fd, err = os.OpenFile(path.Join(root, "recovery_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = appMiddlewareTest.Execute(fd, data)
	if err != nil {
		stderr.Error(err.Error())
	}
}

func (_ *_New) genConfigFile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = appJSON.Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		stderr.Error(err.Error())
	}
}

func (_ *_New) genMainFile(file, app, namespace string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	err = appMain.Execute(fd, templateData{
		Namespace:   namespace,
		Application: app,
	})
	if err != nil {
		stderr.Error(err.Error())
	}
}
