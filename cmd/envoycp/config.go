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
	defaultWebAdminListen      = "0.0.0.0:9902"
	defaultWebAdminLogFileName = "envoycp-admin.log"
	defaultXDSGRPCListen       = "0.0.0.0:9901"
)

// EnvoyCPConfig contains our startup configuration data
type EnvoyCPConfig struct {
	Logger     shared.Logger            `yaml:"logging"`    // log configuration of application
	WebAdmin   webadmin.Config          `yaml:"webadmin"`   // Admin web interface configuration
	Database   cassandra.DatabaseConfig `yaml:"database"`   // Database configuration
	XDS        xdsConfig                `yaml:"xds"`        // Control plane configuration
	Envoyproxy envoyProxyConfig         `yaml:"envoyproxy"` // Envoyproxy configuration
}

type envoyProxyConfig struct {
	ExtAuthz    extAuthzConfig    `yaml:"extauthz"`    // Extauthz configuration options
	RateLimiter rateLimiterConfig `yaml:"ratelimiter"` // Ratelimiter configuration options
}

type extAuthzConfig struct {
	Enable           bool          `yaml:"enable"`           // Enable/disable external authentication of requests
	Cluster          string        `yaml:"cluster"`          // Name of cluster to use for ext authz calls
	Timeout          time.Duration `yaml:"timeout"`          // Max duration of call to ext authz cluster
	FailureModeAllow bool          `yaml:"failuremodeallow"` // Forward request in case extauthz does not respond in time
	RequestBodySize  int16         `yaml:"requestbodysize"`  // Number of bytes to forward to exth authz cluster
}

type rateLimiterConfig struct {
	Enable          bool          `yaml:"enable"`          // Enable/disable ratelimiting of requests
	Cluster         string        `yaml:"cluster"`         // Name of cluster to use for ratelimiter calls
	Timeout         time.Duration `yaml:"timeout"`         // Max duration of call to ext authz cluster
	FailureModeDeny bool          `yaml:"failuremodedeny"` // Forward request in case ratelimiting does not respond time
	Domain          string        `yaml:"domain"`
}

const (
	defaultConfigCompileInterval = 2 * time.Second
)

// String() return our startup configuration as YAML
func (config *EnvoyCPConfig) String() string {

	// We must remove db password from configuration before showing
	redactedConfig := config
	redactedConfig.Database.Password = "[redacted]"

	configAsYAML, err := yaml.Marshal(redactedConfig)
	if err != nil {
		return ""
	}
	return string(configAsYAML)
}

func loadConfiguration(filename *string) (*EnvoyCPConfig, error) {

	defaultConfig := &EnvoyCPConfig{
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
		XDS: xdsConfig{
			GRPCListen:            defaultXDSGRPCListen,
			ConfigCompileInterval: defaultConfigCompileInterval,
		},
	}

	config, err := shared.LoadYAMLConfiguration(filename, defaultConfig)
	return config.(*EnvoyCPConfig), err
}
