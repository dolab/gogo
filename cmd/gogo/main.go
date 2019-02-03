package main

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	_ "github.com/dolab/gogo"
	"github.com/dolab/gogo/cmd/commands"
	"github.com/dolab/gogo/cmd/gogo/templates"
	"github.com/golib/cli"
)

var (
	box *template.Template
)

func init() {
	_, filename, _, _ := runtime.Caller(0)

	var err error

	box, err = template.New("gogo").Funcs(template.FuncMap{
		"lowercase": strings.ToLower,
	}).ParseGlob(path.Join(filepath.Dir(filename), "templates", "*"))
	if err != nil {
		box = templates.Box()
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "gogo"
	app.Version = "3.1.0"
	app.Usage = "gogo COMMAND [ARGS]"

	app.Authors = []cli.Author{
		{
			Name:  "Spring MC",
			Email: "Heresy.MC@gmail.com",
		},
	}

	app.Commands = []cli.Command{
		commands.Packr.Command(box),
		commands.Application.Command(),
		commands.Component.Command(),
	}

	app.Run(os.Args)
}
