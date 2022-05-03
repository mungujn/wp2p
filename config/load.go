package config

import (
	"github.com/mungujn/web-exp/config/reader"
)

func Read() (Config, error) {
	var cfg Config
	err := reader.Read(&cfg)
	return cfg, err
}
