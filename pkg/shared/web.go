package shared

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ServeEmbedFile serves one file from an embedded filesystem
func ServeEmbedFile(filesystem fs.FS, embedFilePath string) gin.HandlerFunc {

	return func(c *gin.Context) {
		c.FileFromFS(embedFilePath, http.FS(filesystem))
	}
}

// ServeEmbedDirectory serves a directory from an embedded filesystem,
// directory must provided via gin's URL param field ":path"
func ServeEmbedDirectory(filesystem fs.FS, embedPath string) gin.HandlerFunc {

	subtreeFileSystem, err := fs.Sub(filesystem, embedPath)
	if err != nil {
		log.Fatal(err)
	}
	return func(c *gin.Context) {
		c.FileFromFS(c.Param("path"), http.FS(subtreeFileSystem))
	}
}
