package main

import (
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/erikbos/gatekeeper/cmd/managementserver/audit"
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
	defaultWebAdminLogFileName = "managementserver-access.log"
	defaultAuditLogFileName    = "managementserver-audit.log"
	defaultCacheSize           = 100 * 1024 * 1024
	defaultCacheTTL            = 30
	defaultCacheNegativeTTL    = 5
)

// ManagementServerConfig contains our startup configuration data
type ManagementServerConfig struct {
	Logger   shared.Logger            // log configuration of application
	WebAdmin webadmin.Config          // Admin web interface configuration
	Audit    audit.Config             // Audit configuration
	Database cassandra.DatabaseConfig // Database configuration
	Cache    cache.Config             // Cache configuration
}

// String() return our startup configuration as YAML
func (config *ManagementServerConfig) String() string {

	// We must remove db password from configuration struct before showing
	redactedConfig := config
	redactedConfig.Database.Password = "[redacted]"

	configAsYAML, err := yaml.Marshal(redactedConfig)
	if err != nil {
		return ""
	}
	return string(configAsYAML)
}

func loadConfiguration(filename string) (*ManagementServerConfig, error) {

	defaultConfig := &ManagementServerConfig{
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
		Audit: audit.Config{
			Logger: shared.Logger{
				Level:    defaultLogLevel,
				Filename: defaultAuditLogFileName,
			},
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

	return toManagementServerConfig(defaultConfig, *viper)
}

func toManagementServerConfig(defaultConfig *ManagementServerConfig, v viper.Viper) (*ManagementServerConfig, error) {
	err := v.Unmarshal(&defaultConfig)
	if err != nil {
		return nil, err
	}

	return defaultConfig, nil
}
