package system

import (
	"context"
)

const (
	html       = "text/html"
	css        = "text/css"
	javascript = "text/javascript"
	plainText  = "text/plain"
)

// Config houses all the configurations for the distributed system
type Config struct {
	Username        string `mapstructure:"USERNAME"  default:"current_folder"`
	LocalRootFolder string `mapstructure:"LOCAL_ROOT_FOLDER"  default:"test_folder"`
}

// FileProvider specifies the interface that file service providers must meet
type FileProvider interface {
	SetUp(ctx context.Context, cfg Config) error
	GetFile(ctx context.Context, username, filename string) ([]byte, error)
}

// System is the main implementation of the applications logic
type System struct {
	cfg          Config
	fileProvider FileProvider
}

// New returns a new instance of the system
func New(ctx context.Context, cfg Config, fileProvider FileProvider) (*System, error) {
	s := &System{cfg: cfg, fileProvider: fileProvider}
	return s, s.SetUp(ctx, cfg)
}

// SetUp sets up the system
func (s *System) SetUp(ctx context.Context, cfg Config) error {
	return s.fileProvider.SetUp(ctx, cfg)
}
