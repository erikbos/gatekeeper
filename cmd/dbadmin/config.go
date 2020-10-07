package main

import (
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

const (
	defaultLogLevel            = "info"
	defaultWebAdminListen      = "0.0.0.0:7777"
	defaultWebAdminLogFileName = "dbadmin-access.log"
)

// DBAdminConfig contains our startup configuration data
type DBAdminConfig struct {
	Logger   shared.Logger            `yaml:"logging"`                  // log configuration of application
	WebAdmin webadmin.Config          `yaml:"webadmin" json:"webadmin"` // Admin web interface configuration
	Database cassandra.DatabaseConfig `yaml:"database" json:"database"` // Database configuration
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
	}

	config, err := shared.LoadYAMLConfiguration(filename, defaultConfig)
	return config.(*DBAdminConfig), err
}
