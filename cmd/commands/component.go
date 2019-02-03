package commands

import (
	"os"
	"path"
	"strings"

	"github.com/dolab/gogo/pkgs/named"
	"github.com/golib/cli"
)

var (
	Component *_Component

	comDirs = map[ComponentType][]string{
		ComTypeController: {"app", "controllers"},
		ComTypeFilter: {"app", "middlewares"},
		ComTypeModel:      {"app", "models"},
	}
)

type ComTemplateModel struct {
	Name string
	Args []string
}

type _Component struct{}

func (_ *_Component) Command() cli.Command {
	return cli.Command{
		Name:    "generate",
		Aliases: []string{"g"},
		Usage:   "generate controller and model components.",
		Flags:   Component.Flags(),
		Action:  Component.Action(),
		Subcommands: cli.Commands{
			{
				Name:    "controller",
				Aliases: []string{"c"},
				Usage:   "generate controller component.",
				Flags: []cli.Flag{
					cli.StringSliceFlag{
						Name:  "actions",
						Usage: "specify actions to generating, defaults to gogo resources.",
						Value: &cli.StringSlice{"index", "create", "show", "update", "destroy"},
					},
				},
				Action: Component.NewController(),
			},
			{
				Name:    "filter",
				Aliases: []string{"f"},
				Usage:   "generate filter func component.",
				Flags:   []cli.Flag{},
				Action:  Component.NewMiddleware(),
			},
			{
				Name:    "model",
				Aliases: []string{"m"},
				Usage:   "generate model component.",
				Flags:   []cli.Flag{},
				Action:  Component.NewModel(),
			},
		},
	}
}

func (_ *_Component) Flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:   "skip-testing",
			EnvVar: "GOGO_COMPONENT",
		},
		cli.StringSliceFlag{
			Name:  "controller-actions",
			Value: &cli.StringSlice{"index", "create", "show", "update", "destroy"},
		},
	}
}

func (_ *_Component) Action() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		// controller
		err := Component.newComponent(ComTypeController, name, ctx.StringSlice("controller-actions")...)
		if err != nil {
			return err
		}

		// model
		err = Component.newComponent(ComTypeModel, name)
		if err != nil {
			return err
		}

		return nil
	}
}

func (_ *_Component) NewController() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		actions := ctx.StringSlice("actions")
		if len(actions) == 0 {
			actions = []string{"index", "create", "show", "update", "destroy"}
		}

		return Component.newComponent(ComTypeController, name, actions...)
	}
}

func (_ *_Component) NewMiddleware() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		return Component.newComponent(ComTypeFilter, name)
	}
}

func (_ *_Component) NewModel() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		return Component.newComponent(ComTypeModel, name)
	}
}

func (_ *_Component) newComponent(com ComponentType, name string, args ...string) error {
	if !com.Valid() {
		return ErrComponentType
	}

	root, err := os.Getwd()
	if err != nil {
		log.Error(err.Error())

		return err
	}

	comRoot := com.Root(root)
	if strings.Count(comRoot, "/app/") != 1 {
		return ErrInvalidRoot
	}

	comName := name
	comArgs := &ComTemplateModel{
		Name: Component.toCamelCase(comName),
		Args: args,
	}

	// generate xxx.go
	fd, err := os.OpenFile(path.Join(comRoot, Component.toFilename(comName)), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())

		return err
	}

	err = box.Lookup("template_"+com.String()).Execute(fd, comArgs)
	if err != nil {
		log.Errorf(err.Error())

		return err
	}

	// generate xxx_test.go
	fd, err = os.OpenFile(path.Join(comRoot, Component.toFilename(comName+"_test")), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(err.Error())

		return err
	}

	err = box.Lookup("template_"+com.String()+"_test").Execute(fd, comArgs)
	if err != nil {
		log.Errorf(err.Error())

		return err
	}

	return nil
}

func (_ *_Component) toCamelCase(name string) (capitalName string) {
	names := named.ToCamelCase(name)
	for i, tmpname := range names {
		tmpnames := strings.Split(tmpname, "_")
		for j := 0; j < len(tmpnames); j++ {
			tmpnames[j] = strings.Title(tmpnames[j])
		}

		names[i] = strings.Join(tmpnames, "")
	}

	return strings.Join(names, "")
}

func (_ *_Component) toFilename(name string) (filename string) {
	filenames := []string{}

	names := named.ToCamelCase(name)
	for i := 0; i < len(names); i++ {
		if names[i] == "_" {
			continue
		}

		filenames = append(filenames, strings.ToLower(names[i]))
	}

	return strings.Join(filenames, "_") + ".go"
}
