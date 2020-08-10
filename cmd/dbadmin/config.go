package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
)

const (
	defaultConfigFileName      = "dbadmin-config.yaml"
	defaultLogLevel            = "info"
	defaultWebAdminListen      = "0.0.0.0:7777"
	defaultWebAdminLogFileName = "dbadmin-access.log"
)

// DBAdminConfig contains our startup configuration data
type DBAdminConfig struct {
	LogLevel string                   `yaml:"loglevel" json:"loglevel"` // Overall logging level of application
	WebAdmin webAdminConfig           `yaml:"webadmin" json:"webadmin"` // Admin web interface configuration
	Database cassandra.DatabaseConfig `yaml:"database" json:"database"` // Database configuration
}

func loadConfiguration(filename *string) *DBAdminConfig {
	// default configuration
	config := DBAdminConfig{
		LogLevel: defaultLogLevel,
		WebAdmin: webAdminConfig{
			Listen:      defaultWebAdminListen,
			LogFileName: defaultWebAdminLogFileName,
		},
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
