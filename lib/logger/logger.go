package logger

import (
	goLog "log"
	"os"

	"go.uber.org/zap"
)

// Get returns a zap logger
func Get() *zap.Logger {
	environment, _ := os.LookupEnv("ENVIRONMENT")

	var logr *zap.Logger
	var err error

	if environment == "production" {
		logr, err = zap.NewProduction()
	} else if environment == "staging" {
		conf := zap.NewProductionConfig()
		conf.Level = zap.NewAtomicLevelAt(zap.DebugLevel) // debug is useful in staging

		logr, err = conf.Build()
	} else {
		logr, err = zap.NewDevelopment(zap.AddStacktrace(zap.FatalLevel))
	}

	if err != nil {
		goLog.Panicf("logger could not be loaded: %s", err.Error())
	}

	return logr
}
