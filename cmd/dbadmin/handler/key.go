package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	maxConsumerKeyLength = 2048
	maxSecretKeyLength   = 2048
)

var (
	errNoConsumerKey      = types.NewBadRequestError(errors.New("consumerKey must be provided"))
	errNoSecretKey        = types.NewBadRequestError(errors.New("secretKey must be provided"))
	errConsumerKeyTooLong = types.NewBadRequestError(errors.New("consumerKey too long"))
	errSecretKeyTooLong   = types.NewBadRequestError(errors.New("secretKey too long"))
	errUnknownStatus      = types.NewBadRequestError(errors.New("unknown status requested"))
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
	if receivedKey.ConsumerKey == nil ||
		(receivedKey.ConsumerKey != nil && *receivedKey.ConsumerKey == "") {
		responseError(c, errNoConsumerKey)
		return
	}
	if len(*receivedKey.ConsumerKey) >= maxConsumerKeyLength {
		responseError(c, errConsumerKeyTooLong)
		return
	}
	if receivedKey.ConsumerSecret == nil ||
		(receivedKey.ConsumerSecret != nil && *receivedKey.ConsumerSecret == "") {
		responseError(c, errNoSecretKey)
		return
	}
	if len(*receivedKey.ConsumerSecret) >= maxSecretKeyLength {
		responseError(c, errSecretKeyTooLong)
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

// updates existing key
// (POST /v1/organizations/{organization_name}/developers/{developer_emailaddress}/apps/{app_name}/keys/{consumer_key})
func (h *Handler) PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameKeysConsumerKey(c *gin.Context, organizationName OrganizationName, developerEmailaddress DeveloperEmailaddress, appName AppName, consumerKey ConsumerKey, params PostV1OrganizationsOrganizationNameDevelopersDeveloperEmailaddressAppsAppNameKeysConsumerKeyParams) {

	if params.Action != nil && c.ContentType() == "application/octet-stream" {
		h.changeKeyStatus(c, string(developerEmailaddress), string(appName), string(consumerKey), string(*params.Action))
		return
	}
	var receivedKey types.Key
	if err := c.ShouldBindJSON(&receivedKey); err != nil {
		responseError(c, types.NewBadRequestError(err))
		return
	}
	// if consumerKey is provided in body it must match with path consumerKey
	if receivedKey.ConsumerKey != "" {
		if receivedKey.ConsumerKey != string(consumerKey) {
			responseErrorNameValueMisMatch(c)
			return
		}
	}
	storedKey, err := h.service.Key.Update(string(consumerKey), &receivedKey, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseKeyUpdated(c, &storedKey)
}

// change status of key
func (h *Handler) changeKeyStatus(c *gin.Context, developerEmailaddress, appName, consumerKey, requestedStatus string) {

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
	key, err := h.service.Key.Get(consumerKey)
	if err != nil {
		responseError(c, err)
		return
	}
	switch requestedStatus {
	case "approve":
		key.Approved()
	case "revoke":
		key.Revoke()
	default:
		responseError(c, errUnknownStatus)
		return
	}
	_, err = h.service.Key.Update(key.ConsumerKey, key, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Returns API response all user details
func (h *Handler) responseKeys(c *gin.Context, keys types.Keys) {

	c.IndentedJSON(http.StatusOK, Keys{
		Key: ToKeySlice(keys),
	})
}

func (h *Handler) responseKey(c *gin.Context, key *types.Key) {

	c.IndentedJSON(http.StatusOK, ToKeyResponse(key))
}

func (h *Handler) responseKeyCreated(c *gin.Context, key *types.Key) {

	c.IndentedJSON(http.StatusCreated, ToKeyResponse(key))
}

func (h *Handler) responseKeyUpdated(c *gin.Context, key *types.Key) {

	c.IndentedJSON(http.StatusOK, ToKeyResponse(key))
}

// type conversion

func ToKeySlice(keys types.Keys) *[]Key {

	all_keys := make([]Key, len(keys))
	for i := range keys {
		all_keys[i] = ToKeyResponse(&keys[i])
	}
	return &all_keys
}

func ToKeyResponse(k *types.Key) Key {

	return Key{
		ConsumerKey:    &k.ConsumerKey,
		ConsumerSecret: &k.ConsumerSecret,
		ExpiresAt:      &k.ExpiresAt,
		IssuedAt:       &k.IssuedAt,
		AppID:          &k.AppID,
		Status:         &k.Status,
		Attributes:     toAttributesResponse(k.Attributes),
		ApiProducts:    toKeyAPIProductStatusesResponse(k.APIProducts),
	}
}

func toKeyAPIProductStatusesResponse(apiProductStatuses types.KeyAPIProductStatuses) *[]KeyProduct {

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
