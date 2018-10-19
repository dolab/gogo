package commands

import (
	"errors"
	"os"
	"path"
	"strings"

	"github.com/golib/cli"
)

const (
	_comType ComponentType = iota
	ComTypeController
	ComTypeMiddleware
	ComTypeModel
	comType_
)

var (
	Component *_Component

	comDirs = map[ComponentType][]string{
		ComTypeController: {"app", "controllers"},
		ComTypeMiddleware: {"app", "middlewares"},
		ComTypeModel:      {"app", "models"},
	}
)

type ComponentType int

func (ct ComponentType) Valid() bool {
	return ct > _comType && ct < comType_
}

func (ct ComponentType) Root(pwd string) string {
	dirs, ok := comDirs[ct]
	if !ok {
		return pwd
	}

	return path.Clean(path.Join(pwd, "gogo", path.Join(dirs...)))
}

func (ct ComponentType) String() string {
	switch ct {
	case ComTypeController:
		return "controller"

	case ComTypeMiddleware:
		return "middleware"

	case ComTypeModel:
		return "model"

	}

	return ""
}

type ComTemplateModel struct {
	Name string
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
				Flags:   []cli.Flag{},
				Action:  Component.NewController(),
			},
			{
				Name:    "middleware",
				Aliases: []string{"w"},
				Usage:   "generate middleware component.",
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
	}
}

func (_ *_Component) Action() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		// controller
		err := Component.newComponent(name, ComTypeController)
		if err != nil {
			return err
		}

		// model
		err = Component.newComponent(name, ComTypeModel)
		if err != nil {
			return err
		}

		return nil
	}
}

func (_ *_Component) NewController() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		return Component.newComponent(name, ComTypeController)
	}
}

func (_ *_Component) NewMiddleware() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		return Component.newComponent(name, ComTypeMiddleware)
	}
}

func (_ *_Component) NewModel() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		name := path.Clean(ctx.Args().First())

		return Component.newComponent(name, ComTypeModel)
	}
}

func (_ *_Component) newComponent(name string, ctype ComponentType) error {
	if !ctype.Valid() {
		return errors.New("Invalid component type")
	}

	root, err := os.Getwd()
	if err != nil {
		stderr.Error(err.Error())

		return err
	}

	comRoot := ctype.Root(root)
	comName := name

	// generate xxx.go
	fd, err := os.OpenFile(path.Join(comRoot, Component.toFilename(comName)), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())

		return err
	}

	err = apptpl.Lookup("template_"+ctype.String()).Execute(fd, &ComTemplateModel{
		Name: Component.toCamelCase(comName),
	})
	if err != nil {
		stderr.Errorf(err.Error())

		return err
	}

	// generate xxx_test.go
	fd, err = os.OpenFile(path.Join(comRoot, Component.toFilename(comName+"_test")), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())

		return err
	}

	err = apptpl.Lookup("template_"+ctype.String()+"_test").Execute(fd, &ComTemplateModel{
		Name: Component.toCamelCase(comName),
	})
	if err != nil {
		stderr.Errorf(err.Error())

		return err
	}

	return nil
}

func (_ *_Component) toCamelCase(name string) (capitalName string) {
	names := ToCamelCase(name)
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

	names := ToCamelCase(name)
	for i := 0; i < len(names); i++ {
		if names[i] == "_" {
			continue
		}

		filenames = append(filenames, strings.ToLower(names[i]))
	}

	return strings.Join(filenames, "_") + ".go"
}
