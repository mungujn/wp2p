package system

import (
	"context"
	"github.com/mungujn/web-exp/config"
)

const (
	html       = "text/html"
	css        = "text/css"
	javascript = "text/javascript"
	plainText  = "text/plain"
)

type System struct {
	cfg config.Config
}

func New(ctx context.Context, cfg config.Config) (*System, error) {
	s := &System{cfg: cfg}
	return s, s.StartUp(ctx)
}

func (s *System) StartUp(ctx context.Context) error {
	return nil
}
