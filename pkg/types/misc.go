package types

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// returnJSONMessage returns an error message
func returnJSONMessage(c *gin.Context, statusCode int, errorMessage error) {
	c.IndentedJSON(statusCode, gin.H{"message": fmt.Sprintf("%s", errorMessage)})
}

// AbortIfContentTypeNotJSON checks for json content-type and abort request
func AbortIfContentTypeNotJSON(c *gin.Context) {
	if c.Request.Header.Get("content-type") != "application/json" {
		returnJSONMessage(c, http.StatusUnsupportedMediaType,
			errors.New("Content-type application/json required when submitting data"))
		// do not continue request handling
		c.Abort()
	}
}

// SetLoggingConfiguration sets logging format and level
func SetLoggingConfiguration(loglevel string) {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000000",
		FullTimestamp:   true,
		DisableColors:   true,
	})

	switch loglevel {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.Fatalf("Cannot set unknown loglevel %s", loglevel)
	}
	log.Info("Log level set to ", loglevel)
}
