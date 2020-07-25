package db

type (
	// Database is our overall database interface
	Database struct {
		Virtualhost  VirtualhostStore
		Route        RouteStore
		Cluster      ClusterStore
		Organization OrganizationStore
		Developer    DeveloperStore
		DeveloperApp DeveloperAppStore
		APIProduct   APIProductStore
		Credential   CredentialStore
		OAuth        OAuthStore
		Readiness
	}
)
