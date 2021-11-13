package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// returns all keys of developer application
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/keys)
func (h *Handler) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameKeys(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	_, err := h.service.Developer.Get(string(developerEmailaddress))
	if err != nil {
		responseError(c, err)
		return
	}
	application, err := h.service.DeveloperApp.GetByName(string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	keys, err := h.service.Key.GetByDeveloperAppID(application.AppID)
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseKeys(c, keys)
}

// Adds provided key to application
// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/keys/create)
func (h *Handler) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameKeysCreate(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName) {

	var receivedKey Key
	if err := c.ShouldBindJSON(&receivedKey); err != nil {
		responseError(c, types.NewBadRequestError(err))
		return
	}
	_, err := h.service.Developer.Get(string(developerEmailaddress))
	if err != nil {
		responseError(c, err)
		return
	}
	application, err := h.service.DeveloperApp.GetByName(string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	newKey := fromKey(receivedKey)
	storedKey, err := h.service.Key.Create(newKey, application, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseKeyCreated(c, &storedKey)
}

// Removes key from application
// (DELETE /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/keys/{consumer_key})
func (h *Handler) DeleteV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameKeysConsumerKey(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, consumerKey ConsumerKey) {

	deletedKey, err := h.service.Key.Delete(string(consumerKey), h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseKey(c, &deletedKey)
}

// returns one key of one developer application
// (GET /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/keys/{consumer_key})
func (h *Handler) GetV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameKeysConsumerKey(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, consumerKey ConsumerKey) {

	_, err := h.service.Developer.Get(string(developerEmailaddress))
	if err != nil {
		responseError(c, err)
		return
	}
	_, err = h.service.DeveloperApp.GetByName(string(appName))
	if err != nil {
		responseError(c, err)
		return
	}
	key, err := h.service.Key.Get(string(consumerKey))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseKey(c, key)
}

// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/keys/{consumer_key})
func (h *Handler) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameKeysConsumerKey(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, consumerKey ConsumerKey) {

	var receivedKey types.Key
	if err := c.ShouldBindJSON(&receivedKey); err != nil {
		responseError(c, types.NewBadRequestError(err))
		return
	}
	// apikey in path must match consumer key in posted body
	if receivedKey.ConsumerKey != string(consumerKey) {
		responseErrorNameValueMisMatch(c)
		return
	}
	storedKey, err := h.service.Key.Update(receivedKey, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseKeyUpdated(c, &storedKey)
}

// Returns API response all user details
func (h *Handler) responseKeys(c *gin.Context, keys types.Keys) {

	all_keys := make([]Key, len(keys))
	for i := range keys {
		all_keys[i] = h.ToKeyResponse(&keys[i])
	}
	c.IndentedJSON(http.StatusOK, Keys{
		Key: &all_keys,
	})
}

func (h *Handler) responseKey(c *gin.Context, key *types.Key) {

	c.IndentedJSON(http.StatusOK, h.ToKeyResponse(key))
}

func (h *Handler) responseKeyCreated(c *gin.Context, key *types.Key) {

	c.IndentedJSON(http.StatusCreated, h.ToKeyResponse(key))
}

func (h *Handler) responseKeyUpdated(c *gin.Context, key *types.Key) {

	c.IndentedJSON(http.StatusOK, h.ToKeyResponse(key))
}

// type conversion

func (h *Handler) ToKeyResponse(l *types.Key) Key {

	key := Key{
		ConsumerKey:    &l.ConsumerKey,
		ConsumerSecret: &l.ConsumerSecret,
		ExpiresAt:      &l.ExpiresAt,
		IssuedAt:       &l.IssuedAt,
		AppID:          &l.AppID,
		Status:         &l.Status,
		Attributes:     toAttributesResponse(l.Attributes),
	}
	if l.APIProducts != nil {
		key.ApiProducts = ToKeyAPIProductStatusesResponse(l.APIProducts)
	}
	return key
}

func ToKeyAPIProductStatusesResponse(apiProductStatuses types.KeyAPIProductStatuses) *[]KeyProduct {

	product_statuses := make([]KeyProduct, len(apiProductStatuses))
	for i, v := range apiProductStatuses {
		product_statuses[i] = KeyProduct{
			Apiproduct: &v.Apiproduct,
			Status:     &v.Status,
		}
	}
	return &product_statuses
}

func fromKey(k Key) types.Key {

	key := types.Key{}
	if k.ConsumerKey != nil {
		key.ConsumerKey = *k.ConsumerKey
	}
	if k.ConsumerSecret != nil {
		key.ConsumerSecret = *k.ConsumerSecret
	}
	if k.ExpiresAt != nil {
		key.ExpiresAt = *k.ExpiresAt
	}
	if k.IssuedAt != nil {
		key.IssuedAt = *k.IssuedAt
	}
	if k.AppID != nil {
		key.AppID = *k.AppID
	}
	if k.Status != nil {
		key.Status = *k.Status
	}
	if k.Attributes != nil {
		key.Attributes = fromAttributesRequest(k.Attributes)
	}
	return key
}

// func fromRoleAllow(a *[]RoleAllow) types.Allows {

// 	if a == nil {
// 		return types.NullAllows
// 	}
// 	all_attributes := make([]types.Allow, len(*a))
// 	for i, a := range *a {
// 		all_attributes[i] = types.Allow{
// 			Methods: *a.Methods,
// 			Paths:   *a.Paths,
// 		}
// 	}
// 	return all_attributes
// }
