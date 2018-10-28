package main

import (
	"os"

	_ "github.com/dolab/gogo"
	"github.com/dolab/gogo/cmd/commands"
	_ "github.com/dolab/gogo/cmd/gogo/templates"
	"github.com/gobuffalo/packr"
	"github.com/golib/cli"
)

var (
	box packr.Box
)

func init() {
	box = packr.NewBox("templates")
}

func main() {
	app := cli.NewApp()
	app.Name = "gogo"
	app.Version = "1.3.0"
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
