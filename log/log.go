package log

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger
var Suger *zap.SugaredLogger

func Init(debug bool) {
	if debug {
		Logger, _ = zap.NewDevelopment()
	} else {
		Logger, _ = zap.NewProduction()
	}

	Suger = Logger.Sugar()
}
