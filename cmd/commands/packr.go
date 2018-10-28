package commands

import (
	"log"

	"github.com/gobuffalo/packr"
	"github.com/golib/cli"
)

var (
	Packr *_Packr
)

type _Packr struct{}

func (_ *_Packr) Command(box packr.Box) cli.Command {
	return cli.Command{
		Name:    "packr",
		Usage:   "print `file` content of named template.",
		Aliases: []string{"p"},
		Action:  Packr.Action(box),
	}
}

func (_ *_Packr) Action(box packr.Box) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := ctx.Args().First()
		if name == "" {
			cli.ShowSubcommandHelp(ctx)
			return nil
		}

		s, err := box.MustString(name)
		if err != nil {
			log.Printf("packr.MustString(%s): %v", ctx.String("name"), err)
		} else {
			println(s)
		}

		return err
	}
}
