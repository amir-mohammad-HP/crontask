package types

type Config struct {
	AppName     string         `mapstructure:"app_name"`
	Environment string         `mapstructure:"environment"`
	LogLevel    string         `mapstructure:"log_level"`
	Worker      WorkerConfig   `mapstructure:"worker"`
	Docker      DockerConfig   `mapstructure:"docker"`
	Shutdown    ShutdownConfig `mapstructure:"shutdown"`
	Logger      LoggerConfig   `mapstructure:"logger"`
}
