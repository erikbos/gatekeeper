package main

import (
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/cmd/envoyauth/oauth"
	"github.com/erikbos/gatekeeper/cmd/envoyauth/policy"
	"github.com/erikbos/gatekeeper/pkg/db/cache"
	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

const (
	defaultLogLevel            = "info"
	defaultLogFileName         = "/dev/stdout"
	defaultWebAdminListen      = "0.0.0.0:7777"
	defaultWebAdminLogFileName = "envoyauth-admin.log"
	defaultExtAuthzListen      = "0.0.0.0:4000"
	defaultOAuthListen         = "0.0.0.0:4001"
	defaultCacheSize           = 100 * 1024 * 1024
	defaultCacheTTL            = 180
	defaultCacheNegativeTTL    = 5
)

// EnvoyAuthConfig contains our startup configuration data
type EnvoyAuthConfig struct {
	Logger    shared.Logger            `yaml:"logging"`   // log configuration of application
	WebAdmin  webadmin.Config          `yaml:"webadmin"`  // Admin web interface configuration
	EnvoyAuth envoyAuthConfig          `yaml:"envoyauth"` // Envoyauth configuration
	OAuth     oauth.Config             `yaml:"oauth"`     // OAuth configuration
	Database  cassandra.DatabaseConfig `yaml:"database"`  // Database configuration
	Cache     cache.Config             `yaml:"cache"`     // Cache configuration
	Geoip     policy.Geoip             `yaml:"geoip"`     // Geoip lookup configuration
}

func loadConfiguration(filename *string) (*EnvoyAuthConfig, error) {

	defaultConfig := &EnvoyAuthConfig{
		Logger: shared.Logger{
			Level:    defaultLogLevel,
			Filename: defaultLogFileName,
		},
		WebAdmin: webadmin.Config{
			Listen: defaultWebAdminListen,
			Logger: shared.Logger{
				Level:    defaultLogLevel,
				Filename: defaultWebAdminLogFileName,
			},
		},
		EnvoyAuth: envoyAuthConfig{
			Listen: defaultExtAuthzListen,
		},
		OAuth: oauth.Config{
			Listen: defaultOAuthListen,
		},
		Cache: cache.Config{
			Size:        defaultCacheSize,
			TTL:         defaultCacheTTL,
			NegativeTTL: defaultCacheNegativeTTL,
		},
	}

	config, err := shared.LoadYAMLConfiguration(filename, defaultConfig)
	if err != nil {
		return nil, err
	}
	return config.(*EnvoyAuthConfig), nil
}

// String() return our startup configuration as YAML
func (config *EnvoyAuthConfig) String() string {

	// We must remove db password from configuration struct before showing
	redactedConfig := config
	redactedConfig.Database.Password = "[redacted]"

	configAsYAML, err := yaml.Marshal(redactedConfig)
	if err != nil {
		return ""
	}
	return string(configAsYAML)
}
