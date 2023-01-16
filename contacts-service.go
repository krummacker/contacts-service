package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Contact is the data structure for a person that we know.
type Contact struct {
	Id       int64
	Name     string
	Phone    string
	Birthday time.Time
}

// db is a handle to the database.
var db *sqlx.DB

// insert is a prepared statement for creating a contact on the database.
var insert *sqlx.NamedStmt

// countWhereName is a prepared statement for counting contacts with a given
// name.
var countWhereName *sqlx.Stmt

// selectAll is a prepared statement for selecting all contacts.
var selectAll *sqlx.Stmt

// selectWhereId is a prepared statement for selecting contacts with a given
// id.
var selectWhereId *sqlx.Stmt

// updateWhereId is a prepared statement for updating a contact with a given
// id.
var updateWhereId *sqlx.Stmt

// deleteWhereId is a prepared statement for deleting a contact with a given
// id.
var deleteWhereId *sqlx.Stmt

// constant for an unset date
var epoch time.Time

func main() {
	setupDatabase()
	populateDatabase()
	setupHttpRouter()
}

// setupDatabase initializes the database connection and prepares all
// statements. The connection parameters are taken from the command line
// parameters.
//
// Usage example:
// > go run contacts-service.go -dbuser=dirk -dbpwd=bullo92
func setupDatabase() {

	dbuserp := flag.String("dbuser", "mysql", "the database user name")
	dbpwdp := flag.String("dbpwd", "", "the password of the database user")
	flag.Parse()

	dsn := fmt.Sprintf("%s:%s@/test?parseTime=true", *dbuserp, *dbpwdp)
	var err error
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	// Prepared statements offer up to 25% speed increase if executed many times.
	insert, err = db.PrepareNamed(`
		INSERT INTO contacts (name, phone, birthday)
		VALUES (:name, :phone, :birthday)
	`)
	if err != nil {
		log.Fatal(err)
	}
	countWhereName, err = db.Preparex(`
		SELECT COUNT(name) FROM contacts WHERE name=?
	`)
	if err != nil {
		log.Fatal(err)
	}
	selectAll, err = db.Preparex(`
		SELECT * FROM contacts
	`)
	if err != nil {
		log.Fatal(err)
	}
	selectWhereId, err = db.Preparex(`
		SELECT * FROM contacts WHERE id=?
	`)
	if err != nil {
		log.Fatal(err)
	}
	updateWhereId, err = db.Preparex(`
		UPDATE contacts SET name=?, phone=?, birthday=? WHERE id=?
	`)
	if err != nil {
		log.Fatal(err)
	}
	deleteWhereId, err = db.Preparex(`
		DELETE FROM contacts WHERE id=?
	`)
	if err != nil {
		log.Fatal(err)
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
		var count int
		err := countWhereName.Get(&count, contact.Name)
		if err != nil {
			log.Panicln(err)
		}
		if count == 0 {
			insert.Exec(&contact)
		}
	}
}

// setupHttpRouter initializes the REST API router and registers all endpoints.
//
// Example API calls:
// > curl http://localhost:8080/contacts
// > curl http://localhost:8080/contacts --request "POST" --include --header "Content-Type: application/json" --data '{"Name": "Hans Wurst", "Phone": "0815", "Birthday": "1969-03-02T00:00:00+00:00"}'
// > curl http://localhost:8080/contacts/56
// > curl http://localhost:8080/contacts/56 --request "PUT" --include --header "Content-Type: application/json" --data '{"Phone": "81970"}'
// > curl http://localhost:8080/contacts/56 --request "PUT" --include --header "Content-Type: application/json" --data '{"Birthday": "1972-06-06T00:00:00+00:00"}'
// > curl http://localhost:8080/contacts/56 --request "DELETE"
func setupHttpRouter() {
	router := gin.Default()
	router.GET("/contacts", findAllContacts)
	router.POST("/contacts", createContact)
	router.GET("/contacts/:id", findContactByID)
	router.PUT("/contacts/:id", updateContactByID)
	router.DELETE("/contacts/:id", deleteContactByID)
	router.Run("localhost:8080")
}

// findAllContacts responds with the list of all contacts as JSON.
func findAllContacts(c *gin.Context) {
	var contacts []Contact
	err := selectAll.Select(&contacts)
	if err != nil {
		log.Panicln(err)
	}
	if len(contacts) == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "contact not found"})
	} else {
		c.IndentedJSON(http.StatusOK, contacts)
	}
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
	result, err := insert.Exec(&newContact)
	if err != nil {
		log.Panicln(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Panicln(err)
	}
	newContact.Id = id
	c.IndentedJSON(http.StatusCreated, newContact)
}

// findContactByID locates the contact whose ID value matches the id parameter
// of the request URL, then returns that contact as a response.
func findContactByID(c *gin.Context) {
	id := c.Param("id")
	var contacts []Contact
	err := selectWhereId.Select(&contacts, id)
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
	err := selectWhereId.Select(&contacts, id)
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
	if submitted.Birthday != epoch {
		found.Birthday = submitted.Birthday
	}

	_, err = updateWhereId.Exec(found.Name, found.Phone, found.Birthday, id)
	if err != nil {
		log.Panicln(err)
	}
	c.IndentedJSON(http.StatusCreated, found)
}

// deleteContactByID deletes the contact whose ID value matches the id
// parameter of the request URL from the database.
func deleteContactByID(c *gin.Context) {
	id := c.Param("id")
	result, err := deleteWhereId.Exec(id)
	if err != nil {
		log.Panicln(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Panicln(err)
	}
	if rowsAffected == 1 {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "contact deleted"})
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "contact not found"})
	}
}
