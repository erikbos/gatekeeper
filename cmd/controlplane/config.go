package main

import (
	"time"

	"github.com/erikbos/gatekeeper/pkg/config"
	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
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
	Logger   shared.Logger            // log configuration of application
	WebAdmin webadmin.Config          // Admin web interface configuration
	Database cassandra.DatabaseConfig // Database configuration
	XDS      xdsConfig                // Control plane configuration
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

func loadConfiguration(filename string) (*ControlPlaneConfig, error) {

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

	viper, err := config.Load(filename)
	if err != nil {
		return nil, err
	}

	return toControlPlaneConfig(defaultConfig, *viper)
}

func toControlPlaneConfig(defaultConfig *ControlPlaneConfig, v viper.Viper) (*ControlPlaneConfig, error) {
	err := v.Unmarshal(&defaultConfig)
	if err != nil {
		return nil, err
	}

	return defaultConfig, nil
}
