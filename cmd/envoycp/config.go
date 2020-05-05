package main

import (
	"flag"
	"os"

	"github.com/erikbos/apiauth/pkg/db"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFilename = "envoycp-config.yaml"
	defaultLogLevel       = "info"
	defaultWebAdminListen = "0.0.0.0:9902"
	defaultXDSGRPCListen  = "0.0.0.0:9901"
	defaultXDSHTTPListen  = "0.0.0.0:9903"
)

// EnvoyCPConfig contains our startup configuration data
type EnvoyCPConfig struct {
	LogLevel       string `yaml:"loglevel"`
	WebAdminListen string `yaml:"webadminlisten"`
	XDSGRPCListen  string `yaml:"xdsgrpclisten"`
	XDSHTTPListen  string `yaml:"xdshttplisten"`
	Database       db.DatabaseConfig
	EnvoyLogging   struct {
		Filename string            `yaml:"filename"`
		Fields   map[string]string `yaml:"fields"`
	} `yaml:"envoylogging"`
}

func loadConfiguration() *EnvoyCPConfig {
	filename := flag.String("config", defaultConfigFilename, "Configuration filename")
	flag.Parse()

	// default configuration
	config := EnvoyCPConfig{
		LogLevel:       defaultLogLevel,
		WebAdminListen: defaultWebAdminListen,
		XDSGRPCListen:  defaultXDSGRPCListen,
		XDSHTTPListen:  defaultXDSHTTPListen,
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
