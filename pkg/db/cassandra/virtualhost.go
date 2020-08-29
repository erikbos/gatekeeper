package cassandra

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

const (
	// Prometheus label for metrics of db interactions
	virtualHostMetricLabel = "virtualhosts"

	// List of organization columns we use
	virtualColumns = `name,
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

// VirtualhostStore holds our OrganizationStore config
type VirtualhostStore struct {
	db *Database
}

// NewVirtualhostStore creates virtualhost instance
func NewVirtualhostStore(database *Database) *VirtualhostStore {
	return &VirtualhostStore{
		db: database,
	}
}

// GetAll retrieves all virtualhosts
func (s *VirtualhostStore) GetAll() ([]shared.VirtualHost, error) {

	query := "SELECT " + virtualColumns + " FROM virtual_hosts"
	virtualhosts, err := s.runGetVirtualHostQuery(query)
	if err != nil {
		return []shared.VirtualHost{}, err
	}

	if len(virtualhosts) == 0 {
		s.db.metrics.QueryMiss(virtualHostMetricLabel)
		return []shared.VirtualHost{}, errors.New("Can not retrieve list of virtualhosts")
	}

	s.db.metrics.QueryHit(virtualHostMetricLabel)
	return virtualhosts, nil
}

// GetByName retrieves a virtualhost
func (s *VirtualhostStore) GetByName(virtualHost string) (*shared.VirtualHost, error) {

	query := "SELECT " + virtualColumns + " FROM virtual_hosts WHERE name = ? LIMIT 1"
	virtualhosts, err := s.runGetVirtualHostQuery(query, virtualHost)
	if err != nil {
		return nil, err
	}

	if len(virtualhosts) == 0 {
		s.db.metrics.QueryMiss(virtualHostMetricLabel)
		return nil, fmt.Errorf("Can not find route (%s)", virtualHost)
	}

	s.db.metrics.QueryHit(virtualHostMetricLabel)
	return &virtualhosts[0], nil
}

// runGetVirtualHostQuery executes CQL query and returns resultset
func (s *VirtualhostStore) runGetVirtualHostQuery(query string,
	queryParameters ...interface{}) ([]shared.VirtualHost, error) {

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	var virtualhosts []shared.VirtualHost

	iter := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		virtualhosts = append(virtualhosts, shared.VirtualHost{
			Name:             columnValueString(m, "name"),
			DisplayName:      columnValueString(m, "display_name"),
			VirtualHosts:     shared.VirtualHost{}.VirtualHosts.Unmarshal(columnValueString(m, "virtual_hosts")),
			Port:             columnValueInt(m, "port"),
			RouteGroup:       columnValueString(m, "route_group"),
			Policies:         columnValueString(m, "policies"),
			Attributes:       shared.VirtualHost{}.Attributes.Unmarshal(columnValueString(m, "attributes")),
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
		return []shared.VirtualHost{}, err
	}
	return virtualhosts, nil
}

// UpdateByName updates a virtualhost
func (s *VirtualhostStore) UpdateByName(vhost *shared.VirtualHost) error {

	vhost.Attributes.Tidy()
	vhost.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	query := "INSERT INTO virtual_hosts (" + virtualColumns + ") VALUES(?,?,?,?,?,?,?,?,?,?,?,?)"
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

		return fmt.Errorf("Cannot update virtualhost '%s', '%v'", vhost.Name, err)
	}
	return nil
}

// DeleteByName deletes a virtualhost
func (s *VirtualhostStore) DeleteByName(virtualHostToDelete string) error {

	_, err := s.GetByName(virtualHostToDelete)
	if err != nil {
		return err
	}

	query := "DELETE FROM virtual_hosts WHERE name = ?"
	return s.db.CassandraSession.Query(query, virtualHostToDelete).Exec()
}
