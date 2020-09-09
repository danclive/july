package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/danclive/july"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment() // or NewProduction, or NewDevelopment
	defer logger.Sync()

	options := july.Options{
		Log: logger,
	}

	crate, err := july.NewCrate(options)
	if err != nil {
		logger.Panic("", zap.String("err", err.Error()))
	}

	crate.Run()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	<-signalCh

	err = crate.Stop()
	if err != nil {
		logger.Panic("", zap.String("err", err.Error()))
	}
}
