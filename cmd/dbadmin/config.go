package main

import (
	"flag"
	"os"

	"github.com/erikbos/apiauth/pkg/db"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFilename = "config/dbadmin-config.yaml"
	defaultLogLevel       = "info"
	defaultWebAdminListen = "0.0.0.0:7777"
)

//DBAdminConfig contains our startup configuration data
//
type DBAdminConfig struct {
	LogLevel       string `yaml:"loglevel"`
	WebAdminListen string `yaml:"webadminlisten"`
	Database       db.DatabaseConfig
}

func loadConfiguration() *DBAdminConfig {
	filename := flag.String("config", defaultConfigFilename, "Configuration filename")
	flag.Parse()

	// default configuration
	config := DBAdminConfig{
		LogLevel:       defaultLogLevel,
		WebAdminListen: defaultWebAdminListen,
	}

	file, err := os.Open(*filename)
	if err != nil {
		log.Fatalf("Cannot load configuration file: %v", err)
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		log.Fatalf("Cannot decode configuration file: %v", err)
	}
	return &config
}
