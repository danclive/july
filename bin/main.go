package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/danclive/nson-go"

	"github.com/danclive/july"
	"github.com/danclive/queen-go/conn"
	"github.com/danclive/queen-go/crypto"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment() // or NewProduction, or NewDevelopment
	defer logger.Sync()

	slotId, err := nson.MessageIdFromHex("017477033867f215f0c5341e")
	if err != nil {
		logger.Panic("", zap.String("err", err.Error()))
	}

	config := conn.Config{
		Addrs:        []string{"snple.com:8888"},
		SlotId:       slotId,
		EnableCrypto: true,
		CryptoMethod: crypto.Aes128Gcm,
		AccessKey:    "fcbd6ea1e8c94dfc6b84405e",
		SecretKey:    "b14cd7bf94f0e3374e7fc4d4",
		Debug:        false,
	}

	options := july.Options{
		Log:         logger,
		QueenEnable: true,
		QueenConfig: &config,
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
