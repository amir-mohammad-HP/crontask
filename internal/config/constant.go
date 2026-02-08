package config

var DEFAULT_CONFIG_YAML = `
# crontaskd Configuration File
# Environment: development, staging, production
# crontaskd.yaml
app_name: "CronTask"
environment: "production"
log_level: "info"

worker:
  interval: 10s
  max_jobs: 50
  retry_attempts: 5

shutdown:
  timeout: 60s

logger:
  level: "info"
  format: "json"  # or "text"
  output: "file"  # stdout, stderr, file, null
  file_path: "/var/log/crontaskd.log"  # Auto-detected if empty
  max_size: 100   # MB
  max_backups: 10
  max_age: 30     # days
  compress: true
  timestamp_format: "2006-01-02T15:04:05.000Z"
  show_caller: false
  colors: false   # No colors in production logs
  async: true
  buffer_size: 5000
`
