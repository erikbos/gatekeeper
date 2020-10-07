package shared

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Readiness contains the readiness state of our application
type Readiness struct {
	// channel to receive notifications from application components
	channel chan ReadinessMessage
	// latest status of application readiness
	status bool
	// status message of latest readiness state change
	message string
	// timestamp of latest state change
	lastStateChange time.Time
	// counter for number of state changes
	transitionCounter *prometheus.CounterVec
	// central logger
	logger *zap.Logger
}

// ReadinessMessage gets send by components to indicate whether they are up or not
type ReadinessMessage struct {
	// Which application component reported its readiness
	Component string
	// Friendly message that provides background on our readiness
	Message string
	// Boolean readiness state of our component
	Up bool
}

// StartReadiness starts the readiness subsystem which waits for incoming ReadinessMessages
func StartReadiness(serviceName string, logger *zap.Logger) *Readiness {

	r := &Readiness{logger: logger}

	r.channel = make(chan ReadinessMessage)
	r.transitionCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: serviceName + "_readiness_transitions_total",
			Help: "Total number of readiness transitions.",
		}, []string{"state"})

	prometheus.MustRegister(r.transitionCounter)

	go r.readinessMainLoop()

	return r
}

// GetChannel return readiness notification channel
func (r *Readiness) GetChannel() chan ReadinessMessage {
	return r.channel
}

// readinessMainLoop runs continously in background waiting for messages
func (r *Readiness) readinessMainLoop() {

	for msg := range r.channel {
		r.logger.Debug("Received readiness update",
			zap.Bool("state", msg.Up),
			zap.String("message", msg.Message))

		r.updateReadinessState(msg.Up, msg.Message)
	}
}

// TODO we should have a failureThreshold and successThreshold before determining whether we are up or down

// updateReadinessState set current readiness state if it has changed
func (r *Readiness) updateReadinessState(newState bool, message string) {
	// we always update state in case:
	// - timestamp is zero (=at application startup)
	// - current state different than newstate
	if r.lastStateChange.IsZero() || r.status != newState {

		r.status = newState
		r.message = message
		r.lastStateChange = time.Now().UTC()

		stateString := fmt.Sprintf("%t", r.status)
		r.transitionCounter.WithLabelValues(stateString).Inc()

		r.logger.Info("Changing readiness state", zap.String("state", stateString))
	}
}

// ReadinessProbe shows our readiness status
func (r *Readiness) ReadinessProbe(c *gin.Context) {

	response := struct {
		Status          bool      `json:"status"`
		Message         string    `json:"message"`
		LastStateChange time.Time `json:"lastStateChange"`
	}{
		Status:          r.status,
		Message:         r.message,
		LastStateChange: r.lastStateChange,
	}

	if r.status {
		c.IndentedJSON(http.StatusOK, response)
	} else {
		c.IndentedJSON(http.StatusServiceUnavailable, response)
	}
}
