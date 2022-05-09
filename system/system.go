package system

import (
	"context"
)

// Config houses all the configurations for the distributed system
type Config struct {
	Username        string `mapstructure:"USERNAME"  default:"me"`
	LocalRootFolder string `mapstructure:"LOCAL_ROOT_FOLDER"  default:"test_folder"`
	LocalServerPort int    `mapstructure:"LOCAL_SERVER_PORT"  default:"8080"`
}

// FileProvider specifies the interface that file service providers must meet
type FileProvider interface {
	GetFile(ctx context.Context, username, filename string) ([]byte, error)
	GetOnlineNodes(ctx context.Context) ([]string, error)
}

// System is the main implementation of the applications logic
type System struct {
	cfg          Config
	fileProvider FileProvider
}

// New returns a new instance of the system
func New(ctx context.Context, cfg Config, fileProvider FileProvider) (*System, error) {
	s := &System{cfg: cfg, fileProvider: fileProvider}
	return s, nil
}
