package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// registerAPIProductRoutes registers all routes we handle
func (h *Handler) registerAPIProductRoutes(r *gin.RouterGroup) {
	r.GET("/apiproducts", h.handler(h.getAllAPIProducts))
	r.POST("/apiproducts", h.handler(h.createAPIProduct))

	r.GET("/apiproducts/:apiproduct", h.handler(h.getAPIProduct))
	r.POST("/apiproducts/:apiproduct", h.handler(h.updateAPIProduct))
	r.DELETE("/apiproducts/:apiproduct", h.handler(h.deleteAPIProduct))

	r.GET("/apiproducts/:apiproduct/attributes", h.handler(h.getAPIProductAttributes))
	r.POST("/apiproducts/:apiproduct/attributes", h.handler(h.updateAPIProductAttributes))

	r.GET("/apiproducts/:apiproduct/attributes/:attribute", h.handler(h.getAPIProductAttributeByName))
	r.POST("/apiproducts/:apiproduct/attributes/:attribute", h.handler(h.updateAPIProductAttributeByName))
	r.DELETE("/apiproducts/:apiproduct/attributes/:attribute", h.handler(h.deleteAPIProductAttributeByName))
}

const (
	attributeParameter = "attribute"

	// Name of apiproduct parameter in the route definition
	apiproductParameter = "apiproduct"
)

// getAllAPIProducts returns all apiproducts
func (h *Handler) getAllAPIProducts(c *gin.Context) handlerResponse {

	apiproducts, err := h.service.APIProduct.GetAll()
	if err != nil {
		return handleError(err)
	}
	return handleOK(StringMap{"apiproducts": apiproducts})
}

// getAPIProduct returns full details of one apiproduct
func (h *Handler) getAPIProduct(c *gin.Context) handlerResponse {

	apiproduct, err := h.service.APIProduct.Get(c.Param(apiproductParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(apiproduct)
}

// getAPIProductAttributes returns attributes of a apiproduct
func (h *Handler) getAPIProductAttributes(c *gin.Context) handlerResponse {

	apiproduct, err := h.service.APIProduct.Get(c.Param(apiproductParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOKAttributes(apiproduct.Attributes)
}

// getAPIProductAttributeByName returns one particular attribute of a apiproduct
func (h *Handler) getAPIProductAttributeByName(c *gin.Context) handlerResponse {

	apiproduct, err := h.service.APIProduct.Get(c.Param(apiproductParameter))
	if err != nil {
		return handleError(err)
	}
	attributeValue, err := apiproduct.Attributes.Get(attributeParameter)
	if err != nil {
		return handleError(err)
	}
	return handleOK(attributeValue)
}

// createAPIProduct creates a new apiproduct
func (h *Handler) createAPIProduct(c *gin.Context) handlerResponse {

	var newAPIProduct types.APIProduct
	if err := c.ShouldBindJSON(&newAPIProduct); err != nil {
		return handleBadRequest(err)
	}
	storedAPIProduct, err := h.service.APIProduct.Create(newAPIProduct, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleCreated(storedAPIProduct)
}

// updateAPIProduct updates an existing apiproduct
func (h *Handler) updateAPIProduct(c *gin.Context) handlerResponse {

	var updatedAPIProduct types.APIProduct
	if err := c.ShouldBindJSON(&updatedAPIProduct); err != nil {
		return handleBadRequest(err)
	}
	storedAPIProduct, err := h.service.APIProduct.Update(updatedAPIProduct, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(storedAPIProduct)
}

// updateAPIProductAttributes updates attributes of apiproduct
func (h *Handler) updateAPIProductAttributes(c *gin.Context) handlerResponse {

	var receivedAttributes struct {
		Attributes types.Attributes `json:"attribute"`
	}
	if err := c.ShouldBindJSON(&receivedAttributes); err != nil {
		return handleBadRequest(err)
	}

	if err := h.service.APIProduct.UpdateAttributes(c.Param(apiproductParameter),
		receivedAttributes.Attributes, h.who(c)); err != nil {
		return handleError(err)
	}
	return handleOKAttributes(receivedAttributes.Attributes)
}

// updateAPIProductAttributeByName update an attribute of apiproduct
func (h *Handler) updateAPIProductAttributeByName(c *gin.Context) handlerResponse {

	var receivedValue AttributeValue
	if err := c.ShouldBindJSON(&receivedValue); err != nil {
		return handleBadRequest(err)
	}

	newAttribute := types.Attribute{
		Name:  c.Param(attributeParameter),
		Value: receivedValue.Value,
	}

	if err := h.service.APIProduct.UpdateAttribute(c.Param(apiproductParameter),
		newAttribute, h.who(c)); err != nil {
		return handleError(err)
	}
	return handleOKAttribute(newAttribute)
}

// deleteAPIProductAttributeByName removes an attribute of apiproduct
func (h *Handler) deleteAPIProductAttributeByName(c *gin.Context) handlerResponse {

	attributeToDelete := c.Param(attributeParameter)
	oldValue, err := h.service.APIProduct.DeleteAttribute(c.Param(apiproductParameter), attributeToDelete, h.who(c))
	if err != nil {
		return handleBadRequest(err)
	}
	return handleOKAttribute(types.Attribute{
		Name:  attributeToDelete,
		Value: oldValue,
	})
}

// deleteAPIProduct deletes of one apiproduct
func (h *Handler) deleteAPIProduct(c *gin.Context) handlerResponse {

	deletedAPIproduct, err := h.service.APIProduct.Delete(c.Param(apiproductParameter), h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(deletedAPIproduct)
}
