package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/erikbos/apiauth/pkg/db"
	"github.com/erikbos/apiauth/pkg/geoip"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func main() {
	a := authorizationServer{}

	configFilename := flag.String("configfilename", "apiauth-config.yaml", "Configuration filename")
	flag.Parse()
	a.config.loadConfiguration(*configFilename)
	// FIXME we should check if we have all required parameters (use viper package?)

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000000",
		FullTimestamp:   true,
		DisableColors:   true,
	})
	log.SetLevel(log.DebugLevel)

	var err error
	a.db, err = db.Connect(a.config.DatabaseHostname, a.config.DatabasePort,
		a.config.DatabaseUsername, a.config.DatabasePassword, a.config.DatabaseKeyspace)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	a.c = db.CacheInit(a.config.CacheSize, a.config.CacheTTL, a.config.CacheNegativeTTL)

	a.g, err = geoip.OpenDatabase(a.config.MaxMindFilename)
	if err != nil {
		log.Fatalf("Geoip db load failed: %v", err)
	}

	StartWebAdminServer(a)
	startGRPCAuthenticationServer(a)
}

//APIAuthConfig contains our startup configuration data
//
type APIAuthConfig struct {
	GRPCAuthListen   string `yaml:"auth_gprc_listen"`
	WebAdminListen   string `yaml:"auth_admin_listen"`
	DatabaseHostname string `yaml:"database_hostname"`
	DatabasePort     int    `yaml:"database_port"`
	DatabaseUsername string `yaml:"database_username"`
	DatabasePassword string `yaml:"database_password"`
	DatabaseKeyspace string `yaml:"database_keyspace"`
	CacheSize        int    `yaml:"cache_size"`
	CacheTTL         int    `yaml:"cache_ttl"`
	CacheNegativeTTL int    `yaml:"cache_negative_ttl"`
	MaxMindFilename  string `yaml:"maxmind_filename"`
}

func (c *APIAuthConfig) loadConfiguration(filename string) *APIAuthConfig {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("Cannot load configuration file: #%v", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Could not parse configuration file contents: %v", err)
		os.Exit(1)
	}

	return c
}
