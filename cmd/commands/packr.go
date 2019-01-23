package commands

import (
	"bytes"
	"text/template"

	"github.com/golib/cli"
)

var (
	Packr *_Packr
)

type _Packr struct{}

func (_ *_Packr) Command(box *template.Template) cli.Command {
	return cli.Command{
		Name:    "packr",
		Usage:   "print `file` content of named template.",
		Aliases: []string{"p"},
		Action:  Packr.Action(box),
	}
}

func (_ *_Packr) Action(box *template.Template) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := ctx.Args().First()
		if name == "" {
			cli.ShowSubcommandHelp(ctx)
			return nil
		}

		tpl := box.Lookup(name)
		if tpl != nil {
			if name == "doc.go" {
				buf := bytes.NewBuffer(nil)

				tpl.Execute(buf, nil)

				log.Printf("box.Lookup(%s): \n%s", name, buf.String())
			} else {
				log.Printf("box.Lookup(%s): OK", name)
			}
		} else {
			log.Printf("box.Lookup(%s): Not Found", name)
		}

		return nil
	}
}
