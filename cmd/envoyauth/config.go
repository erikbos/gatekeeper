package main

import (
	"os"

	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
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
	Logger    shared.Logger            `yaml:"logging"`   // log configuration of application
	WebAdmin  webadmin.Config          `yaml:"webadmin"`  // Admin web interface configuration
	EnvoyAuth envoyAuthConfig          `yaml:"envoyauth"` // Envoyauth configuration
	OAuth     oauthServerConfig        `yaml:"oauth"`     // OAuth configuration
	Database  cassandra.DatabaseConfig `yaml:"database"`  // Database configuration
	Cache     cacheConfig              `yaml:"cache"`     // In-memory cache configuration
	Geoip     Geoip                    `yaml:"geoip"`     // Geoip lookup configuration
}

func loadConfiguration(filename *string) (*APIAuthConfig, error) {
	// default configuration
	config := APIAuthConfig{
		Logger: shared.Logger{
			Level:    defaultLogLevel,
			Filename: "/dev/stdout",
		},
		WebAdmin: webadmin.Config{
			Listen: defaultWebAdminListen,
			Logger: shared.Logger{
				Level:    defaultLogLevel,
				Filename: defaultWebAdminLogFileName,
			},
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
		return nil, err
	}
	defer file.Close()

	yamlDecoder := yaml.NewDecoder(file)
	yamlDecoder.SetStrict(true)
	if err := yamlDecoder.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
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
