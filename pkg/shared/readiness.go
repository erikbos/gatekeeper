package shared

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Readiness contains the rreadiness of the application
type Readiness struct {
	Status            bool      `json:"status"`
	LastStateChange   time.Time `json:"lastStateChange"`
	transitionCounter *prometheus.CounterVec
}

// DisplayStatus shows our readiness status
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
	// In case timestamp is zero (= @ startup) we always set status
	if r.LastStateChange.IsZero() || r.Status != newState {
		r.Status = newState
		r.LastStateChange = time.Now().UTC()
		r.increaseTransitionMetric(r.Status)

		log.Infof("Setting readiness state to %t", r.Status)
	}
}

// RegisterMetrics registers our readiness metric
func (r *Readiness) RegisterMetrics(serviceName string) {
	r.transitionCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: serviceName + "_readiness",
			Help: "Number of readiness transitions.",
		}, []string{"state"})

	prometheus.MustRegister(r.transitionCounter)
}

// increaseTransitionMetric increase counter
func (r *Readiness) increaseTransitionMetric(state bool) {
	stateString := fmt.Sprintf("%t", state)

	r.transitionCounter.WithLabelValues(stateString).Inc()
}
