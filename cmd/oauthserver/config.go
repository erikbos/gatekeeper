package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/db"
)

const (
	defaultConfigFilename  = "oauthserver-config.yaml"
	defaultLogLevel        = "info"
	defaultWebAdminListen  = "0.0.0.0:1000"
	defaultWebAdminLogFile = "oauthserver-access.log"
)

// OAuthServerConfig contains configuration data
type OAuthServerConfig struct {
	LogLevel string            `yaml:"loglevel" json:"loglevel"`
	Database db.DatabaseConfig `yaml:"database" json:"database"`
}

func loadConfiguration() *OAuthServerConfig {
	filename := flag.String("config", defaultConfigFilename, "Configuration filename")
	flag.Parse()

	config := OAuthServerConfig{
		LogLevel: defaultLogLevel,
	}

	file, err := os.Open(*filename)
	if err != nil {
		log.Fatalf("Cannot load configuration file: %v", err)
	}
	defer file.Close()

	yamlDecoder := yaml.NewDecoder(file)
	yamlDecoder.SetStrict(true)
	if err := yamlDecoder.Decode(&config); err != nil {
		log.Fatalf("Cannot decode configuration file: %v", err)
	}

	return &config
}
