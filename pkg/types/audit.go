package types

// Audit holds an audit
type (
	Audit struct {
		// Unique id
		ID string

		// Event timestamp in epoch milliseconds
		Timestamp int64
		AuditType string

		IPaddress string
		RequestID string
		Role      string
		User      string
		UserAgent string

		Organization string
		Company      string
		DeveloperID  string
		AppID        string

		EntityType string
		EntityID   string
		OldValue   map[string]interface{}
		NewValue   map[string]interface{}
	}

	// Audits holds one or more audits
	Audits []Audit
)

var (
	// NullAudit is an empty audit type
	NullAudit = Audit{}

	// NullAudits is an empty audit slice
	NullAudits = Audits{}
)
