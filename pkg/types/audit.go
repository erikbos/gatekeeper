package types

// Audit holds an audit
type (
	Audit struct {
		// Unique id
		Id string

		// Event timestamp in epoch milliseconds
		Timestamp    int64
		ChangeType   string
		EntityType   string
		EntityId     string
		IPaddress    string
		RequestID    string
		Role         string
		User         string
		UserAgent    string
		Organization string
		DeveloperID  string
		AppID        string
		Old          string
		New          string
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
