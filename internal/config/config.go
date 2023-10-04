package config

import "github.com/RacoonMediaServer/rms-packages/pkg/configuration"

type Bot struct {
	Token string
}

type Layout struct {
	Directory string
	Downloads string
	Limit     uint // Gigabytes
}

type Remote struct {
	Scheme string
	Host   string
	Port   int
	Path   string
}

type HealthCheck struct {
	Enabled bool
	Fifo    string
}

type Service struct {
	Server   string
	Username string
	Password string
}

type UserControl struct {
	Enabled      bool
	DefaultToken string `json:"default-token"`
}

// Configuration represents entire service configuration
type Configuration struct {
	Database    configuration.Database
	Bot         Bot
	Layout      Layout
	Remote      Remote
	Service     Service
	HealthCheck HealthCheck
	UserControl UserControl `json:"user-control"`
}

var config Configuration

// Load open and parses configuration file
func Load(configFilePath string) error {
	return configuration.Load(configFilePath, &config)
}

// Config returns loaded configuration
func Config() Configuration {
	return config
}
