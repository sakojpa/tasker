package config

import (
	"time"
)

// Config represents application settings including server, database, and authentication configurations.
type Config struct {
	Server Server `yaml:"server" json:"server"`
	DB     DB     `yaml:"db" json:"db"`
	Auth   Auth   `yaml:"auth" json:"auth"`
}

// Server holds HTTP server configuration details such as host, port, timeouts, and static directory.
type Server struct {
	Scheme       string        `yaml:"server_scheme" json:"server_scheme"`
	Host         string        `yaml:"server_address" json:"server_address"`
	Port         string        `yaml:"server_port" json:"server_port"`
	ReadTimeout  time.Duration `yaml:"server_timeout_read" json:"server_timeout_read"`
	WriteTimeout time.Duration `yaml:"server_timeout_write" json:"server_timeout_write"`
	IdleTimeout  time.Duration `yaml:"server_timeout_idle" json:"server_timeout_idle"`
	StaticDir    string        `yaml:"static_dir" json:"static_dir"`
}

// DB stores the database file path configuration.
type DB struct {
	FilePath string `yaml:"database_filename" json:"database_filename"`
}

// Auth contains authentication settings like enablement and password.
type Auth struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Password string `yaml:"password" json:"password"`
}
