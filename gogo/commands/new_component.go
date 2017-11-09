package commands

import (
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/golib/cli"
)

type _NewComponent struct{}

var (
	NewComponent *_NewComponent

	modelsDir      = []string{"app", "models"}
	controllersDir = []string{"app", "controllers"}

	model          = template.Must(template.New("gogo").Parse(modelTemplate))
	modelTest      = template.Must(template.New("gogo").Parse(modelTestTemplate))
	controller     = template.Must(template.New("gogo").Parse(controllerTemplate))
	controllerTest = template.Must(template.New("gogo").Parse(controllerTestTemplate))
)

type TemplateModel struct {
	Name          string
	LowerCaseName string
}

func (_ *_NewComponent) getNames(name string) (capitalName, lowercaseName string) {
	firstCase := name[0:1]
	rightCase := name[1:]
	capitalName = strings.ToUpper(firstCase) + rightCase
	lowercaseName = strings.ToLower(firstCase) + rightCase
	return
}

func (_ *_NewComponent) Command() cli.Command {
	return cli.Command{
		Name:   "generate",
		Usage:  "generate model/controller/test.",
		Flags:  NewComponent.Flags(),
		Action: NewComponent.Action(),
	}
}

func (_ *_NewComponent) Flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "topic",
			Value:  "component",
			Usage:  "specify topic name",
			EnvVar: "GOGO_COMPONENT",
		},
	}
}

func (_ *_NewComponent) Action() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		root, err := os.Getwd()
		if err != nil {
			stderr.Error(err.Error())

			return err
		}

		modelPath := path.Join(root, "models")
		controllerPath := path.Join(root, "controllers")

		componentName := path.Clean(ctx.Args().First())
		capitalName, lowercaseName := NewComponent.getNames(componentName)

		NewComponent.genControllerFile(controllerPath, capitalName, lowercaseName)
		NewComponent.genControllerTestFile(controllerPath, capitalName, lowercaseName)
		NewComponent.genModelFile(modelPath, capitalName, lowercaseName)
		NewComponent.genModelTestFile(modelPath, capitalName, lowercaseName)
		return nil
	}
}

func (_ *_NewComponent) genModelFile(file, capitalName, lowercaseName string) {
	f, err := os.OpenFile(path.Join(file, lowercaseName+".go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())

		return
	}

	err = model.Execute(f, &TemplateModel{
		Name:          capitalName,
		LowerCaseName: lowercaseName,
	})
	if err != nil {
		stderr.Errorf(err.Error())
	}
}

func (_ *_NewComponent) genModelTestFile(file, capitalName, lowercaseName string) {
	f, err := os.OpenFile(path.Join(file, lowercaseName+"_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())

		return
	}

	err = modelTest.Execute(f, &TemplateModel{
		Name:          capitalName,
		LowerCaseName: lowercaseName,
	})
	if err != nil {
		stderr.Errorf(err.Error())
	}
}

func (_ *_NewComponent) genControllerFile(file, capitalName, lowercaseName string) {
	f, err := os.OpenFile(path.Join(file, lowercaseName+".go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())

		return
	}

	err = controller.Execute(f, &TemplateModel{
		Name:          capitalName,
		LowerCaseName: lowercaseName,
	})
	if err != nil {
		stderr.Errorf(err.Error())
	}
}

func (_ *_NewComponent) genControllerTestFile(file, capitalName, lowercaseName string) {
	f, err := os.OpenFile(path.Join(file, lowercaseName+"_test.go"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())

		return
	}

	err = controllerTest.Execute(f, &TemplateModel{
		Name:          capitalName,
		LowerCaseName: lowercaseName,
	})
	if err != nil {
		stderr.Errorf(err.Error())
	}
}
