package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
)

const (
	defaultConfigFileName      = "envoycp-config.yaml"
	defaultLogLevel            = "info"
	defaultWebAdminListen      = "0.0.0.0:9902"
	defaultWebAdminLogFileName = "envoycp-admin.log"
	defaultXDSGRPCListen       = "0.0.0.0:9901"
	defaultXDSHTTPListen       = "0.0.0.0:9903"
)

// EnvoyCPConfig contains our startup configuration data
type EnvoyCPConfig struct {
	LogLevel string                   `yaml:"loglevel"`
	WebAdmin webAdminConfig           `yaml:"webadmin"`
	Database cassandra.DatabaseConfig `yaml:"database"`
	XDS      xdsConfig                `yaml:"xds"`
}

type xdsConfig struct {
	GRPCListen         string        `yaml:"grpclisten"`
	HTTPListen         string        `yaml:"httplisten"`
	ConfigPushInterval time.Duration `yaml:"configpushinterval"`
	Envoy              envoyConfig   `yaml:"envoy"`
}

type envoyConfig struct {
	ExtAuthz extAuthzConfig `yaml:"extauthz"`
	Logging  envoyLogConfig `yaml:"logging"`
}

type extAuthzConfig struct {
	Enabled          bool          `yaml:"enabled"`
	Cluster          string        `yaml:"cluster"`
	Timeout          time.Duration `yaml:"timeout"`
	FailureModeAllow bool          `yaml:"failuremodeallow"`
	RequestBodySize  int16         `yaml:"requestbodysize"`
}

type envoyLogConfig struct {
	File struct {
		LogFileName string            `yaml:"logfilename"`
		Fields      map[string]string `yaml:"fields"`
	} `yaml:"file"`
	GRPC struct {
		BufferSize uint32        `yaml:"buffersize"`
		Cluster    string        `yaml:"cluster"`
		LogName    string        `yaml:"logname"`
		Timeout    time.Duration `yaml:"timeout"`
	} `yaml:"grpc"`
}

const (
	defaultConfigPushInterval = 2 * time.Second
)

func loadConfiguration(filename *string) *EnvoyCPConfig {
	// default configuration
	config := EnvoyCPConfig{
		LogLevel: defaultLogLevel,
		WebAdmin: webAdminConfig{
			Listen:      defaultWebAdminListen,
			LogFileName: defaultWebAdminLogFileName,
		},
		XDS: xdsConfig{
			GRPCListen:         defaultXDSGRPCListen,
			HTTPListen:         defaultXDSHTTPListen,
			ConfigPushInterval: defaultConfigPushInterval,
		},
	}

	file, err := os.Open(*filename)
	if err != nil {
		log.Fatalf("Cannot load configuration file: %v", err)
	}
	defer file.Close()

	yamlDecoder := yaml.NewDecoder(file)
	yamlDecoder.SetStrict(true)
	if err := yamlDecoder.Decode(&config); err != nil {
		log.Fatalf("Cannot decode configuration file: %v", err)
	}

	return &config
}
