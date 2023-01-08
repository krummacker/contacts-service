package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Contact is the data structure for a person that we know.
type Contact struct {
	gorm.Model
	Name     string
	Phone    string
	Birthday time.Time
}

// db is a handle to the OR mapper.
var db *gorm.DB
var err error

func main() {
	setupORMapper()
	db.AutoMigrate(&Contact{}) // Define database schema.
	populateDatabase()
	setupHttpRouter()
}

// setupORMapper initializes the object relational mapper and the database
// connection. The connection parameters are taken from the system's
// environment variables.
//
// Usage example:
// > export HOST=localhost && export DBUSER=postgres && export DBPASSWORD=Hztju8zgf
// > go run contacts-service.go
func setupORMapper() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=5432",
		os.Getenv("HOST"), os.Getenv("DBUSER"), os.Getenv("DBPASSWORD"))
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		panic("failed to connect to database")
	}
}

// populateDatabase enters initial test data into the database. If the data is
// already present in the table then it is not added again.
func populateDatabase() {
	initialContacts := []Contact{
		{
			Name:     "Dirk Krummacker",
			Phone:    "+420 123 456 789",
			Birthday: time.Date(1974, time.November, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			Name:     "Pavla Krummackerova",
			Phone:    "+420 023 454 244",
			Birthday: time.Date(1980, time.January, 27, 0, 0, 0, 0, time.UTC),
		},
		{
			Name:     "Adam Krummacker",
			Phone:    "+420 333 555 777",
			Birthday: time.Date(2009, time.March, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			Name:     "David Krummacker",
			Phone:    "+420 333 555 777",
			Birthday: time.Date(2011, time.December, 11, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, contact := range initialContacts {
		var inDB []Contact
		db.Where("Name = ?", contact.Name).Find(&inDB)
		if len(inDB) == 0 {
			db.Create(&contact)
		}
	}
}

// setupHttpRouter initializes the REST API router and registers all endpoints.
func setupHttpRouter() {
	router := gin.Default()
	router.GET("/contacts", findAllContacts)
	router.Run("localhost:8080")
}

// findAllContacts responds with the list of all contacts as JSON.
func findAllContacts(c *gin.Context) {
	var contacts []Contact
	db.Find(&contacts)
	c.IndentedJSON(http.StatusOK, contacts)
}
