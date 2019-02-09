package commands

import (
	"io/ioutil"
	"path"
	"text/template"

	"github.com/dolab/gogo/cmd/gogo/templates"
	"github.com/dolab/logger"
	yaml "gopkg.in/yaml.v2"
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
	log.SetFlag(1 | 2)

	box = templates.Box()
}

// An AppData is useful for template rendering
type AppData struct {
	Version     string `yaml:"version"`
	Namespace   string `yaml:"namespace"`
	Application string `yaml:"application"`
}

// LoadAppData reads main.yml created with generation and parses metadata of app
func LoadAppData(root string) (app *AppData, err error) {
	root = ensureAppRoot(root)

	data, err := ioutil.ReadFile(path.Join(root, "app", "main.yml"))
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &app)
	return
}

// ImportPrefix returns import path of generated application
func (app *AppData) ImportPrefix() string {
	return app.Namespace + "/" + app.Application
}
