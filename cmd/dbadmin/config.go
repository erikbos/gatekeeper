package main

import (
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/service"
	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

const (
	defaultLogLevel            = "info"
	defaultWebAdminListen      = "0.0.0.0:7777"
	defaultWebAdminLogFileName = "dbadmin-access.log"
	defaultChangeLogFileName   = "dbadmin-changelog.log"
)

// DBAdminConfig contains our startup configuration data
type DBAdminConfig struct {
	Logger    shared.Logger            `yaml:"logging"`   // log configuration of application
	WebAdmin  webadmin.Config          `yaml:"webadmin"`  // Admin web interface configuration
	Changelog service.ChangelogConfig  `yaml:"changelog"` // Changelog configuration
	Database  cassandra.DatabaseConfig `yaml:"database"`  // Database configuration
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
			Filename: "/dev/stdout",
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
	}

	config, err := shared.LoadYAMLConfiguration(filename, defaultConfig)
	return config.(*DBAdminConfig), err
}
