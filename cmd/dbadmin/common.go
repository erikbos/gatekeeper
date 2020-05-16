package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func setLastModifiedHeader(c *gin.Context, timeStamp int64) {
	c.Header("Last-Modified",
		time.Unix(0, timeStamp*int64(time.Millisecond)).UTC().Format(http.TimeFormat))
}

// returnJSONMessage returns an error message
func returnJSONMessage(c *gin.Context, statusCode int, errorMessage error) {

	c.IndentedJSON(statusCode,
		gin.H{
			"message": fmt.Sprintf("%s", errorMessage),
		})
}

func returnCanNotFindAttribute(c *gin.Context, name string) {

	returnJSONMessage(c,
		http.StatusNotFound,
		fmt.Errorf("Could not find attribute '%s'", name),
	)
}
