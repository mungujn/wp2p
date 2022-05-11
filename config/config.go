package config

import (
	"github.com/mungujn/web-exp/server"
	"github.com/mungujn/web-exp/system"
)

// Config houses all the configurations for the application
type Config struct {
	LogLevel          string        `mapstructure:"LOG_LEVEL" default:"DEBUG"`
	HTTPServer        server.Config `mapstructure:"HTTP_SERVER"`
	DistributedSystem system.Config `mapstructure:"DISTRIBUTED_SYSTEM"`
}
