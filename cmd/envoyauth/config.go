package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFilename = "envoyauth-config.yaml"
	defaultLogLevel       = "info"
	defaultWebAdminListen = "0.0.0.0:7777"
	defaultAuthGRPCListen = "0.0.0.0:7778"
)

//APIAuthConfig contains our startup configuration data
//
type APIAuthConfig struct {
	LogLevel       string `yaml:"loglevel"`
	WebAdminListen string `yaml:"webadminlisten"`
	AuthGRPCListen string `yaml:"authgrpclisten"`
	Database       struct {
		Hostname string `yaml:"hostname"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Keyspace string `yaml:"keyspace"`
	} `yaml:"database"`
	Cache struct {
		Size        int `yaml:"size"`
		TTL         int `yaml:"ttl"`
		NegativeTTL int `yaml:"negativettl"`
	} `yaml:"cache"`
	Geoip struct {
		Filename string `yaml:"filename"`
	} `yaml:"geoip"`
}

func loadConfiguration() *APIAuthConfig {
	filename := flag.String("config", defaultConfigFilename, "Configuration filename")
	flag.Parse()

	// default configuration
	config := APIAuthConfig{
		LogLevel:       defaultLogLevel,
		WebAdminListen: defaultWebAdminListen,
		AuthGRPCListen: defaultAuthGRPCListen,
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
