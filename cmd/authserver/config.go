package main

import (
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/erikbos/gatekeeper/cmd/authserver/oauth"
	"github.com/erikbos/gatekeeper/cmd/authserver/policy"
	"github.com/erikbos/gatekeeper/pkg/config"
	"github.com/erikbos/gatekeeper/pkg/db/cache"
	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

const (
	defaultLogLevel            = "info"
	defaultLogFileName         = "/dev/stdout"
	defaultWebAdminListen      = "0.0.0.0:7777"
	defaultWebAdminLogFileName = "authserver-admin.log"
	defaultExtAuthzListen      = "0.0.0.0:4000"
	defaultOAuthListen         = "0.0.0.0:4001"
	defaultCacheSize           = 100 * 1024 * 1024
	defaultCacheTTL            = 180
	defaultCacheNegativeTTL    = 5
	defaultOrganization        = "default"
)

// AuthServerConfig contains our startup configuration data
type AuthServerConfig struct {
	Logger    shared.Logger            // log configuration of application
	WebAdmin  webadmin.Config          // Admin web interface configuration
	EnvoyAuth envoyAuthConfig          // Authserver configuration
	OAuth     oauth.Config             // OAuth configuration
	Database  cassandra.DatabaseConfig // Database configuration
	Cache     cache.Config             // Cache configuration
	Geoip     policy.Geoip             // Geoip lookup configuration
}

func loadConfiguration(filename string) (*AuthServerConfig, error) {
	defaultConfig := &AuthServerConfig{
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
			Listen:              defaultExtAuthzListen,
			defaultOrganization: defaultOrganization,
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

	viper, err := config.Load(filename)
	if err != nil {
		return nil, err
	}

	return toAuthServerConfig(defaultConfig, *viper)
}

func toAuthServerConfig(defaultConfig *AuthServerConfig, v viper.Viper) (*AuthServerConfig, error) {
	err := v.Unmarshal(&defaultConfig)
	if err != nil {
		return nil, err
	}

	return defaultConfig, nil
}

// String() return our startup configuration as YAML
func (config *AuthServerConfig) String() string {
	// We must remove db password from configuration struct before showing
	redactedConfig := config
	redactedConfig.Database.Password = "[redacted]"

	configAsYAML, err := yaml.Marshal(redactedConfig)
	if err != nil {
		return ""
	}
	return string(configAsYAML)
}
