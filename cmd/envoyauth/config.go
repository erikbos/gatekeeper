package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
)

const (
	defaultLogLevel            = "info"
	defaultWebAdminListen      = "0.0.0.0:7777"
	defaultWebAdminLogFileName = "envoyauth-admin.log"
	defaultAuthGRPCListen      = "0.0.0.0:4000"
	defaultOAuthListen         = "0.0.0.0:4001"
)

// APIAuthConfig contains our startup configuration data
type APIAuthConfig struct {
	LogLevel  string                   `yaml:"loglevel"`  // Overall logging level of application
	WebAdmin  webAdminConfig           `yaml:"webadmin"`  // Admin web interface configuration
	EnvoyAuth envoyAuthConfig          `yaml:"envoyauth"` // Envoyauth configuration
	OAuth     oauthServerConfig        `yaml:"oauth"`     // OAuth configuration
	Database  cassandra.DatabaseConfig `yaml:"database"`  // Database configuration
	Cache     cacheConfig              `yaml:"cache"`     // In-memory cache configuration
	Geoip     Geoip                    `yaml:"geoip"`     // Geoip lookup configuration
}

func loadConfiguration(filename *string) *APIAuthConfig {
	// default configuration
	config := APIAuthConfig{
		LogLevel: defaultLogLevel,
		WebAdmin: webAdminConfig{
			Listen:  defaultWebAdminListen,
			LogFile: defaultWebAdminLogFileName,
		},
		EnvoyAuth: envoyAuthConfig{
			Listen: defaultAuthGRPCListen,
		},
		OAuth: oauthServerConfig{
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

// String() return our startup configuration as YAML
func (config *APIAuthConfig) String() string {

	// We must remove db password from configuration struct before showing
	redactedConfig := config
	redactedConfig.Database.Password = "[redacted]"

	configAsYAML, err := yaml.Marshal(redactedConfig)
	if err != nil {
		return ""
	}
	return string(configAsYAML)
}
