package main

import (
	"time"

	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

const (
	defaultLogLevel             = "info"
	defaultLogFileName          = "/dev/stdout"
	defaultWebAdminListen       = "0.0.0.0:8888"
	defaultWebAdminLogFileName  = "accesslogserver-admin.log"
	defaultALSListen            = "0.0.0.0:8001"
	defaultALSLogFileName       = "envoyproxy.log"
	defaultALSMaxStreamDuration = 1 * time.Hour
)

// accesslogserverConfig contains our startup configuration data
type accesslogserverConfig struct {
	Logger       shared.Logger         `yaml:"logging"`   // log configuration of application
	WebAdmin     webadmin.Config       `yaml:"webadmin"`  // Admin web interface configuration
	AccessLogger AccessLogServerConfig `yaml:"accesslog"` // Access logging configuration
}

// String() return our startup configuration as YAML
func (config *accesslogserverConfig) String() string {

	configAsYAML, err := yaml.Marshal(config)
	if err != nil {
		return ""
	}
	return string(configAsYAML)
}

func loadConfiguration(filename *string) (*accesslogserverConfig, error) {

	defaultConfig := &accesslogserverConfig{
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
		AccessLogger: AccessLogServerConfig{
			Listen:            defaultALSListen,
			MaxStreamDuration: defaultALSMaxStreamDuration,
			Logger: shared.Logger{
				Level:    defaultLogLevel,
				Filename: defaultALSLogFileName,
			},
		},
	}

	config, err := shared.LoadYAMLConfiguration(filename, defaultConfig)
	if err != nil {
		return nil, err
	}
	return config.(*accesslogserverConfig), nil
}
