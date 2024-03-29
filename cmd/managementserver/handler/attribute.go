package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// API responses

func (h *Handler) responseAttributes(c *gin.Context, attributes types.Attributes) {

	c.IndentedJSON(http.StatusOK, Attributes{
		Attribute: toAttributesResponse(attributes),
	})
}

func (h *Handler) responseAttributeRetrieved(c *gin.Context, attribute *types.Attribute) {

	c.IndentedJSON(http.StatusOK, AttributeRetrieved{
		Name:  &attribute.Name,
		Value: &attribute.Value,
	})
}

func (h *Handler) responseAttributeUpdated(c *gin.Context, attribute *types.Attribute) {

	c.IndentedJSON(http.StatusOK, AttributeUpdated{
		Name:  &attribute.Name,
		Value: &attribute.Value,
	})
}

func (h *Handler) responseAttributeDeleted(c *gin.Context, attribute *types.Attribute) {

	c.IndentedJSON(http.StatusOK, AttributeDeleted{
		Name:  &attribute.Name,
		Value: &attribute.Value,
	})
}

func toAttributesResponse(attributes types.Attributes) *[]Attribute {

	allAttributes := make([]Attribute, len(attributes))
	for i := range attributes {
		allAttributes[i] = Attribute{
			Name:  &attributes[i].Name,
			Value: &attributes[i].Value,
		}
	}
	return &allAttributes
}

func ToAttributeResponse(attribute *types.Attribute) Attribute {

	return Attribute{
		Name:  &attribute.Name,
		Value: &attribute.Value,
	}
}

func fromAttributesRequest(attributes *[]Attribute) types.Attributes {

	if attributes == nil {
		return types.Attributes{}
	}
	allAttributes := make([]types.Attribute, len(*attributes))
	for i, a := range *attributes {
		allAttributes[i] = fromAttributeRequest(a)
	}
	return allAttributes
}

func fromAttributeRequest(attribute Attribute) types.Attribute {

	a := types.Attribute{}
	if attribute.Name != nil {
		a.Name = *attribute.Name
	}
	if attribute.Value != nil {
		a.Value = *attribute.Value
	}
	return a
}
