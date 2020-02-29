package db

//GetDeveloperAppByName returns details of a DeveloperApplication looked up by Name
//
func (d *Database) GetDeveloperAppByName(organization, developerAppName string) (types.DeveloperApp, error) {
	query := "SELECT * FROM apps WHERE name = ? LIMIT 1"
	developerapps := d.runGetDeveloperAppQuery(query, developerAppName)
	if len(developerapps) > 0 {
		d.dbLookupHitsCounter.WithLabelValues(d.Hostname, "apps").Inc()
		return developerapps[0], nil
	}
	d.dbLookupMissesCounter.WithLabelValues(d.Hostname, "apps").Inc()
	return types.DeveloperApp{}, errors.New("Could not find developer by name")
}

//GetDeveloperAppByID returns details of a DeveloperApplication looked up by ID
//
func (d *Database) GetDeveloperAppByID(organization, developerAppID string) (types.DeveloperApp, error) {
	query := "SELECT * FROM apps WHERE key = ? LIMIT 1"
	developerapps := d.runGetDeveloperAppQuery(query, developerAppID)
	if len(developerapps) > 0 {
		d.dbLookupHitsCounter.WithLabelValues(d.Hostname, "apps").Inc()
		return developerapps[0], nil
	}
	d.dbLookupMissesCounter.WithLabelValues(d.Hostname, "apps").Inc()
	return types.DeveloperApp{}, errors.New("Could not find developer by app id")
}

func (d *Database) runGetDeveloperAppQuery(query, queryParameter string) []types.DeveloperApp {
	var developerapps []types.DeveloperApp

	//Set timer to record how long this function run
	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameter).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		developerapps = append(developerapps, types.DeveloperApp{
			AccessType:  m["access_type"].(string),
			AppFamily:   m["app_family"].(string),
			AppID:       m["app_id"].(string),
			AppType:     m["app_type"].(string),
			Attributes:  d.unmarshallJSONArrayOfAttributes(m["attributes"].(string), false),
			CallbackURL: m["callback_url"].(string),
			CreatedAt:   m["created_at"].(int64),
			CreatedBy:   m["created_by"].(string),
			// DeveloperAppID:   developerAppID,
			DisplayName:      m["display_name"].(string),
			Key:              m["key"].(string),
			LastmodifiedAt:   m["lastmodified_at"].(int64),
			LastmodifiedBy:   m["lastmodified_by"].(string),
			Name:             m["name"].(string),
			OrganizationName: m["organization_name"].(string),
			ParentID:         m["parent_id"].(string),
			ParentStatus:     m["parent_status"].(string),
			Status:           m["status"].(string),
		})
		m = map[string]interface{}{}
	}
	return developerapps
}

///////////////////////////////////////////////////////////////////////////////

//GetAppCredentialByKey returns details of a single apikey
//
func (d *Database) GetAppCredentialByKey(key string) (types.AppCredential, error) {
	var appcredentials []types.AppCredential

	query := "SELECT * FROM app_credentials WHERE key = ? LIMIT 1"
	appcredentials = d.runGetAppCredentialQuery(query, key)
	if len(appcredentials) > 0 {
		d.dbLookupHitsCounter.WithLabelValues(d.Hostname, "app_credentials").Inc()
		return appcredentials[0], nil
	}
	d.dbLookupMissesCounter.WithLabelValues(d.Hostname, "app_credentials").Inc()
	return types.AppCredential{}, errors.New("Could not find apikey")
}

//GetAppCredentialByDeveloperAppID returns an array with apikey details of a developer app
// FIXME contains LIMIT
func (d *Database) GetAppCredentialByDeveloperAppID(organizationAppID string) ([]types.AppCredential, error) {
	var appcredentials []types.AppCredential

	// FIXME hardcoded row limit
	query := "SELECT * FROM app_credentials WHERE organization_app_id = ? LIMIT 1000"
	appcredentials = d.runGetAppCredentialQuery(query, organizationAppID)
	if len(appcredentials) > 0 {
		d.dbLookupHitsCounter.WithLabelValues(d.Hostname, "app_credentials").Inc()
		return appcredentials, nil
	}
	d.dbLookupMissesCounter.WithLabelValues(d.Hostname, "app_credentials").Inc()
	return appcredentials, errors.New("Could not find apikeys of developer app")
}

//runAppCredentialQuery executes CQL query and returns resulset
//
func (d *Database) runGetAppCredentialQuery(query, queryParameter string) []types.AppCredential {
	var appcredentials []types.AppCredential

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iterable := d.cassandraSession.Query(query, queryParameter).Iter()
	m := make(map[string]interface{})
	for iterable.MapScan(m) {
		appcredential := types.AppCredential{
			ConsumerKey:       m["key"].(string),
			AppStatus:         m["app_status"].(string),
			Attributes:        m["attributes"].(string),
			CompanyStatus:     m["company_status"].(string),
			ConsumerSecret:    m["consumer_secret"].(string),
			CredentialMethod:  m["credential_method"].(string),
			DeveloperStatus:   m["developer_status"].(string),
			ExpiresAt:         m["expires_at"].(int64),
			IssuedAt:          m["issued_at"].(int64),
			OrganizationAppID: m["organization_app_id"].(string),
			OrganizationName:  m["organization_name"].(string),
			Scopes:            m["scopes"].(string),
			Status:            m["status"].(string),
		}
		if m["api_products"].(string) != "" {
			appcredential.APIProducts = make([]types.APIProductStatus, 0)
			json.Unmarshal([]byte(m["api_products"].(string)), &appcredential.APIProducts)
		}
		appcredentials = append(appcredentials, appcredential)
		m = map[string]interface{}{}
	}
	return appcredentials
}
