package main

import (
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/service"
	"github.com/erikbos/gatekeeper/pkg/db/cache"
	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

const (
	defaultLogLevel            = "info"
	defaultLogFileName         = "/dev/stdout"
	defaultWebAdminListen      = "0.0.0.0:7777"
	defaultWebAdminLogFileName = "dbadmin-access.log"
	defaultChangeLogFileName   = "dbadmin-changelog.log"
	defaultCacheSize           = 100 * 1024 * 1024
	defaultCacheTTL            = 30
	defaultCacheNegativeTTL    = 5
	defaultOrganizationName    = "default"
)

// DBAdminConfig contains our startup configuration data
type DBAdminConfig struct {
	Logger       shared.Logger            `yaml:"logging"`      // log configuration of application
	WebAdmin     webadmin.Config          `yaml:"webadmin"`     // Admin web interface configuration
	Changelog    service.ChangelogConfig  `yaml:"changelog"`    // Changelog configuration
	Database     cassandra.DatabaseConfig `yaml:"database"`     // Database configuration
	Cache        cache.Config             `yaml:"cache"`        // Cache configuration
	Organization string                   `yaml:"organization"` // Organization name
}

// String() return our startup configuration as YAML
func (config *DBAdminConfig) String() string {

	// We must remove db password from configuration struct before showing
	redactedConfig := config
	redactedConfig.Database.Password = "[redacted]"

	configAsYAML, err := yaml.Marshal(redactedConfig)
	if err != nil {
		return ""
	}
	return string(configAsYAML)
}

func loadConfiguration(filename *string) (*DBAdminConfig, error) {

	defaultConfig := &DBAdminConfig{
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
		Changelog: service.ChangelogConfig{
			Logger: shared.Logger{
				Level:    defaultLogLevel,
				Filename: defaultChangeLogFileName,
			},
		},
		Cache: cache.Config{
			Size:        defaultCacheSize,
			TTL:         defaultCacheTTL,
			NegativeTTL: defaultCacheNegativeTTL,
		},
		Organization: defaultOrganizationName,
	}

	config, err := shared.LoadYAMLConfiguration(filename, defaultConfig)
	if err != nil {
		return nil, err
	}
	return config.(*DBAdminConfig), nil
}
