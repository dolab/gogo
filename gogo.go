package gogo

import (
	"log"
	"os"
	"path"
)

var (
	// FindModeConfigFile returns config file for specified run mode.
	// You could custom your own run mode config file by overwriting.
	FindModeConfigFile = func(runMode, srcPath string) string {
		// adjust srcPath
		srcPath = path.Clean(srcPath)

		// is srcPath exist?
		finfo, ferr := os.Stat(srcPath)
		if ferr != nil {
			return SchemaConfig
		}

		// is srcPath a regular file?
		if !finfo.IsDir() && (finfo.Mode()&os.ModeSymlink == 0) {
			return srcPath
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

		file := path.Join(srcPath, "config", filename)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			file = path.Join(srcPath, "config", "application.json")
		}

		return file
	}
)

// New creates application server with config resolved
// from file <srcPath>/config/application[.<runMode>].json.
// NOTE: You can custom resolver by overwriting FindModeConfigFile.
func New(runMode, srcPath string) *AppServer {
	// resolve config from application.json
	config, err := NewAppConfig(FindModeConfigFile(runMode, srcPath))
	if err != nil {
		log.Fatalf("[GOGO] NewAppConfig(%s): %v", FindModeConfigFile(runMode, srcPath), err)
	}

	return NewWithConfiger(config)
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

	return NewAppServer(config, logger)
}
