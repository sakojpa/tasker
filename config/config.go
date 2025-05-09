package config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultEnvPrefix        = "TODO_"
	DefaultServerScheme     = "http"
	DefaultServerHost       = "0.0.0.0"
	DefaultServerPort       = "7540"
	DefaultServerTimeout    = 60 * time.Second
	DefaultServerStaticDir  = "web"
	DefaultDatabaseFilePath = "scheduler.db"
	DefaultAuthEnabled      = false
	DefailtAuthPassword     = "123"
)

var (
	DefaultEnvsFields = []string{"port", "dbfile"}
)

func newConfig() *Config {
	return &Config{
		Server: Server{
			Scheme:       DefaultServerScheme,
			Host:         DefaultServerHost,
			Port:         DefaultServerPort,
			WriteTimeout: DefaultServerTimeout,
			ReadTimeout:  DefaultServerTimeout,
			IdleTimeout:  DefaultServerTimeout,
			StaticDir:    DefaultServerStaticDir,
		},
		DB: DB{
			FilePath: DefaultDatabaseFilePath,
		},
		Auth: Auth{
			Enabled:  DefaultAuthEnabled,
			Password: DefailtAuthPassword,
		},
	}
}

// GetConfig loads or generates config based on file and environment variables.
func GetConfig(file string) (*Config, error) {
	c := newConfig()
	if strings.HasPrefix(file, "~/") {
		usr, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not get current user: %v", err)
		}
		file = filepath.Join(usr.HomeDir, file[2:])
	}
	_, err := os.Stat(file)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			c, err = NewDefaultConfig(file)
			if err != nil {
				return nil, fmt.Errorf("could not create default config for %s: %v", file, err)
			}
		}
	} else {
		cfg, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("could not read %s: %v", file, err)
		}
		if err := yaml.Unmarshal(cfg, &c); err != nil {
			return nil, fmt.Errorf("could not unmarshal %s: %v", file, err)
		}
	}
	envVars := loadVariablesFromEnv(DefaultEnvPrefix, DefaultEnvsFields)
	for k, v := range envVars {
		switch k {
		case "port":
			c.Server.Port = v
		case "dbfile":
			c.DB.FilePath = v
		default:
			continue
		}
	}
	return c, nil
}

// NewDefaultConfig creates a new default configuration file in YAML format.
func NewDefaultConfig(file string) (*Config, error) {
	c := newConfig()
	cfg, err := yaml.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("could not marshal config: %v", err)
	}
	f, err := os.Create(file)
	if err != nil {
		return nil, fmt.Errorf("could not create file: %v", err)
	}
	defer f.Close()
	_, err = f.Write(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not parse %s: %v", file, err)
	}
	return c, nil
}

func loadVariablesFromEnv(prefix string, envs []string) map[string]string {
	envVars := make(map[string]string)
	for _, field := range envs {
		key := prefix + strings.ToUpper(field)
		value := os.Getenv(key)
		if value != "" {
			envVars[field] = value
		}
	}
	return envVars
}
