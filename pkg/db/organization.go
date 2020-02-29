package db

//GetOrganizationByName retrieves an organization from database
//
func (d *Database) GetOrganizationByName(organization string) (types.Organization, error) {
	query := "SELECT * FROM organization WHERE name = ? LIMIT 1"
	developers := d.runGetDeveloperQuery(query, organization)
	if len(developers) > 0 {
		d.dbLookupHitsCounter.WithLabelValues(d.Hostname, "developers").Inc()
		return developers[0], nil
	}
	d.dbLookupMissesCounter.WithLabelValues(d.Hostname, "developers").Inc()
	return types.Developer{}, fmt.Errorf("Could not find organization (%s)", organization)
}
