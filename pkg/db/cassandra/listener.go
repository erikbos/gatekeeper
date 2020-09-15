package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// Prometheus label for metrics of db interactions
	listenerMetricLabel = "listeners"

	// List of listener columns we use
	listenerColumns = `name,
display_name,
virtual_hosts,
port,
route_group,
policies,
attributes,
organization_name,
created_at,
created_by,
lastmodified_at,
lastmodified_by`
)

// ListenerStore holds our OrganizationStore config
type ListenerStore struct {
	db *Database
}

// NewListenerStore creates listener instance
func NewListenerStore(database *Database) *ListenerStore {
	return &ListenerStore{
		db: database,
	}
}

// GetAll retrieves all listeners
func (s *ListenerStore) GetAll() (types.Listeners, error) {

	query := "SELECT " + listenerColumns + " FROM listeners"
	listeners, err := s.runGetListenerQuery(query)
	if err != nil {
		s.db.metrics.QueryMiss(listenerMetricLabel)
		return types.Listeners{}, err
	}

	s.db.metrics.QueryHit(listenerMetricLabel)
	return listeners, nil
}

// GetByName retrieves a listener
func (s *ListenerStore) GetByName(listenerName string) (*types.Listener, error) {

	query := "SELECT " + listenerColumns + " FROM listeners WHERE name = ? LIMIT 1"
	listeners, err := s.runGetListenerQuery(query, listenerName)
	if err != nil {
		return nil, err
	}

	if len(listeners) == 0 {
		s.db.metrics.QueryMiss(listenerMetricLabel)
		return nil, fmt.Errorf("Can not find listener (%s)", listenerName)
	}

	s.db.metrics.QueryHit(listenerMetricLabel)
	return &listeners[0], nil
}

// runGetListenerQuery executes CQL query and returns resultset
func (s *ListenerStore) runGetListenerQuery(query string,
	queryParameters ...interface{}) (types.Listeners, error) {

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	var listeners types.Listeners

	iter := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		listeners = append(listeners, types.Listener{
			Name:             columnValueString(m, "name"),
			DisplayName:      columnValueString(m, "display_name"),
			VirtualHosts:     types.Listener{}.VirtualHosts.Unmarshal(columnValueString(m, "virtual_hosts")),
			Port:             columnValueInt(m, "port"),
			RouteGroup:       columnValueString(m, "route_group"),
			Policies:         columnValueString(m, "policies"),
			Attributes:       types.Listener{}.Attributes.Unmarshal(columnValueString(m, "attributes")),
			OrganizationName: columnValueString(m, "organization_name"),
			CreatedAt:        columnValueInt64(m, "created_at"),
			CreatedBy:        columnValueString(m, "created_by"),
			LastmodifiedAt:   columnValueInt64(m, "lastmodified_at"),
			LastmodifiedBy:   columnValueString(m, "lastmodified_by"),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		log.Error(err)
		return types.Listeners{}, err
	}
	return listeners, nil
}

// UpdateByName updates a listener
func (s *ListenerStore) UpdateByName(vhost *types.Listener) error {

	vhost.Attributes.Tidy()
	vhost.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	query := "INSERT INTO listeners (" + listenerColumns + ") VALUES(?,?,?,?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		vhost.Name,
		vhost.DisplayName,
		vhost.VirtualHosts.Marshal(),
		vhost.Port,
		vhost.RouteGroup,
		vhost.Policies,
		vhost.Attributes.Marshal(),
		vhost.OrganizationName,
		vhost.CreatedAt,
		vhost.CreatedBy,
		vhost.LastmodifiedAt,
		vhost.LastmodifiedBy).Exec(); err != nil {

		return fmt.Errorf("Cannot update listener '%s', '%v'", vhost.Name, err)
	}
	return nil
}

// DeleteByName deletes a listener
func (s *ListenerStore) DeleteByName(listenerToDelete string) error {

	_, err := s.GetByName(listenerToDelete)
	if err != nil {
		return err
	}

	query := "DELETE FROM listeners WHERE name = ?"
	return s.db.CassandraSession.Query(query, listenerToDelete).Exec()
}
