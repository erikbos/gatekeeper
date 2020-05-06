package shared

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Readiness contains the rreadiness of the application
type Readiness struct {
	Status          bool
	LastStateChange time.Time
}

// DisplayStatus returns rreadiness status
func (r *Readiness) DisplayStatus(c *gin.Context) {
	if r.Status {
		c.IndentedJSON(http.StatusOK, gin.H{"readiness": r})
	} else {
		c.IndentedJSON(http.StatusServiceUnavailable, gin.H{"readiness": r})
	}
}

// Down changes status to down
func (r *Readiness) Down() {
	r.updateReadinessState(false)
}

// Up changes status to up
func (r *Readiness) Up() {
	r.updateReadinessState(true)
}

// TODO we should have a failureThreshold and successThreshold before determining whether we are up or down

// updateReadinessState set current readiness state if it has changed
func (r *Readiness) updateReadinessState(newState bool) {
	// In case timestamp is zero (=startup) we always set state
	if r.LastStateChange.IsZero() || r.Status != newState {
		r.Status = newState
		r.LastStateChange = time.Now().UTC()

		log.Infof("Setting readiness state to %t", r.Status)
	}
}
