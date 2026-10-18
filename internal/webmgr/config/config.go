package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"

	"github.com/metatube-community/metatube-sdk-go/internal/webmgr/multiplexer"
)

type Config struct {
	App         AppConfig         `yaml:"app"`
	Metatube    MetatubeConfig    `yaml:"metatube"`
	Log         LogConfig         `yaml:"log"`
	Embedded    EmbeddedConfig    `yaml:"embedded"`
	External    ExternalConfig    `yaml:"external"`
	Multiplexer MultiplexerConfig `yaml:"multiplexer"`
}

type MultiplexerConfig struct {
	MovieSystems []multiplexer.System `yaml:"movie_systems"`
}

type LogConfig struct {
	Filename  string `yaml:"filename"`
	MaxSizeMB int    `yaml:"max_size_mb"`
	MaxFiles  int    `yaml:"max_files"`
}

type AppConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	DataDir string `yaml:"data_dir"`
}

type MetatubeConfig struct {
	Mode          string `yaml:"mode"`
	Token         string `yaml:"token"`
	MovieProvider string `yaml:"movie_provider"`
	APIHost       string `yaml:"api_host"`
	APIPort       int    `yaml:"api_port"`
}

type ExternalConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type EmbeddedConfig struct {
	Port int `yaml:"port"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.applyEnvOverrides()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) applyEnvOverrides() {
	if val := os.Getenv("APP_HOST"); val != "" {
		c.App.Host = val
	}
	if val := os.Getenv("APP_PORT"); val != "" {
		fmt.Sscanf(val, "%d", &c.App.Port)
	}
	if val := os.Getenv("APP_DATA_DIR"); val != "" {
		c.App.DataDir = val
	}
	if val := os.Getenv("METATUBE_MODE"); val != "" {
		c.Metatube.Mode = val
	}
	if val := os.Getenv("METATUBE_TOKEN"); val != "" {
		c.Metatube.Token = val
	}
	if val := os.Getenv("METATUBE_MOVIE_PROVIDER"); val != "" {
		c.Metatube.MovieProvider = val
	}
	if val := os.Getenv("METATUBE_API_HOST"); val != "" {
		c.Metatube.APIHost = val
	}
	if val := os.Getenv("METATUBE_API_PORT"); val != "" {
		fmt.Sscanf(val, "%d", &c.Metatube.APIPort)
	}
	if val := os.Getenv("METATUBE_HOST"); val != "" {
		c.External.Host = val
	}
	if val := os.Getenv("METATUBE_PORT"); val != "" {
		fmt.Sscanf(val, "%d", &c.External.Port)
	}
	if val := os.Getenv("LOG_FILENAME"); val != "" {
		c.Log.Filename = val
	}
	if val := os.Getenv("LOG_MAX_SIZE_MB"); val != "" {
		fmt.Sscanf(val, "%d", &c.Log.MaxSizeMB)
	}
	if val := os.Getenv("LOG_MAX_FILES"); val != "" {
		fmt.Sscanf(val, "%d", &c.Log.MaxFiles)
	}
}

func (c *Config) validate() error {
	if c.App.Port <= 0 || c.App.Port > 65535 {
		return fmt.Errorf("invalid app port: %d", c.App.Port)
	}
	if c.App.DataDir == "" {
		c.App.DataDir = "./data"
	}
	if c.Log.Filename == "" {
		c.Log.Filename = "webmgr.log"
	}
	if c.Log.MaxSizeMB <= 0 {
		c.Log.MaxSizeMB = 16
	}
	if c.Log.MaxFiles <= 0 {
		c.Log.MaxFiles = 5
	}
	if c.Metatube.Mode == "" {
		c.Metatube.Mode = "embedded"
	}
	if c.Metatube.MovieProvider == "" {
		c.Metatube.MovieProvider = "JavBus"
	}
	if c.Metatube.APIHost == "" {
		c.Metatube.APIHost = "0.0.0.0"
	}
	if c.Metatube.APIPort <= 0 || c.Metatube.APIPort > 65535 {
		c.Metatube.APIPort = 8080
	}
	if c.Metatube.Mode != "embedded" && c.Metatube.Mode != "external" {
		return fmt.Errorf("invalid metatube mode: %s (must be 'embedded' or 'external')", c.Metatube.Mode)
	}
	return nil
}

func (c *Config) GetMetatubeBinaryPath() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	binaryName := fmt.Sprintf("metatube-server-%s-%s", goos, goarch)
	if goos == "windows" {
		binaryName += ".exe"
	}

	return filepath.Join("bin", binaryName)
}

func (c *Config) GetDBPath() string {
	return filepath.Join(c.App.DataDir, "data.db")
}

func (c *Config) GetMetatubeURL() string {
	if c.Metatube.Mode == "external" {
		return fmt.Sprintf("http://%s:%d", c.External.Host, c.External.Port)
	}
	return fmt.Sprintf("http://localhost:%d", c.Embedded.Port)
}
