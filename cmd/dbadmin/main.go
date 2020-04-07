package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/dchest/uniuri"
	"github.com/erikbos/apiauth/pkg/db"
	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func main() {
	configFilename := flag.String("configfilename", "dbadmin-config.yaml", "Configuration filename")
	flag.Parse()

	var config = RESTAPIConfig{}
	config.loadConfiguration(*configFilename)
	// FIXME we should check if we have all required parameters (use viper package?)

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000000",
		FullTimestamp:   true,
		DisableColors:   true,
	})
	log.SetLevel(log.DebugLevel)

	db, err := db.Connect(config.DatabaseHostname, config.DatabasePort, config.DatabaseUsername, config.DatabasePassword, config.DatabaseKeyspace)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
		os.Exit(1)
	}

	startRESTAPIServer(config.RESTAPIListen, db)
}

//RESTAPIConfig contains our startup configuration data
//
type RESTAPIConfig struct {
	RESTAPIListen    string `yaml:"dbadmin_admin_listen"`
	DatabaseHostname string `yaml:"database_hostname"`
	DatabasePort     int    `yaml:"database_port"`
	DatabaseUsername string `yaml:"database_username"`
	DatabasePassword string `yaml:"database_password"`
	DatabaseKeyspace string `yaml:"database_keyspace"`
	// CacheSize        int    `yaml:"cache_size"`
	// CacheTTL         int    `yaml:"cache_ttl"`
	// CacheNegativeTTL int    `yaml:"cache_negative_ttl"`
}

func (c *RESTAPIConfig) loadConfiguration(filename string) *RESTAPIConfig {
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

type env struct {
	db     *db.Database
	router *gin.Engine
}

func startRESTAPIServer(listenport string, db *db.Database) {
	// Store database handle so we can use it when answering apicalls

	e := &env{}
	e.db = db

	e.router = gin.New()

	// r.Use(gin.Logger())
	e.router.Use(gin.LoggerWithFormatter(logRequstparam))

	e.registerOrganizationRoutes(e.router)
	e.registerDeveloperRoutes(e.router)
	e.registerDeveloperAppRoutes(e.router)
	e.registerCredentialRoutes(e.router)
	e.registerAPIProductRoutes(e.router)
	e.registerClusterRoutes(e.router)

	e.router.Static("/assets", "./assets")
	e.router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	e.router.GET("/dump_routes", e.dumpRoutes)
	e.router.GET("/ready", e.GetReady)

	e.router.Run(listenport)
}

func logRequstparam(param gin.LogFormatterParams) string {
	return fmt.Sprintf("%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\"\n",
		param.ClientIP,
		param.TimeStamp.Format(time.RFC3339),
		param.Method,
		param.Path,
		param.Request.Proto,
		param.StatusCode,
		param.Latency/time.Millisecond,
		param.Request.UserAgent(),
		param.ErrorMessage,
	)
}

// boiler plate for later log actual API user
func (e *env) whoAmI() string {
	return "rest-api@test"
}

// restGetReady returns ready as readyness check
func (e *env) GetReady(c *gin.Context) {
	e.returnJSONMessage(c, http.StatusOK, errors.New("Ready"))
}

// restGetReady returns ready as readyness check
func (e *env) dumpRoutes(c *gin.Context) {
	routes := e.router.Routes()
	for _, v := range routes {
		log.Printf("%+v", v)
	}
	// c.IndentedJSON(http.StatusOK, gin.H{"routes": routes})
}

// CheckForJSONContentType checks for json content-type
func (e *env) CheckForJSONContentType(c *gin.Context) {
	if c.Request.Header.Get("content-type") != "application/json" {
		e.returnJSONMessage(c, http.StatusUnsupportedMediaType,
			errors.New("Content-type application/json required when submitting data"))
		// do not continue request handling
		c.Abort()
	}
}

func (e *env) SetLastModifiedHeader(c *gin.Context, timeStamp int64) {
	c.Header("Last-Modified",
		time.Unix(0, timeStamp*int64(time.Millisecond)).UTC().Format(http.TimeFormat))
}

// returnJSONMessage returns an error message in case we do not handle API request
func (e *env) returnJSONMessage(c *gin.Context, statusCode int, errorMessage error) {
	c.IndentedJSON(statusCode, gin.H{"message": fmt.Sprintf("%s", errorMessage)})
}

// getCurrentTimeMilliseconds returns current epoch time in milliseconds
func (e *env) getCurrentTimeMilliseconds() int64 {
	return time.Now().UTC().UnixNano() / 1000000
}

// findAttributePositionInAttributeArray find attribute in slice
func (e *env) findAttributePositionInAttributeArray(attributes []types.AttributeKeyValues, name string) int {
	for index, element := range attributes {
		if element.Name == name {
			return index
		}
	}
	return -1
}

// removeDuplicateAttributes removes duplicate attributes from array.
func (e *env) removeDuplicateAttributes(attributes []types.AttributeKeyValues) []types.AttributeKeyValues {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []types.AttributeKeyValues{}

	for v := range attributes {
		if encountered[attributes[v].Name] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[attributes[v].Name] = true
			// Append to result slice.
			result = append(result, attributes[v])
		}
	}
	return result
}

// GeneratePrimaryKeyOfDeveloper creates unique primary key for developer db row
func (e *env) GeneratePrimaryKeyOfDeveloper(organization, developer string) string {
	return (fmt.Sprintf("%s@@@%s", organization, uniuri.New()))
}

// GenerateDeveloperAppPrimaryKey creates unique primary key for developer app row
func (e *env) GenerateDeveloperAppPrimaryKey() string {
	return (fmt.Sprintf("%s", uuid.New()))
}

// GeneratePrimaryKeyOfDeveloper creates unique primary key for developer db row
func (e *env) GenerateDeveloperAppID(organization, primaryKey string) string {
	return (fmt.Sprintf("%s@@@%s", organization, primaryKey))
}

// GeneratePrimaryKeyOfAPIProduct creates unique primary key for apiproduct row
func (e *env) GeneratePrimaryKeyOfAPIProduct(organization, name string) string {
	return (fmt.Sprintf("%s@@@%s", organization, name))
}

// GenerateCredentialConsumerKey returns a random string to be used as apikey (32 character base62)
func (e *env) GenerateCredentialConsumerKey() string {
	return uniuri.NewLen(32)
}

// GenerateCredentialConsumerSecret returns a random string to be used as consumer key (16 character base62)
func (e *env) GenerateCredentialConsumerSecret() string {
	return uniuri.New()
}
