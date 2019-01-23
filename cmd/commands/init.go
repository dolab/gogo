package commands

import (
	"text/template"

	"github.com/dolab/gogo/cmd/gogo/templates"
	"github.com/dolab/logger"
)

var (
	box *template.Template
	log *logger.Logger
)

func init() {
	var err error

	// setup logger
	log, err = logger.New("stderr")
	if err != nil {
		panic(err.Error())
	}

	log.SetLevelByName("info")
	// log.SetFlag(log.Lshortfile)

	box = templates.Box()
}

type templateData struct {
	Namespace   string
	Application string
}
