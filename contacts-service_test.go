package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// createMockObjects builds a mock database handle and a mock object for defining our expected SQL
// calls.
func createMockObjects(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return db, mock
}

// expectPreparedStatements instructs the mock object to always expect thatin the beginning,
// several statements are being prepared.
func expectPreparedStatements(mock sqlmock.Sqlmock) {
	mock.ExpectPrepare("INSERT INTO contacts")
	mock.ExpectPrepare("SELECT \\* FROM contacts")
	mock.ExpectPrepare("SELECT \\* FROM contacts WHERE id=?")
	mock.ExpectPrepare("DELETE FROM contacts WHERE id=?")
}

// initializeContactsService sets up the contacts service with the mock database and returns a
// handle to the gin engine against which requests can be executed.
func initializeContactsService(db *sql.DB) *gin.Engine {
	setupDatabaseWrapper(db)
	gin.SetMode(gin.ReleaseMode)
	return setupHttpRouter()
}

// runTest executes the HTTP request with the specified arguments and returns the response.
func runTest(db *sql.DB, method string, url string, body *strings.Reader) *httptest.ResponseRecorder {
	router := initializeContactsService(db)
	recorder := httptest.NewRecorder()
	if body == nil {
		body = strings.NewReader("")
	}
	request, _ := http.NewRequest(method, url, body)
	router.ServeHTTP(recorder, request)
	return recorder
}

// TestGet executes a GET request for a single contact with a valid ID. It expects that the JSON
// for a contact is returned.
func TestGet(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)
	rows := mock.NewRows([]string{"id", "name", "phone", "birthday"}).
		AddRow(
			1,
			"Erika Mustermann",
			"+49 0815 4711",
			time.Date(1969, time.March, 2, 0, 0, 0, 0, time.UTC),
		)
	mock.ExpectQuery("SELECT \\* FROM contacts WHERE id=?").
		WithArgs("1").
		WillReturnRows(rows)

	recorder := runTest(db, "GET", "/contacts/1", nil)

	// Compare results
	assert.Equal(t, http.StatusOK, recorder.Code)
	var getBody map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &getBody)
	assert.Equal(t, 1.0, getBody["id"])
	assert.Equal(t, "Erika Mustermann", getBody["name"])
	assert.Equal(t, "+49 0815 4711", getBody["phone"])
	assert.Equal(t, "1969-03-02T00:00:00Z", getBody["birthday"])
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestGetInvalidID executes a GET request with an invalid ID for a single contact. It expects
// that the HTTP request is answered with the NOT FOUND status code.
func TestGetInvalidID(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)
	mock.ExpectQuery("SELECT \\* FROM contacts WHERE id=?").
		WithArgs("invalid").
		WillReturnRows(mock.NewRows([]string{"id", "name", "phone", "birthday"}))

	recorder := runTest(db, "GET", "/contacts/invalid", nil)

	// Compare results
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPost(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)
	mock.ExpectExec("INSERT INTO contacts").
		WithArgs(
			"Erika Mustermann",
			"+49 0815 4711",
			time.Date(1969, time.March, 4, 0, 0, 0, 0, time.UTC),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	recorder := runTest(db, "POST", "/contacts", strings.NewReader(`
		{
			"name": "Erika Mustermann", 
			"phone": "+49 0815 4711", 
			"birthday": "1969-03-04T00:00:00Z"
		}
	`))

	// Compare results
	assert.Equal(t, http.StatusCreated, recorder.Code)
	var postBody map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &postBody)
	assert.Equal(t, 1.0, postBody["id"])
	assert.Equal(t, "Erika Mustermann", postBody["name"])
	assert.Equal(t, "+49 0815 4711", postBody["phone"])
	assert.Equal(t, "1969-03-04T00:00:00Z", postBody["birthday"])
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
