package config

import (
	"github.com/mungujn/web-exp/server"
)

type Config struct {
	LogLevel      string        `mapstructure:"LOG_LEVEL" default:"DEBUG"`
	HTTPServerCfg server.Config `mapstructure:"HTTP_SERVER"`
	XFlags        XFlags        `mapstructure:"X"`
}

// XFlags experimental flags
type XFlags struct {
	DevelopmentDomains []string `mapstructure:"DEVELOPMENT_DOMAINS" default:"localhost"`
}
