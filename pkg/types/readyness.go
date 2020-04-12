package types

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Readyness contains the readyness of the application
type Readyness struct {
	status bool
}

// DisplayReadyness returns readyness status
func (r *Readyness) DisplayReadyness(c *gin.Context) {
	if r.status == true {
		c.JSON(http.StatusOK, "Ready")
	} else {
		c.JSON(http.StatusServiceUnavailable, "Not ready")
	}
}

// we should use failureThreshold and successThreshold before determining whether we are up or down

// Down changes status to down
func (r *Readyness) Down() {
	r.status = false
}

// Up changes status to up
func (r *Readyness) Up() {
	r.status = true
}
