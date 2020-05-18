package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Person is a person
type Person struct {
	ID        string   `json:"id,omitempty"`
	Firstname string   `json:"firstname,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
	Address   *Address `json:"address,omitempty"`
}

// Address is an address
type Address struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
}

var people []Person

func main() {
	gin.SetMode(gin.ReleaseMode)

	people = append(people, Person{ID: "1", Firstname: "John", Lastname: "Doe", Address: &Address{City: "City X", State: "State X"}})
	people = append(people, Person{ID: "2", Firstname: "Koko", Lastname: "Doe", Address: &Address{City: "City Z", State: "State Y"}})

	r := gin.New()
	r.GET("/people", getPeople)
	r.GET("/people/:id", getPerson)
	r.POST("/people/:id", createPerson)
	r.DELETE("/people/:id", deletePerson)

	listen := "0.0.0.0:8000"
	log.Info("Webadmin listening on ", listen)
	if err := r.Run(listen); err != nil {
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
	log.Printf("q: %s", c.Param("id"))
	c.Status(http.StatusNotFound)
}

func createPerson(c *gin.Context) {
	var person Person

	if err := c.ShouldBindJSON(&person); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
	}
	person.ID = c.Param("id")

	people = append(people, person)

	c.IndentedJSON(http.StatusOK, people)
}

func deletePerson(c *gin.Context) {
	for index, item := range people {
		if item.ID == c.Param("id") {
			people = append(people[:index], people[index+1:]...)
			break
		}
		c.IndentedJSON(http.StatusOK, people)
	}
}

// func handler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Printf("RemoteAddr = %q\n", r.RemoteAddr)
// 	fmt.Printf("Host = %q\n", r.Host)
// 	fmt.Printf("Protocol = %q\n", r.Proto)
// 	fmt.Printf("Method = %q\n", r.Method)
// 	fmt.Printf("URL = %q\n", r.URL)
// 	for k, v := range r.Header {
// 		fmt.Printf("Rx header [%q] = %q\n", k, v)
// 	}
// }

// func loggingMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		handler(w, r)
// 		// Call the next handler, which can be another middleware in the chain, or the final handler.
// 		next.ServeHTTP(w, r)
// 	})
// }
