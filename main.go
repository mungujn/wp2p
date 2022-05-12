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
	"github.com/mungujn/web-exp/remote"
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

	log.Info("service starting...")

	// prepare main context
	ctx, cancel := context.WithCancel(context.Background())
	setupGracefulShutdown(cancel)

	// init distributed capabilites provider
	dcfg := cfg.DistributedSystem
	provider := remote.New(dcfg)
	err = provider.StartHost(ctx)
	if err != nil {
		log.WithError(err).Fatal("host init error")
	}

	// init distributed system
	sys, err := system.New(ctx, dcfg, provider)
	if err != nil {
		log.WithError(err).Fatal("system init error")
	}

	// init http server
	httpSrv, err := serverHTTP.New(
		cfg.HTTPServer,
		sys,
	)
	if err != nil {
		log.WithError(err).Error("http server init error, http web server will not be available")
	}

	var wg = &sync.WaitGroup{}

	// run srv
	httpSrv.Run(ctx, wg)

	// wait while services work
	wg.Wait()
	log.Info("service stopped")
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
