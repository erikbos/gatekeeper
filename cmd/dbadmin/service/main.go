package service

import (
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/db"
)

// New sets up services for all entities
func New(database *db.Database, changelogLogger *zap.Logger) *Service {

	changelog := NewChangelog(database, changelogLogger)
	return &Service{
		Listener:     NewListener(database, changelog),
		Route:        NewRoute(database, changelog),
		Cluster:      NewCluster(database, changelog),
		Organization: NewOrganization(database, changelog),
		Developer:    NewDeveloper(database, changelog),
		DeveloperApp: NewDeveloperApp(database, changelog),
		Key:          NewKey(database, changelog),
		APIProduct:   NewAPIProduct(database, changelog),
		User:         NewUser(database, changelog),
		Role:         NewRole(database, changelog),
	}
}
