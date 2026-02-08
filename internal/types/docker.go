package types

import "time"

// DockerConfig for container monitoring
type DockerConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	SocketPath   string        `mapstructure:"socket_path"`
	PollInterval time.Duration `mapstructure:"poll_interval"`
	LabelPrefix  string        `mapstructure:"label_prefix"`
}
