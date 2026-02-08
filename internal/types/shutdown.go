package types

import "time"

type ShutdownConfig struct {
	Timeout time.Duration `mapstructure:"timeout"`
}
