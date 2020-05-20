package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/erikbos/gatekeeper/pkg/db"
)

const (
	defaultConfigFilename  = "envoycp-config.yaml"
	defaultLogLevel        = "info"
	defaultWebAdminListen  = "0.0.0.0:9902"
	defaultWebAdminLogFile = "envoycp-admin.log"
	defaultXDSGRPCListen   = "0.0.0.0:9901"
	defaultXDSHTTPListen   = "0.0.0.0:9903"
)

// EnvoyCPConfig contains our startup configuration data
type EnvoyCPConfig struct {
	LogLevel string            `yaml:"loglevel"`
	WebAdmin webAdminConfig    `yaml:"webadmin"`
	Database db.DatabaseConfig `yaml:"database"`
	XDS      xdsConfig         `yaml:"xds"`
}

type xdsConfig struct {
	GRPCListen  string         `yaml:"xdsgrpclisten"`
	HTTPListen  string         `yaml:"xdshttplisten"`
	XDSInterval string         `yaml:"xdsinterval"`
	Envoy       envoyConfig    `yaml:"envoy"`
	ExtAuthz    extAuthzConfig `yaml:"extauthz"`
}

type envoyConfig struct {
	LogFilename string            `yaml:"logfilename"`
	LogFields   map[string]string `yaml:"logfields"`
}

type extAuthzConfig struct {
	Enabled          bool   `yaml:"enabled"`
	Cluster          string `yaml:"cluster"`
	Timeout          string `yaml:"timeout"`
	FailureModeAllow bool   `yaml:"failuremodeallow"`
	RequestBodySize  int16  `yaml:"requestbodysize"`
}

func loadConfiguration() *EnvoyCPConfig {
	filename := flag.String("config", defaultConfigFilename, "Configuration filename")
	flag.Parse()

	// default configuration
	config := EnvoyCPConfig{
		LogLevel: defaultLogLevel,
		WebAdmin: webAdminConfig{
			Listen:  defaultWebAdminListen,
			LogFile: defaultWebAdminLogFile,
		},
		XDS: xdsConfig{
			GRPCListen: defaultXDSGRPCListen,
			HTTPListen: defaultXDSHTTPListen,
		},
	}

	file, err := os.Open(*filename)
	if err != nil {
		log.Fatalf("Cannot load configuration file: %v", err)
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		log.Fatalf("Cannot decode configuration file: %v", err)
	}
	return &config
}
