package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

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
created_at,
created_by,
lastmodified_at,
lastmodified_by`
)

// ListenerStore holds our ListenerStore config
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
func (s *ListenerStore) GetAll() (types.Listeners, types.Error) {

	query := "SELECT " + listenerColumns + " FROM listeners"
	listeners, err := s.runGetListenerQuery(query)
	if err != nil {
		s.db.metrics.QueryFailed(listenerMetricLabel)
		return types.NullListeners, types.NewDatabaseError(err)

	}

	s.db.metrics.QueryHit(listenerMetricLabel)
	return listeners, nil
}

// Get retrieves a listener
func (s *ListenerStore) Get(listenerName string) (*types.Listener, types.Error) {

	query := "SELECT " + listenerColumns + " FROM listeners WHERE name = ? LIMIT 1"
	listeners, err := s.runGetListenerQuery(query, listenerName)
	if err != nil {
		s.db.metrics.QueryFailed(listenerMetricLabel)
		return nil, types.NewDatabaseError(err)

	}

	if len(listeners) == 0 {
		s.db.metrics.QueryMiss(listenerMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("can not find listener '%s'", listenerName))
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
			Attributes:     AttributesUnmarshal(columnValueString(m, "attributes")),
			CreatedAt:      columnValueInt64(m, "created_at"),
			CreatedBy:      columnValueString(m, "created_by"),
			Name:           columnValueString(m, "name"),
			DisplayName:    columnValueString(m, "display_name"),
			LastModifiedAt: columnValueInt64(m, "lastmodified_at"),
			LastModifiedBy: columnValueString(m, "lastmodified_by"),
			Policies:       columnValueString(m, "policies"),
			Port:           columnValueInt(m, "port"),
			RouteGroup:     columnValueString(m, "route_group"),
			VirtualHosts:   stringSliceUnmarshal(columnValueString(m, "virtual_hosts")),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		return types.Listeners{}, err
	}
	return listeners, nil
}

// Update updates a listener
func (s *ListenerStore) Update(l *types.Listener) types.Error {

	query := "INSERT INTO listeners (" + listenerColumns + ") VALUES(?,?,?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		l.Name,
		l.DisplayName,
		stringSliceMarshal(l.VirtualHosts),
		l.Port,
		l.RouteGroup,
		l.Policies,
		AttributesMarshal(l.Attributes),
		l.CreatedAt,
		l.CreatedBy,
		l.LastModifiedAt,
		l.LastModifiedBy).Exec(); err != nil {

		s.db.metrics.QueryFailed(listenerMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update listener '%s'", l.Name))
	}
	return nil
}

// Delete deletes a listener
func (s *ListenerStore) Delete(listenerToDelete string) types.Error {

	query := "DELETE FROM listeners WHERE name = ?"
	if err := s.db.CassandraSession.Query(query, listenerToDelete).Exec(); err != nil {
		s.db.metrics.QueryFailed(listenerMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}
