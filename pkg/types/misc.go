package types

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// returnJSONMessage returns an error message in case we do not handle API request
func returnJSONMessage(c *gin.Context, statusCode int, errorMessage error) {
	c.IndentedJSON(statusCode, gin.H{"message": fmt.Sprintf("%s", errorMessage)})
}

// AbortIfContentTypeNotJSON checks for json content-type and abort request
func AbortIfContentTypeNotJSON(c *gin.Context) {
	if c.Request.Header.Get("content-type") != "application/json" {
		returnJSONMessage(c, http.StatusUnsupportedMediaType,
			errors.New("Content-type application/json required when submitting data"))
		// do not continue request handling
		c.Abort()
	}
}
