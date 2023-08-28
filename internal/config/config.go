package config

import "github.com/RacoonMediaServer/rms-packages/pkg/configuration"

type Bot struct {
	Token string
}

type Layout struct {
	Directory string
	Downloads string
}

type Remote struct {
	Scheme string
	Host   string
	Port   int
	Path   string
}

// Configuration represents entire service configuration
type Configuration struct {
	Database configuration.Database
	Bot      Bot
	Layout   Layout
	Remote   Remote
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
