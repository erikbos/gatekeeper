package shared

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// SetLoggingConfiguration sets logging format and level
func SetLoggingConfiguration(loglevel string) {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000000",
		FullTimestamp:   true,
		DisableColors:   true,
	})
	level, err := log.ParseLevel(loglevel)
	if err != nil {
		log.Fatalf("Cannot set unknown loglevel %s", loglevel)
	}
	log.SetLevel(level)
	log.Info("Log level set to ", loglevel)
}

// LogHTTPRequest logs details of an HTTP request
func LogHTTPRequest(param gin.LogFormatterParams) string {
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
