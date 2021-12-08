package main

import (
	"time"

	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

const (
	defaultLogLevel            = "info"
	defaultLogFileName         = "/dev/stdout"
	defaultWebAdminListen      = "0.0.0.0:9902"
	defaultWebAdminLogFileName = "controlplane-admin.log"
	defaultXDSGRPCListen       = "0.0.0.0:9901"
)

// ControlPlaneConfig contains our startup configuration data
type ControlPlaneConfig struct {
	Logger   shared.Logger            `yaml:"logging"`  // log configuration of application
	WebAdmin webadmin.Config          `yaml:"webadmin"` // Admin web interface configuration
	Database cassandra.DatabaseConfig `yaml:"database"` // Database configuration
	XDS      xdsConfig                `yaml:"xds"`      // Control plane configuration
}

const (
	defaultConfigCompileInterval = 2 * time.Second
)

// String() return our startup configuration as YAML
func (config *ControlPlaneConfig) String() string {

	// We must remove db password from configuration before showing
	redactedConfig := config
	redactedConfig.Database.Password = "[redacted]"

	configAsYAML, err := yaml.Marshal(redactedConfig)
	if err != nil {
		return ""
	}
	return string(configAsYAML)
}

func loadConfiguration(filename *string) (*ControlPlaneConfig, error) {

	defaultConfig := &ControlPlaneConfig{
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
		XDS: xdsConfig{
			Listen:                defaultXDSGRPCListen,
			ConfigCompileInterval: defaultConfigCompileInterval,
		},
	}

	config, err := shared.LoadYAMLConfiguration(filename, defaultConfig)
	if err != nil {
		return nil, err
	}
	return config.(*ControlPlaneConfig), nil
}
