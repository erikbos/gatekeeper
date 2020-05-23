package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Person is a person
type Person struct {
	ID        string `json:"id,omitempty"`
	Firstname string `json:"firstname,omitempty"`
	Lastname  string `json:"lastname,omitempty"`
	Street    string `json:"street,omitempty"`
	City      string `json:"city,omitempty"`
}

var people []Person

func main() {

	listen := flag.String("listen", "0.0.0.0:8000", "listen port")
	certificate := flag.String("certificate", "/config/tls.crt", "certificate filename")
	certificateKey := flag.String("certificatekey", "/config/tls.key", "certificate key filename")
	logFile := flag.String("accesslogfile", "/dev/null", "logfile")

	flag.Parse()

	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	f, _ := os.Create(*logFile)
	gin.DefaultWriter = io.MultiWriter(f)

	people = append(people, Person{ID: "1", Firstname: "John", Lastname: "Doe", Street: "1", City: "City X"})
	people = append(people, Person{ID: "2", Firstname: "Koko", Lastname: "Doe", Street: "2", City: "City Z"})

	r := gin.New()
	r.Use(gin.LoggerWithFormatter(LogHTTPRequest))

	r.GET("/people", getPeople)
	r.GET("/people/:id", getPerson)
	r.POST("/people/:id", createPerson)
	r.DELETE("/people/:id", deletePerson)

	log.Print("Webadmin listening on ", *listen)
	if err := http.ListenAndServeTLS(*listen, *certificate, *certificateKey, r); err != nil {
		log.Fatal(err)
	}
}

func getPeople(c *gin.Context) {
	c.IndentedJSON(http.StatusOK,
		gin.H{
			"people": people,
		})
}

func getPerson(c *gin.Context) {
	for _, item := range people {
		if item.ID == c.Param("id") {
			c.IndentedJSON(http.StatusOK, item)
			return
		}
	}
	c.Status(http.StatusNotFound)
}

func createPerson(c *gin.Context) {
	var person Person

	if err := c.ShouldBindJSON(&person); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
	}
	person.ID = c.Param("id")

	people = append(people, person)

	c.IndentedJSON(http.StatusOK, person)
}

func deletePerson(c *gin.Context) {
	var deletePerson Person

	for index, item := range people {
		if item.ID == c.Param("id") {
			deletePerson = people[index]
			people = append(people[:index], people[index+1:]...)
			break
		}
		c.IndentedJSON(http.StatusOK, deletePerson)
	}
}

// LogHTTPRequest logs details of an HTTP request
func LogHTTPRequest(param gin.LogFormatterParams) string {

	return fmt.Sprintf("%s - - %s \"%s %s %s\" %d %d \"%s\" \"%s\"\n",
		param.TimeStamp.Format(time.RFC3339),
		param.ClientIP,
		param.Method,
		param.Path,
		param.Request.Proto,
		param.StatusCode,
		param.Latency/time.Millisecond,
		param.Request.UserAgent(),
		param.ErrorMessage,
	)
}
