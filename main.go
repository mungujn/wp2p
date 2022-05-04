package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/mungujn/web-exp/config"
	serverHTTP "github.com/mungujn/web-exp/server"
	"github.com/mungujn/web-exp/system"
)

// main is the entry point of the application
func main() { // nolint:funlen,gocyclo
	// read service cfg from os env
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}

	// init logger
	initLogger(cfg.LogLevel)

	log.Info("Service starting...")

	// prepare main context
	ctx, cancel := context.WithCancel(context.Background())
	setupGracefulShutdown(cancel)

	// init distributed system
	sys, err := system.New(ctx, cfg)
	if err != nil {
		log.WithError(err).Fatal("system init error")
	}

	// initializing http server
	httpSrv, err := serverHTTP.New(
		cfg.HTTPServerCfg,
		sys,
	)
	if err != nil {
		log.WithError(err).Fatal("http server init")
	}

	var wg = &sync.WaitGroup{}

	// run srv
	httpSrv.Run(ctx, wg)

	// wait while services work
	wg.Wait()
	log.Info("Service stopped")
}

// initLogger initializes logger
func initLogger(logLevel string) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stderr)

	switch strings.ToLower(logLevel) {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.DebugLevel)
	}
}

// setupGracefulShutdown sets up graceful web server shutdown
func setupGracefulShutdown(stop func()) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		log.Error("Got Interrupt signal")
		stop()
	}()
}
