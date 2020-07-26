package db

import (
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type (
	// Readiness the readiness storage interface
	Readiness interface {
		// RunReadinessCheck runs a database readiness check continously
		RunReadinessCheck(chan shared.ReadinessMessage)
	}
)
