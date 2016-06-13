package commands

import (
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/golib/cli"
)

var (
	NewModel *_NewModel

	modelDir = []string{"app", "models"}

	appModel     = template.Must(template.New("gogo").Parse(modelTemplate))
	appModelTest = template.Must(template.New("gogo").Parse(modelTestTemplate))
)

type templateModel struct {
	LowerTopic string
	Topic      string
}

type _NewModel struct{}

func (_ *_NewModel) Command() cli.Command {
	return cli.Command{
		Name:   "newmodel",
		Usage:  "Create a new gogo model in current path.use like: gogo newmodel --topic:modelName",
		Flags:  NewModel.Flags(),
		Action: NewModel.Action(),
	}
}

func (_ *_NewModel) Flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "topic",
			Value:  "model",
			Usage:  "specify topic name",
			EnvVar: "model",
		},
	}
}
func (_ *_NewModel) getTopics(topicName string) (upTopic, LowTopic string) {
	firstCase := topicName[0:1]
	rightCase := topicName[1:]
	upTopic = strings.ToUpper(firstCase) + rightCase
	LowTopic = strings.ToLower(firstCase) + rightCase
	return
}

func (_ *_NewModel) Action() func(*cli.Context) {
	return func(ctx *cli.Context) {
		modelName := "model"
		if ctx.NumFlags() > 0 {
			modelName = ctx.String("topic")
		}
		root, err := os.Getwd()
		if err != nil {
			stderr.Error(err.Error())
			return
		}

		// generate modelFile
		NewModel.genModelFile(path.Join(root, modelName+".go"), modelName)
		NewModel.genModelTestFile(path.Join(root, modelName+"_test.go"), modelName)
	}
}

func (_ *_NewModel) genModelFile(file, topic string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	upTopic, lowTopic := NewModel.getTopics(topic)
	err = appModel.Execute(fd, templateModel{
		Topic:      upTopic,
		LowerTopic: lowTopic,
	})
	if err != nil {
		stderr.Error(err.Error())
	}
}

func (_ *_NewModel) genModelTestFile(file, topic string) {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		stderr.Error(err.Error())
		return
	}

	upTopic, _ := NewModel.getTopics(topic)
	err = appModelTest.Execute(fd, templateModel{
		Topic: upTopic,
	})
	if err != nil {
		stderr.Error(err.Error())
	}
}
