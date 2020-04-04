package types

//Organization contains everything about a Organization
//
type Organization struct {
	Attributes     []AttributeKeyValues `json:"attributes"`
	CreatedAt      int64                `json:"createdAt"`
	CreatedBy      string               `json:"createdBy"`
	DisplayName    string               `json:"displayName"`
	Key            string               `json:"key"`
	LastmodifiedAt int64                `json:"lastmodifiedAt"`
	LastmodifiedBy string               `json:"lastmodifiedBy"`
	Name           string               `json:"name" binding:"required"`
}

//Developer contains everything about a Developer
//
type Developer struct {
	DeveloperID      string               `json:"developerId"`
	Apps             []string             `json:"apps"`
	Attributes       []AttributeKeyValues `json:"attributes"`
	CreatedAt        int64                `json:"createdAt"`
	CreatedBy        string               `json:"createdBy"`
	Email            string               `json:"email" binding:"required"`
	FirstName        string               `json:"firstName" binding:"required"`
	LastName         string               `json:"lastName" binding:"required"`
	LastmodifiedAt   int64                `json:"lastmodifiedAt"`
	LastmodifiedBy   string               `json:"lastmodifiedBy"`
	OrganizationName string               `json:"organizationName"`
	Salt             string               `json:"salt"`
	Status           string               `json:"status"`
	UserName         string               `json:"userName" binding:"required"`
}

//DeveloperApp contains everything about a Developer Application
//
type DeveloperApp struct {
	DeveloperAppID   string               `json:"key"`
	AccessType       string               `json:"accessType"`
	AppFamily        string               `json:"appFamily"`
	AppID            string               `json:"appId"`
	AppType          string               `json:"appType"`
	Attributes       []AttributeKeyValues `json:"attributes"`
	CallbackURL      string               `json:"callbackUrl"`
	CreatedAt        int64                `json:"createdAt"`
	CreatedBy        string               `json:"createdBy"`
	Credentials      []AppCredential      `json:"credentials"`
	DisplayName      string               `json:"displayName"`
	LastmodifiedAt   int64                `json:"lastmodifiedAt"`
	LastmodifiedBy   string               `json:"lastmodifiedBy"`
	Name             string               `json:"name" binding:"required"`
	OrganizationName string               `json:"organizationName"`
	ParentID         string               `json:"parentId"`
	ParentStatus     string               `json:"parentStatus"`
	Status           string               `json:"status"`
	// Key              string               `json:"DeveloperAppID"`
}

//AppCredential contains an apikey entitlement
//
type AppCredential struct {
	ConsumerKey       string               `json:"key"`
	APIProducts       []APIProductStatus   `json:"apiProducts"`
	AppStatus         string               `json:"appStatus"`
	Attributes        []AttributeKeyValues `json:"attributes"`
	CompanyStatus     string               `json:"companyStatus"`
	ConsumerSecret    string               `json:"consumerSecret"`
	CredentialMethod  string               `json:"credentialMethod"`
	DeveloperStatus   string               `json:"developerStatus"`
	ExpiresAt         int64                `json:"expiresAt"`
	IssuedAt          int64                `json:"issuesAt"`
	OrganizationAppID string               `json:"organizationAppId"`
	OrganizationName  string               `json:"organizationName"`
	Scopes            string               `json:"scopes"`
	Status            string               `json:"status"`
}

// APIProductStatus contains whether an apikey's assigned apiproduct has been approved
type APIProductStatus struct {
	Status     string `json:"status"`
	Apiproduct string `json:"apiProduct"`
}

//APIProduct type contains everything about an API product
//
type APIProduct struct {
	Key              string               `json:"key"`
	APIResources     []string             `json:"api_resources"`
	ApprovalType     string               `json:"approval_type"`
	Attributes       []AttributeKeyValues `json:"attributes"`
	CreatedAt        int64                `json:"created_at"`
	CreatedBy        string               `json:"created_by"`
	Description      string               `json:"description"`
	DisplayName      string               `json:"display_name"`
	Environments     string               `json:"environments"`
	LastmodifiedAt   int64                `json:"lastmodified_at"`
	LastmodifiedBy   string               `json:"lastmodified_by"`
	Name             string               `json:"name"`
	OrganizationName string               `json:"organization_name"`
	Proxies          []string             `json:"proxies"`
	Scopes           string               `json:"scopes"`
}

//APIProxy contains mapping of paths to upstream
//
type APIProxy struct {
	Name             string               `json:"name"`
	DisplayName      string               `json:"display_name"`
	OrganizationName string               `json:"organization_name"`
	Policies         []string             `json:"policies"`
	Attributes       []AttributeKeyValues `json:"attributes"`
	CreatedAt        int64                `json:"created_at"`
	CreatedBy        string               `json:"created_by"`
	LastmodifiedAt   int64                `json:"lastmodified_at"`
	LastmodifiedBy   string               `json:"lastmodified_by"`
	BasePath         string
	Routes           []struct {
		prefix   string
		upstream string
	}
	VirtualHosts []string `json:"virtual_hosts"`
}

//AttributeKeyValues is an array with attributes of developer or developer app
type AttributeKeyValues struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// bla
type VirtualHost struct {
	Name              string   `json:"name"`
	DisplayName       string   `json:"display_name"`
	Description       string   `json:"description"`
	VirtualHosts      []string `json:"virtual_hosts"`
	TLSCipherSuites   string
	TLSMinimumVersion string
}

// func initialize() {
// 	// vhosts := make([]VirtualHost, 50)
// 	vhosts := []VirtualHost{
// 		VirtualHost{
// 			Name: "test1",
// 			VirtualHosts: []string{
// 				"nozomi.sievie.com",
// 			},
// 			TLSMinimumVersion: "TLSv1_2",
// 			TLSCipherSuites: "[ECDHE-RSA-CHACHA20-POLY1305|ECDHE-RSA-AES256-GCM-SHA384|ECDHE-RSA-AES128-GCM-SHA256]",
// 		},
// 		VirtualHost{
// 			Name: "test2",
// 			VirtualHosts: []string{
// 				"nozomi.sievie.be",
// 			},
// 		},
// 	}
// }

// Cluster holds configuration of an upstream cluster
type Cluster struct {
	Name           string `json:"name"`
	DisplayName    string `json:"displayName"`
	HostName       string `json:"hostName"`
	HostPort       int16  `json:"hostPort"`
	CreatedAt      int64  `json:"createdAt"`
	CreatedBy      string `json:"createdBy"`
	LastmodifiedAt int64  `json:"lastmodifiedAt"`
	LastmodifiedBy string `json:"lastmodifiedBy"`
}
