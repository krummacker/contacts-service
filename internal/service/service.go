package service

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"gitlab.com/dirk.krummacker/contacts-service/internal/model"
)

// maxInt is the largest possible int value
const maxInt = int(^uint(0) >> 1)

// db is a handle to the database.
var db *sqlx.DB

// insert is a prepared statement for creating a contact on the database.
var insert *sqlx.NamedStmt

// selectWhereId is a prepared statement for selecting contacts with a given id.
var selectWhereId *sqlx.Stmt

// deleteWhereId is a prepared statement for deleting a contact with a given id.
var deleteWhereId *sqlx.Stmt

// allowedOrderby are the allowed values for the 'orderby' URL parameter.
var allowedOrderby = []string{"id", "firstname", "lastname", "phone", "birthday"}

// allowedAscending are the allowed values for the 'ascending' URL parameter.
var allowedAscending = []string{"true", "false"}

// CreateDatabase initializes and returns a database connection. The connection parameters are
// taken from the system's environment variables.
func CreateDatabase() *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/test?parseTime=true",
		os.Getenv("DBUSER"), os.Getenv("DBPWD"), os.Getenv("DBHOST"))
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	return sqlDB
}

// SetupDatabaseWrapper initializes the sqlx database wrapper with the specified sql database. It
// then prepares all statements. The database argument can be a real database for production use
// or a mock database within unit tests.
func SetupDatabaseWrapper(sqlDB *sql.DB) {
	var err error
	db = sqlx.NewDb(sqlDB, "mysql")

	// Prepared statements offer a significant speed increase if executed many times.
	insert, err = db.PrepareNamed(`
		INSERT INTO contacts (firstname, lastname, phone, birthday)
		VALUES (:firstname, :lastname, :phone, :birthday)
	`)
	if err != nil {
		log.Fatal(err)
	}
	selectWhereId, err = db.Preparex(`
		SELECT * FROM contacts WHERE id = ?
	`)
	if err != nil {
		log.Fatal(err)
	}
	deleteWhereId, err = db.Preparex(`
		DELETE FROM contacts WHERE id = ?
	`)
	if err != nil {
		log.Fatal(err)
	}
}

// SetupHttpRouter initializes the REST API router and registers all endpoints.
func SetupHttpRouter() *gin.Engine {
	var router *gin.Engine
	if strings.EqualFold(os.Getenv("GIN_LOGGING"), "off") {
		fmt.Println("Turning off HTTP request logging.")
		router = gin.New()
	} else {
		router = gin.Default()
	}
	router.GET("/contacts", findContacts)
	router.POST("/contacts", createContact)
	router.GET("/contacts/:id", findContactByID)
	router.PUT("/contacts/:id", updateContactByID)
	router.DELETE("/contacts/:id", deleteContactByID)
	return router
}

// findContacts responds with a list of contacts as JSON.
//
// The URL parameters 'firstname' and 'lastname' are interpreted as the beginning of the first name
// or last name of the contact.
//
// The URL parameter 'birthday' consists of a month part and a day part, separated by '-'. The call
// returns all contacts that have their birthday on this month and day, regardless of the year.
//
// The URL parameter 'limit' specifies how many contacts matching the search criteria are returned.
// The URL parameter 'offset' specifies how many items from the sorted list of results are skipped
// in the beginning. Together with the 'limit' parameter, one can implement search result paging.
//
// The URL parameter 'orderby' specifies the contact property by which the results shall be sorted.
// Valid values are 'id', 'firstname', 'lastname', 'phone', and 'birthday'. If this URL parameter
// is not specified, the contacts will be sorted by id.
//
// If the URL parameter 'ascending' is set to 'false' then the sort order is reversed, starting
// with the 'highest' value. If it is set to 'true', or if this URL parameter is omitted, the
// result starts with the lowest value.
//
// REST API calls:
//
//	> curl "http://localhost:8080/contacts"
//	> curl "http://localhost:8080/contacts?firstname=Ji"
//	> curl "http://localhost:8080/contacts?lastname=Smi"
//	> curl "http://localhost:8080/contacts?birthday=11-29"
//	> curl "http://localhost:8080/contacts?limit=20&offset=60"
//	> curl "http://localhost:8080/contacts?orderby=birthday&ascending=false"
func findContacts(c *gin.Context) {
	first, last, bday, bmonth, successNameAndBirthday := parseNameAndBirthday(c)
	if !successNameAndBirthday {
		return
	}
	limit, offset, successLimitAndOffset := parseLimitAndOffset(c)
	if !successLimitAndOffset {
		return
	}
	orderby, ascending, successOrderbyAndAscending := parseOrderbyAndAscending(c)
	if !successOrderbyAndAscending {
		return
	}
	var contacts []model.Contact
	var err error
	if (first != "" || last != "") && (bmonth != 0 || bday != 0) {
		sql := fmt.Sprintf(`
			SELECT *
			FROM contacts
			WHERE firstname LIKE ?
				AND lastname LIKE ?
				AND MONTH(birthday) = ?
				AND DAY(birthday) = ?
			ORDER BY %s %s
			LIMIT ?
			OFFSET ?`, orderby, ascending)
		err = db.Select(&contacts, sql, first+"%", last+"%", bmonth, bday, limit, offset)
	} else if (first != "" || last != "") && bmonth == 0 && bday == 0 {
		sql := fmt.Sprintf(`
			SELECT *
			FROM contacts
			WHERE firstname LIKE ?
				AND lastname LIKE ?
			ORDER BY %s %s
			LIMIT ?
			OFFSET ?`, orderby, ascending)
		err = db.Select(&contacts, sql, first+"%", last+"%", limit, offset)
	} else if first == "" && last == "" && (bmonth != 0 || bday != 0) {
		sql := fmt.Sprintf(`
			SELECT *
			FROM contacts
			WHERE MONTH(birthday) = ?
				AND DAY(birthday) = ?
			ORDER BY %s %s
			LIMIT ?
			OFFSET ?`, orderby, ascending)
		err = db.Select(&contacts, sql, bmonth, bday, limit, offset)
	} else {
		sql := fmt.Sprintf(`
			SELECT *
			FROM contacts
			ORDER BY %s %s
			LIMIT ?
			OFFSET ?`, orderby, ascending)
		err = db.Select(&contacts, sql, limit, offset)
	}
	if err != nil {
		log.Panicln(err)
	}
	if len(contacts) == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "contact not found"})
	} else {
		c.IndentedJSON(http.StatusOK, contacts)
	}
}

