package main

import (
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/cmd/managementserver/audit"
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

// managementServerConfig contains our startup configuration data
type managementServerConfig struct {
	Logger   shared.Logger            `yaml:"logging"`  // log configuration of application
	WebAdmin webadmin.Config          `yaml:"webadmin"` // Admin web interface configuration
	Audit    audit.Config             `yaml:"audit"`    // Audit configuration
	Database cassandra.DatabaseConfig `yaml:"database"` // Database configuration
	Cache    cache.Config             `yaml:"cache"`    // Cache configuration
}

// String() return our startup configuration as YAML
func (config *managementServerConfig) String() string {

	// We must remove db password from configuration struct before showing
	redactedConfig := config
	redactedConfig.Database.Password = "[redacted]"

	configAsYAML, err := yaml.Marshal(redactedConfig)
	if err != nil {
		return ""
	}
	return string(configAsYAML)
}

func loadConfiguration(filename *string) (*managementServerConfig, error) {

	defaultConfig := &managementServerConfig{
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

	config, err := shared.LoadYAMLConfiguration(filename, defaultConfig)
	if err != nil {
		return nil, err
	}
	return config.(*managementServerConfig), nil
}
