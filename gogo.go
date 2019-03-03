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
	FindModeConfigFile = func(mode, cfgfile string) string {
		// adjust cfgfile
		cfgfile = path.Clean(cfgfile)

		if len(cfgfile) == 0 || cfgfile == "." || cfgfile == ".." {
			return GogoSchema
		}

		// is cfgfile exist?
		finfo, ferr := os.Stat(cfgfile)
		if ferr != nil {
			return GogoSchema
		}

		// is cfgfile a regular file?
		if finfo.Mode()&os.ModeType == 0 {
			return cfgfile
		}

		filename := "application.yml"
		switch RunMode(mode) {
		case Development:
			// try application.development.yml
			filename = "application.development.yml"

		case Test:
			// try application.test.yml
			filename = "application.test.yml"

		case Production:
			// skip

		}

		filepath := path.Join(cfgfile, "config", filename)
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			filepath = path.Join(cfgfile, "config", "application.yml")
		}

		return filepath
	}

	// FindInterceptorConfigFile returns config file for middlewares.
	FindInterceptorConfigFile = func(mode, cfgfile string) string {
		// adjust cfgfile
		cfgfile = path.Clean(cfgfile)

		if len(cfgfile) == 0 || cfgfile == "." || cfgfile == ".." {
			return GogoSchema
		}

		// is cfgfile exist?
		finfo, ferr := os.Stat(cfgfile)
		if ferr != nil {
			return GogoSchema
		}

		// is cfgfile a regular file?
		if finfo.Mode()&os.ModeType == 0 {
			return cfgfile
		}

		// resolve cfgfile with run mode
		filename := "interceptors.yml"
		switch RunMode(mode) {
		case Development:
			filename = "interceptors.development.yml"

		case Test:
			filename = "interceptors.test.yml"

		case Production:
			// skip

		}

		filepath := path.Join(cfgfile, "config", filename)
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			filepath = path.Join(cfgfile, "config", "interceptors.yml")
		}

		return filepath
	}
)

// New creates application server with config resolved
// from file <cfgPath>/config/application[.<runMode>].yml.
//
// NOTE: You can custom resolver by overwriting FindModeConfigFile.
func New(runMode, cfgPath string) *AppServer {
	// resolve config from application.yml
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
	if err := config.LoadInterceptors(); err != nil {
		logger.Errorf("config.LoadInterceptors(): %v", err)
	}

	return NewAppServer(config, logger)
}

// NewParams returns *params.Params related with http request
func NewParams(r *http.Request) *params.Params {
	return params.New(r)
}
