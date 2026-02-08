package types

// Config holds logger configuration
type LoggerConfig struct {
	Level           string `mapstructure:"level"`            // Log level: debug, info, warn, error, fatal
	Format          string `mapstructure:"format"`           // Output format: text, json
	Output          string `mapstructure:"output"`           // Output: stdout, stderr, file, syslog
	FilePath        string `mapstructure:"file_path"`        // File path for file output
	MaxSize         int    `mapstructure:"max_size"`         // Max file size in MB for rotation
	MaxBackups      int    `mapstructure:"max_backups"`      // Max number of old log files
	MaxAge          int    `mapstructure:"max_age"`          // Max age in days
	Compress        bool   `mapstructure:"compress"`         // Compress rotated files
	TimestampFormat string `mapstructure:"timestamp_format"` // Time format
	ShowCaller      bool   `mapstructure:"show_caller"`      // Show caller information
	Colors          bool   `mapstructure:"colors"`           // Enable colors in console
	Async           bool   `mapstructure:"async"`            // Async logging
	BufferSize      int    `mapstructure:"buffer_size"`      // Buffer size for async logging
}
