package gogo

import (
	"log"
	"net/http"
	"os"
	"path"

	"github.com/dolab/gogo/internal/params"
)

var (
	// FindModeConfigFile returns config file for specified run mode.
	// You could custom your own resolver by overwriting it.
	FindModeConfigFile = func(runMode, cfgPath string) string {
		if len(cfgPath) == 0 {
			return GogoSchema
		}

		// adjust cfgPath
		cfgPath = path.Clean(cfgPath)

		// is cfgPath exist?
		finfo, ferr := os.Stat(cfgPath)
		if ferr != nil {
			return GogoSchema
		}

		// is cfgPath a regular file?
		if !finfo.IsDir() && (finfo.Mode()&os.ModeSymlink == 0) {
			return cfgPath
		}

		filename := "application.json"
		switch RunMode(runMode) {
		case Development:
			// try application.development.json
			filename = "application.development.json"

		case Test:
			// try application.test.json
			filename = "application.test.json"

		case Production:
			// skip

		}

		filepath := path.Join(cfgPath, "config", filename)
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			filepath = path.Join(cfgPath, "config", "application.json")
		}

		return filepath
	}

	// FindMiddlewareConfigFile returns config file for middlewares.
	FindMiddlewareConfigFile = func(cfgPath string) string {
		if len(cfgPath) == 0 {
			return GogoSchema
		}

		// adjust cfgPath
		cfgPath = path.Clean(cfgPath)

		// is cfgPath exist?
		finfo, ferr := os.Stat(cfgPath)
		if ferr != nil {
			return GogoSchema
		}

		// is cfgPath a regular file?
		if !finfo.IsDir() && (finfo.Mode()&os.ModeSymlink == 0) {
			return cfgPath
		}

		filename := "middlewares.yaml"
		filepath := path.Join(cfgPath, "config", filename)
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			filepath = path.Join(cfgPath, "config", "middlewares.json")
		}

		return filepath
	}
)

// New creates application server with config resolved
// from file <cfgPath>/config/application[.<runMode>].json.
//
// NOTE: You can custom resolver by overwriting FindModeConfigFile.
func New(runMode, cfgPath string) *AppServer {
	// resolve config from application.json
	config, err := NewAppConfig(FindModeConfigFile(runMode, cfgPath))
	if err != nil {
		log.Fatalf("[GOGO] NewAppConfig(%s): %v", FindModeConfigFile(runMode, cfgPath), err)
	}

	return NewWithConfiger(config)
}

// NewDefaults creates application server with defaults.
func NewDefaults() *AppServer {
	return New("development", "")
}

// NewWithConfiger creates application server with custom Configer and
// default Logger, see Configer for implements a new config provider.
func NewWithConfiger(config Configer) *AppServer {
	// init default logger
	logger := NewAppLogger(config.Section().Logger.Output, config.RunMode().String())

	return NewWithLogger(config, logger)
}

// NewWithLogger creates application server with custom Configer and Logger
func NewWithLogger(config Configer, logger Logger) *AppServer {
	// overwrite logger level and colorful
	logger.SetLevelByName(config.Section().Logger.LevelName)
	logger.SetColor(!config.RunMode().IsProduction())

	logger.Printf("Initialized %s in %s mode", config.RunName(), config.RunMode())

	// try load config of middlewares
	// NOTE: ignore returned error is ok!
	if err := config.LoadMiddlewares(); err != nil {
		logger.Errorf("LoadMiddlewares(): %v", err)
	}

	return NewAppServer(config, logger)
}

// NewParams returns *params.Params related with http request
func NewParams(r *http.Request) *params.Params {
	return params.New(r)
}
