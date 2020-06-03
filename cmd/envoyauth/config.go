package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/db"
)

const (
	defaultConfigFilename  = "envoyauth-config.yaml"
	defaultLogLevel        = "info"
	defaultWebAdminListen  = "0.0.0.0:7777"
	defaultWebAdminLogFile = "envoyauth-admin.log"
	defaultAuthGRPCListen  = "0.0.0.0:4000"
	defaultOAuthListen     = "0.0.0.0:4001"
)

// APIAuthConfig contains our startup configuration data
type APIAuthConfig struct {
	LogLevel       string            `yaml:"loglevel"`
	WebAdmin       webAdminConfig    `yaml:"webadmin"`
	AuthGRPCListen string            `yaml:"authgrpclisten"`
	Oauth          oauthServerConfig `yaml:"oauth"`
	Database       db.DatabaseConfig `yaml:"database"`
	Cache          CacheConfig       `yaml:"cache"`
	Geoip          struct {
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
		Oauth: oauthServerConfig{
			Listen: defaultOAuthListen,
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
