package system

import (
	"context"
)

// Config houses all the configurations for the distributed system
type Config struct {
	Username           string `mapstructure:"USERNAME"  default:"me"`
	LocalRootFolder    string `mapstructure:"LOCAL_ROOT_FOLDER"  default:"test_folder"`
	LocalWebServerPort int    `mapstructure:"LOCAL_WEB_SERVER_PORT"  default:"8080"`
	LocalNodeHost      string `mapstructure:"LOCAL_NODE_HOST"  default:"0.0.0.0"`
	LocalNodePort      int    `mapstructure:"LOCAL_NODE_PORT"  default:"4040"`
	NetworkName        string `mapstructure:"NETWORK_NAME"  default:"local"`
	ProtocolId         string `mapstructure:"PROTOCOL_ID"  default:"localfiles"`
	ProtocolVersion    string `mapstructure:"PROTOCOL_VERSION"  default:"0.1"`
}

// FileProvider specifies the interface that file service providers must meet
type FileProvider interface {
	StartHost(ctx context.Context) error
	GetFile(ctx context.Context, username, filename string) ([]byte, error)
	GetOnlineNodes() []string
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
