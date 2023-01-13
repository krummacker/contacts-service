package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Contact is the data structure for a person that we know.
type Contact struct {
	Id    int
	Name  string
	Phone string
}

// db is a handle to the database.
var db *sqlx.DB
var err error

func main() {
	setupDatabase()
	populateDatabase()
	setupHttpRouter()
}

// setupDatabase initializes the database connection. The connection parameters
// are taken from the command line parameters.
//
// Usage example:
// > go run contacts-service.go -dbuser=dirk -dbpwd=bullo92
func setupDatabase() {

	dbuserp := flag.String("dbuser", "mysql", "the database user name")
	dbpwdp := flag.String("dbpwd", "", "the password of the database user")
	flag.Parse()

	dsn := fmt.Sprintf("%s:%s@/test", *dbuserp, *dbpwdp)
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalln(err)
	}
}

// populateDatabase enters initial test data into the database. If the data is
// already present in the table then it is not added again.
func populateDatabase() {
	initialContacts := []Contact{
		{
			Name:  "Dirk Krummacker",
			Phone: "+420 123 456 789",
		},
		{
			Name:  "Pavla Krummackerova",
			Phone: "+420 023 454 244",
		},
		{
			Name:  "Adam Krummacker",
			Phone: "+420 333 555 777",
		},
		{
			Name:  "David Krummacker",
			Phone: "+420 333 555 777",
		},
	}
	for _, contact := range initialContacts {
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM contacts WHERE name=?", contact.Name)
		if err != nil {
			log.Panicln(err)
		}
		if count == 0 {
			db.NamedExec("INSERT INTO contacts (name, phone) VALUES (:name, :phone)", &contact)
		}
	}
}

// setupHttpRouter initializes the REST API router and registers all endpoints.
//
// Example API calls:
// > curl http://localhost:8080/contacts
// > curl http://localhost:8080/contacts --include --header "Content-Type: application/json" --request "POST" --data '{"Name": "Hans Wurst", "Phone": "0815"}'
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
	err := db.Select(&contacts, "SELECT id, name, phone FROM contacts")
	if err != nil {
		log.Panicln(err)
	}
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
	_, err := db.NamedExec("INSERT INTO contacts (name, phone) VALUES (:name, :phone)", &newContact)
	if err != nil {
		log.Panicln(err)
	}
	c.IndentedJSON(http.StatusCreated, newContact)
}

// findContactByID locates the contact whose ID value matches the id parameter
// of the request URL, then returns that contact as a response.
func findContactByID(c *gin.Context) {
	id := c.Param("id")
	var contacts []Contact
	err := db.Select(&contacts, "SELECT id, name, phone FROM contacts WHERE id=?", id)
	if err != nil {
		log.Panicln(err)
	}
	if len(contacts) == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "contact not found"})
	} else {
		c.IndentedJSON(http.StatusOK, contacts[0])
	}
}

// updateContactByID updates the contact whose ID value matches the id
// parameter of the request URL, updates the values specified in the JSON (and
// only those), and finally responds with the new version of the contact.
func updateContactByID(c *gin.Context) {
	id := c.Param("id")
	var contacts []Contact
	err := db.Select(&contacts, "SELECT id, name, phone FROM contacts WHERE id=?", id)
	if err != nil {
		log.Panicln(err)
	}
	if len(contacts) == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "contact not found"})
		return
	}
	found := contacts[0]

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

	db.MustExec("UPDATE contacts SET name=?, phone=? WHERE id=?",
		found.Name, found.Phone, id)
	c.IndentedJSON(http.StatusCreated, found)
}
