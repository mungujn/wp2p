package config

import (
	"github.com/mungujn/web-exp/app"
	"github.com/mungujn/web-exp/server"
)

// Config houses all the configurations for the application
type Config struct {
	LogLevel          string        `mapstructure:"LOG_LEVEL" default:"DEBUG"`
	HTTPServer        server.Config `mapstructure:"HTTP_SERVER"`
	DistributedSystem app.Config    `mapstructure:"DISTRIBUTED_SYSTEM"`
}
