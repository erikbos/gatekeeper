package cassandra

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const virtualHostMetricLabel = "virtualhosts"

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

	query := "SELECT * FROM virtual_hosts"
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

	query := "SELECT * FROM virtual_hosts WHERE name = ? LIMIT 1"
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
			Name:             m["name"].(string),
			DisplayName:      m["display_name"].(string),
			Port:             m["port"].(int),
			VirtualHosts:     shared.VirtualHost{}.VirtualHosts.Unmarshal(m["virtual_hosts"].(string)),
			Attributes:       shared.VirtualHost{}.Attributes.Unmarshal(m["attributes"].(string)),
			RouteGroup:       m["route_group"].(string),
			Policies:         m["policies"].(string),
			OrganizationName: m["organization_name"].(string),
			CreatedAt:        m["created_at"].(int64),
			CreatedBy:        m["created_by"].(string),
			LastmodifiedAt:   m["lastmodified_at"].(int64),
			LastmodifiedBy:   m["lastmodified_by"].(string),
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

	if err := s.db.CassandraSession.Query(`INSERT INTO virtual_hosts (
name,
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
lastmodified_by) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,

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
