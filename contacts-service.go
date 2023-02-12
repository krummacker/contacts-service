package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Contact is the data structure for a person that we know.
// All fields with the exception of the Id field are optional.
type Contact struct {
	Id       int64      `json:"id"                 db:"id"`
	Name     *string    `json:"name,omitempty"     db:"name"`
	Phone    *string    `json:"phone,omitempty"    db:"phone"`
	Birthday *time.Time `json:"birthday,omitempty" db:"birthday"`
}

// db is a handle to the database.
var db *sqlx.DB

// insert is a prepared statement for creating a contact on the database.
var insert *sqlx.NamedStmt

// selectAll is a prepared statement for selecting all contacts.
var selectAll *sqlx.Stmt

// selectWhereId is a prepared statement for selecting contacts with a given id.
var selectWhereId *sqlx.Stmt

// deleteWhereId is a prepared statement for deleting a contact with a given id.
var deleteWhereId *sqlx.Stmt

// Usage example on the command line:
// > DBUSER=dirk DBPWD=bullo92 GIN_LOGGING=OFF go run contacts-service.go
func main() {
	sqlDB := createDatabase()
	setupDatabaseWrapper(sqlDB)
	router := setupHttpRouter()
	router.Run(":8080")
}

// createDatabase initializes and returns a database connection. The connection parameters are
// taken from the system's environment variables.
func createDatabase() *sql.DB {
	dsn := fmt.Sprintf("%s:%s@/test?parseTime=true", os.Getenv("DBUSER"), os.Getenv("DBPWD"))
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	return sqlDB
}

// setupDatabaseWrapper initializes the sqlx database wrapper with the specified sql database. It
// then prepares all statements. The database argument can be a real database for production use
// or a mock database within unit tests.
func setupDatabaseWrapper(sqlDB *sql.DB) {
	var err error
	db = sqlx.NewDb(sqlDB, "mysql")

	// Prepared statements offer a significant speed increase if executed many times.
	insert, err = db.PrepareNamed(`
		INSERT INTO contacts (name, phone, birthday)
		VALUES (:name, :phone, :birthday)
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
	deleteWhereId, err = db.Preparex(`
		DELETE FROM contacts WHERE id=?
	`)
	if err != nil {
		log.Fatal(err)
	}
}

// setupHttpRouter initializes the REST API router and registers all endpoints.
func setupHttpRouter() *gin.Engine {
	var router *gin.Engine
	if strings.EqualFold(os.Getenv("GIN_LOGGING"), "off") {
		fmt.Println("Turning off HTTP request logging.")
		router = gin.New()
	} else {
		router = gin.Default()
	}
	router.GET("/contacts", findAllContacts)
	router.POST("/contacts", createContact)
	router.GET("/contacts/:id", findContactByID)
	router.PUT("/contacts/:id", updateContactByID)
	router.DELETE("/contacts/:id", deleteContactByID)
	return router
}

// findAllContacts responds with the list of all contacts as JSON.
//
// Example REST API call:
// > curl http://localhost:8080/contacts
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

// createContact inserts the contact specified in the request's JSON into the database. It responds
// with the full contact data including the newly assigned id.
//
// Limitations:
// - If name or phone are not specified then an empty string is stored.
// - If birthday is not specified then January 1 in the year 1 AD is stored.
//
// Example REST API call:
// > curl http://localhost:8080/contacts --request "POST" --include --header "Content-Type: application/json" --data '{"name": "Hans Wurst", "phone": "0815", "birthday": "1969-03-02T00:00:00+00:00"}'
func createContact(c *gin.Context) {
	var newContact Contact
	if err := c.BindJSON(&newContact); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON"})
		return
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

// findContactByID locates the contact whose ID value matches the id parameter of the request URL,
// then returns that contact as a response.
//
// Example REST API call:
// > curl http://localhost:8080/contacts/56
func findContactByID(c *gin.Context) {
	id := c.Param("id")
	_, errConv := strconv.Atoi(id)
	if errConv != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "invalid id parameter"})
		return
	}

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

// updateContactByID updates the contact whose ID value matches the id parameter of the request
// URL, updates the values specified in the JSON (and only those), and finally responds with the
// new version of the contact.
//
// Example REST API calls:
// > curl http://localhost:8080/contacts/56 --request "PUT" --include --header "Content-Type: application/json" --data '{"phone": "81970"}'
// > curl http://localhost:8080/contacts/56 --request "PUT" --include --header "Content-Type: application/json" --data '{"birthday": "1972-06-06T00:00:00+00:00"}'
func updateContactByID(c *gin.Context) {
	id := c.Param("id")
	_, errConv := strconv.Atoi(id)
	if errConv != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "invalid id parameter"})
		return
	}

	var submitted Contact
	if errBind := c.BindJSON(&submitted); errBind != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON"})
		return
	}

	var args []interface{}
	sql := "UPDATE contacts SET "
	if submitted.Name != nil {
		args = append(args, submitted.Name)
		sql += "name=?, "
	}
	if submitted.Phone != nil {
		args = append(args, submitted.Phone)
		sql += "phone=?, "
	}
	if submitted.Birthday != nil {
		args = append(args, &submitted.Birthday)
		sql += "birthday=?, "
	}

	// It only makes sense to continue if we have at least one value to update.
	if len(args) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "no values to be updated"})
		return
	}

	sql = sql[:len(sql)-2]
	sql += " WHERE id=?"
	args = append(args, id)
	result := db.MustExec(sql, args...)
	rowsAffected, errRows := result.RowsAffected()
	if errRows != nil {
		log.Panicln(errRows)
	}
	if rowsAffected == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "contact not found"})
		return
	}

	// In the HTTP response, return the full contact after the update.
	var contacts []Contact
	errSelect := selectWhereId.Select(&contacts, id)
	if errSelect != nil {
		log.Panicln(errRows)
	}
	if len(contacts) == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "contact not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, contacts[0])
}

// deleteContactByID deletes the contact whose ID value matches the id parameter of the request URL
// from the database.
//
// Example REST API call:
// > curl http://localhost:8080/contacts/56 --request "DELETE"
func deleteContactByID(c *gin.Context) {
	id := c.Param("id")
	_, error := strconv.Atoi(id)
	if error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "invalid id parameter"})
		return
	}

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
