package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"

	"os"
)

//EnvoyCPConfig contains our startup configuration data
type EnvoyCPConfig struct {
	GRPCXDSListen    string `yaml:"envoy_control_plane_gprc_listen"`
	HTTPXDSListen    string `yaml:"envoy_control_plane_http_listen"`
	DatabaseHostname string `yaml:"database_hostname"`
	DatabasePort     int    `yaml:"database_port"`
	DatabaseUsername string `yaml:"database_username"`
	DatabasePassword string `yaml:"database_password"`
	DatabaseKeyspace string `yaml:"database_keyspace"`
}

func loadConfiguration(filename string) *EnvoyCPConfig {
	var c EnvoyCPConfig
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("Cannot load configuration file: #%v", err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Could not parse configuration file contents: %v", err)
		os.Exit(1)
	}
	return &c
}