// parseNameAndBirthday inspects the URL parameters and determines values for first name, last
// name, day and month of the contact's birthday.
func parseNameAndBirthday(c *gin.Context) (firstname string, lastname string, bday int, bmonth int, success bool) {
	firstname = c.Query("firstname")
	lastname = c.Query("lastname")
	birthday := c.Query("birthday")
	if birthday != "" {
		var err error
		before, after, found := strings.Cut(birthday, "-")
		if !found {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid birthday URL parameter"})
			return "", "", 0, 0, false
		}
		bmonth, err = strconv.Atoi(before)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid birthday URL parameter"})
			return "", "", 0, 0, false
		}
		bday, err = strconv.Atoi(after)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid birthday URL parameter"})
			return "", "", 0, 0, false
		}
	}
	return firstname, lastname, bday, bmonth, true
}

// parseLimitAndOffset inspects the URL parameters and determines values for limit and offset of
// the result set.
func parseLimitAndOffset(c *gin.Context) (limit string, offset string, success bool) {
	limit = c.Query("limit")
	offset = c.Query("offset")
	if limit != "" {
		limitAsInt, errConv := strconv.Atoi(limit)
		if errConv != nil || limitAsInt < 1 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid limit parameter"})
			return "", "", false
		}
	} else {
		limit = strconv.Itoa(maxInt)
	}
	if offset != "" {
		offsetAsIt, errConv := strconv.Atoi(offset)
		if errConv != nil || offsetAsIt < 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid offset parameter"})
			return "", "", false
		}
	} else {
		offset = "0"
	}
	return limit, offset, true
}

// parseOrderbyAndAscending inspects the URL parameters and determines values for the orderby and
// ascending values of the result set.
func parseOrderbyAndAscending(c *gin.Context) (orderby string, ascending string, success bool) {
	orderby = c.Query("orderby")
	if orderby == "" {
		orderby = "id"
	}
	if !contains(allowedOrderby, orderby) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid orderby parameter"})
		return "", "", false
	}
	ascendingAsString := c.Query("ascending")
	if ascendingAsString == "" {
		ascendingAsString = "true"
	}
	if !contains(allowedAscending, ascendingAsString) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid ascending parameter"})
		return orderby, "", false
	}
	if ascendingAsString == "true" {
		ascending = "ASC"
	} else {
		ascending = "DESC"
	}
	return orderby, ascending, true
}

// contains returns true if a string is present in a slice.
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// createContact inserts the contact specified in the request's JSON into the database. It responds
// with the full contact data including the newly assigned id.
//
// Limitations:
// - If firstname, lastname or phone are not specified then an empty string is stored.
// - If birthday is not specified then January 1 in the year 1 AD is stored.
//
// Example REST API call:
//
//	> curl http://localhost:8080/contacts --request "POST" --include --header "Content-Type: application/json" --data '{"firstname": "Hans", "lastname": "Wurst", "phone": "0815", "birthday": "1969-03-02T00:00:00+00:00"}'
func createContact(c *gin.Context) {
	var newContact model.Contact
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
//
//	> curl http://localhost:8080/contacts/56
func findContactByID(c *gin.Context) {
	id := c.Param("id")
	_, errConv := strconv.Atoi(id)
	if errConv != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "invalid id parameter"})
		return
	}

	var contacts []model.Contact
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
//
//	> curl http://localhost:8080/contacts/56 --request "PUT" --include --header "Content-Type: application/json" --data '{"phone": "81970"}'
//	> curl http://localhost:8080/contacts/56 --request "PUT" --include --header "Content-Type: application/json" --data '{"birthday": "1972-06-06T00:00:00+00:00"}'
func updateContactByID(c *gin.Context) {
	id := c.Param("id")
	_, errConv := strconv.Atoi(id)
	if errConv != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "invalid id parameter"})
		return
	}

	var submitted model.Contact
	if errBind := c.BindJSON(&submitted); errBind != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON"})
		return
	}

	var args []interface{}
	sql := "UPDATE contacts SET "
	if submitted.FirstName != nil {
		args = append(args, submitted.FirstName)
		sql += "firstname=?, "
	}
	if submitted.LastName != nil {
		args = append(args, submitted.LastName)
		sql += "lastname=?, "
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
	var contacts []model.Contact
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
//
//	> curl http://localhost:8080/contacts/56 --request "DELETE"
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
