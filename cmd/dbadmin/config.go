package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/db"
)

const (
	defaultConfigFilename  = "/config/dbadmin-config.yaml"
	defaultLogLevel        = "info"
	defaultWebAdminListen  = "0.0.0.0:7777"
	defaultWebAdminLogFile = "dbadmin-access.log"
)

//DBAdminConfig contains our startup configuration data
//
type DBAdminConfig struct {
	LogLevel string         `yaml:"loglevel"`
	WebAdmin webAdminConfig `yaml:"webadmin"`
	Database db.DatabaseConfig
}

func loadConfiguration() *DBAdminConfig {
	filename := flag.String("config", defaultConfigFilename, "Configuration filename")
	flag.Parse()

	// default configuration
	config := DBAdminConfig{
		LogLevel: defaultLogLevel,
		WebAdmin: webAdminConfig{
			Listen:  defaultWebAdminListen,
			LogFile: defaultWebAdminLogFile,
		},
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
