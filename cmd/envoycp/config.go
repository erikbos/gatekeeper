package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/db/cassandra"
)

const (
	defaultLogLevel            = "info"
	defaultWebAdminListen      = "0.0.0.0:9902"
	defaultWebAdminLogFileName = "envoycp-admin.log"
	defaultXDSGRPCListen       = "0.0.0.0:9901"
	defaultXDSHTTPListen       = "0.0.0.0:9903"
)

// EnvoyCPConfig contains our startup configuration data
type EnvoyCPConfig struct {
	LogLevel   string                   `yaml:"loglevel"`   // Overall logging level of application
	WebAdmin   webAdminConfig           `yaml:"webadmin"`   // Admin web interface configuration
	Database   cassandra.DatabaseConfig `yaml:"database"`   // Database configuration
	XDS        xdsConfig                `yaml:"xds"`        // Control plane configuration
	Envoyproxy envoyProxyConfig         `yaml:"envoyproxy"` // Envoyproxy configuration
}

type xdsConfig struct {
	GRPCListen         string        `yaml:"grpclisten"`         // grpc listen port
	HTTPListen         string        `yaml:"httplisten"`         // http listen port
	ConfigPushInterval time.Duration `yaml:"configpushinterval"` // XDS Config push update interval to envoyproxy
}

type envoyProxyConfig struct {
	ExtAuthz extAuthzConfig `yaml:"extauthz"` // Extauthz configuration options
	Logging  envoyLogConfig `yaml:"logging"`  // Logging configuration options
}

type extAuthzConfig struct {
	Enable           bool          `yaml:"enable"`           // Enable/disable external authentication of requests
	Cluster          string        `yaml:"cluster"`          // Name of cluster to use for ext authz calls
	Timeout          time.Duration `yaml:"timeout"`          // Max duration of call to ext authz cluster
	FailureModeAllow bool          `yaml:"failuremodeallow"` // Should envoyproxy forward request in case extauthz cluster does not respond
	RequestBodySize  int16         `yaml:"requestbodysize"`  // Number of bytes to forward to exth authz cluster
}

type envoyLogConfig struct {
	File struct {
		LogFileName string            `yaml:"logfilename"` // Filename to write access logs to
		Fields      map[string]string `yaml:"fields"`      // Field names and field values to log
	} `yaml:"file"`
	GRPC struct {
		BufferSize uint32        `yaml:"buffersize"` // Buffersize for streaming access logs
		Cluster    string        `yaml:"cluster"`    // Cluster to stream access logs to
		LogName    string        `yaml:"logname"`    // Logname to use in acceas log
		Timeout    time.Duration `yaml:"timeout"`    // Timeout for connecting to access log cluster
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
