package main

import (
	"time"

	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/config"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
	"github.com/spf13/viper"
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

// accessLogServerConfig contains our startup configuration data
type accessLogServerConfig struct {
	Logger       shared.Logger         // log configuration of application
	WebAdmin     webadmin.Config       // Admin web interface configuration
	AccessLogger AccessLogServerConfig // Access logging configuration
}

// String() return our startup configuration as YAML
func (config *accessLogServerConfig) String() string {

	configAsYAML, err := yaml.Marshal(config)
	if err != nil {
		return ""
	}
	return string(configAsYAML)
}

func loadConfiguration(filename string) (*accessLogServerConfig, error) {

	defaultConfig := &accessLogServerConfig{
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

	viper, err := config.Load(filename)
	if err != nil {
		return nil, err
	}

	return toAccessLogServerConfig(defaultConfig, *viper)
}

func toAccessLogServerConfig(defaultConfig *accessLogServerConfig, v viper.Viper) (*accessLogServerConfig, error) {
	err := v.Unmarshal(&defaultConfig)
	if err != nil {
		return nil, err
	}

	return defaultConfig, nil
}
