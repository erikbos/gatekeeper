package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/erikbos/apiauth/pkg/db"
)

const (
	defaultConfigFilename  = "envoyauth-config.yaml"
	defaultLogLevel        = "info"
	defaultWebAdminListen  = "0.0.0.0:7777"
	defaultWebAdminLogFile = "envoyauth-admin.log"
	defaultAuthGRPCListen  = "0.0.0.0:7778"
)

// APIAuthConfig contains our startup configuration data
type APIAuthConfig struct {
	LogLevel       string         `yaml:"loglevel"`
	WebAdmin       webAdminConfig `yaml:"webadmin"`
	AuthGRPCListen string         `yaml:"authgrpclisten"`
	Database       db.DatabaseConfig
	Cache          struct {
		Size        int `yaml:"size"`
		TTL         int `yaml:"ttl"`
		NegativeTTL int `yaml:"negativettl"`
	} `yaml:"cache"`
	Geoip struct {
		Filename string `yaml:"filename"`
	} `yaml:"geoip"`
}

func loadConfiguration() *APIAuthConfig {
	filename := flag.String("config", defaultConfigFilename, "Configuration filename")
	flag.Parse()

	// default configuration
	config := APIAuthConfig{
		LogLevel: defaultLogLevel,
		WebAdmin: webAdminConfig{
			Listen:  defaultWebAdminListen,
			LogFile: defaultWebAdminLogFile,
		},
		AuthGRPCListen: defaultAuthGRPCListen,
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
