package gen

import (
	"github.com/dolab/logger"
)

var (
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

}

func Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

func Failf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}
