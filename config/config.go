package config

import (
	"github.com/mungujn/web-exp/server"
	"github.com/mungujn/web-exp/system"
)

// Config houses all the configurations for the application
type Config struct {
	LogLevel                string        `mapstructure:"LOG_LEVEL" default:"DEBUG"`
	HTTPServerCfg           server.Config `mapstructure:"HTTP_SERVER"`
	DistributedSystemConfig system.Config `mapstructure:"DISTRIBUTED_SYSTEM"`
}

