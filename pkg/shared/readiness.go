package shared

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Readiness contains the rreadiness of the application
type Readiness struct {
	status bool
}

// DisplayStatus returns rreadiness status
func (r *Readiness) DisplayStatus(c *gin.Context) {
	if r.status {
		c.JSON(http.StatusOK, "Ready")
	} else {
		c.JSON(http.StatusServiceUnavailable, "Not ready")
	}
}

// FIX we should use failureThreshold and successThreshold before determining whether we are up or down

// Down changes status to down
func (r *Readiness) Down() {
	r.status = false

	log.Debugf("Setting readiness state to false")
}

// Up changes status to up
func (r *Readiness) Up() {
	r.status = true

	log.Debugf("Setting readiness state to true")
}
