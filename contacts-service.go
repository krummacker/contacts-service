package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
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

// constant for an unset date
var epoch time.Time

func main() {
	setupORMapper()
	db.AutoMigrate(&Contact{}) // Define database schema.
	populateDatabase()
	setupHttpRouter()
}

// setupORMapper initializes the object relational mapper and the database
// connection. The connection parameters are taken from the command line
// parameters.
//
// Usage example:
// > go run contacts-service.go -host=localhost -dbuser=dirk -dbpwd=bullo92
func setupORMapper() {

	hostp := flag.String("host", "localhost", "the host name of the database")
	dbuserp := flag.String("dbuser", "mysql", "the database user name")
	dbpwdp := flag.String("dbpwd", "", "the password of the database user")
	flag.Parse()

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/test?charset=utf8&parseTime=True&loc=Local",
		*dbuserp, *dbpwdp, *hostp)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
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
//
// Example API calls:
// > curl http://localhost:8080/contacts
// > curl http://localhost:8080/contacts --include --header "Content-Type: application/json" --request "POST" --data '{"Name": "Hans Wurst", "Phone": "0815", "Birthday": "1974-11-29T00:00:00+00:00"}'
// > curl http://localhost:8080/contacts/4
// > curl http://localhost:8080/contacts/5 --include --header "Content-Type: application/json" --request "PUT" --data '{"Phone": "81970"}'
func setupHttpRouter() {
	router := gin.Default()
	router.GET("/contacts", findAllContacts)
	router.POST("/contacts", createContact)
	router.GET("/contacts/:id", findContactByID)
	router.PUT("/contacts/:id", updateContactByID)
	router.Run("localhost:8080")
}

// findAllContacts responds with the list of all contacts as JSON.
func findAllContacts(c *gin.Context) {
	var contacts []Contact
	db.Find(&contacts)
	c.IndentedJSON(http.StatusOK, contacts)
}

// createContact inserts the contact specified in the request's JSON into the
// database. It responds with the full contact data including the newly
// assigned id.
func createContact(c *gin.Context) {
	var newContact Contact
	if err := c.BindJSON(&newContact); err != nil {
		// Bad request
		log.Panicln(err)
	}
	db.Create(&newContact)
	c.IndentedJSON(http.StatusCreated, newContact)
}

// findContactByID locates the contact whose ID value matches the id parameter
// of the request URL, then returns that contact as a response.
func findContactByID(c *gin.Context) {
	id := c.Param("id")
	var contact Contact
	var count int64
	db.First(&contact, id).Count(&count)
	if count == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "contact not found"})
	} else {
		c.IndentedJSON(http.StatusOK, contact)
	}
}

// updateContactByID updates the contact whose ID value matches the id
// parameter of the request URL, updates the values specified in the JSON (and
// only those), and finally responds with the new version of the contact.
func updateContactByID(c *gin.Context) {
	id := c.Param("id")

	var found Contact
	var count int64
	db.First(&found, id).Count(&count)
	if count == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "contact not found"})
		return
	}

	var submitted Contact
	if err := c.BindJSON(&submitted); err != nil {
		// Bad request
		log.Panicln(err)
	}

	if len(submitted.Name) > 0 {
		found.Name = submitted.Name
	}
	if len(submitted.Phone) > 0 {
		found.Phone = submitted.Phone
	}
	if submitted.Birthday != epoch {
		found.Birthday = submitted.Birthday
	}

	db.Save(&found)
	c.IndentedJSON(http.StatusCreated, found)
}
