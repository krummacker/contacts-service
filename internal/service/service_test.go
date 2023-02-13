package service

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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

// expectPreparedStatements instructs the mock object to expect that several statements are being
// prepared.
func expectPreparedStatements(mock sqlmock.Sqlmock) {
	mock.ExpectPrepare("INSERT INTO contacts")
	mock.ExpectPrepare("SELECT \\* FROM contacts")
	mock.ExpectPrepare("SELECT \\* FROM contacts WHERE id=?")
	mock.ExpectPrepare("DELETE FROM contacts WHERE id=?")
}

// expectSingleRowSelect instructs the mock object to expect that a select statement for a single
// contact will be executed.
func expectSingleRowSelect(mock sqlmock.Sqlmock, id int, name string, phone string, birthday time.Time) {
	rows := mock.NewRows([]string{"id", "name", "phone", "birthday"}).
		AddRow(id, name, phone, birthday)
	mock.ExpectQuery("SELECT \\* FROM contacts WHERE id=?").
		WithArgs(strconv.Itoa(id)).
		WillReturnRows(rows)
}

// initializeContactsService sets up the contacts service with the mock database and returns a
// handle to the gin engine against which requests can be executed.
func initializeContactsService(db *sql.DB) *gin.Engine {
	SetupDatabaseWrapper(db)
	gin.SetMode(gin.ReleaseMode)
	return SetupHttpRouter()
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

// TestGetAll executes a GET request for all contacts in the database. It expects that the JSON
// for a list of contacts is returned.
func TestGetAll(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)
	rows := mock.NewRows([]string{"id", "name", "phone", "birthday"}).
		AddRow(1, "Aaron", "+420 111", time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)).
		AddRow(2, "Berta", "+420 222", time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)).
		AddRow(3, "Carla", "+420 333", time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC))
	mock.ExpectQuery("SELECT \\* FROM contacts").
		WillReturnRows(rows)

	// Run test and compare results
	recorder := runTest(db, "GET", "/contacts", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var contacts []Contact
	json.Unmarshal(recorder.Body.Bytes(), &contacts)
	assert.Equal(t, 3, len(contacts))

	assert.Equal(t, int64(1), contacts[0].Id)
	assert.Equal(t, "Aaron", *contacts[0].Name)
	assert.Equal(t, "+420 111", *contacts[0].Phone)
	assert.Equal(t, time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC), *contacts[0].Birthday)

	assert.Equal(t, int64(2), contacts[1].Id)
	assert.Equal(t, "Berta", *contacts[1].Name)
	assert.Equal(t, "+420 222", *contacts[1].Phone)
	assert.Equal(t, time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC), *contacts[1].Birthday)

	assert.Equal(t, int64(3), contacts[2].Id)
	assert.Equal(t, "Carla", *contacts[2].Name)
	assert.Equal(t, "+420 333", *contacts[2].Phone)
	assert.Equal(t, time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC), *contacts[2].Birthday)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestGet executes a GET request for a single contact with a valid ID. It expects that the JSON
// for the contact is returned.
func TestGet(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)
	expectSingleRowSelect(mock,
		29,
		"Erika Mustermann",
		"+49 0815 4711",
		time.Date(1969, time.March, 2, 0, 0, 0, 0, time.UTC),
	)

	// Run test and compare results
	recorder := runTest(db, "GET", "/contacts/29", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)
	var getBody map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &getBody)
	assert.Equal(t, 29.0, getBody["id"])
	assert.Equal(t, "Erika Mustermann", getBody["name"])
	assert.Equal(t, "+49 0815 4711", getBody["phone"])
	assert.Equal(t, "1969-03-02T00:00:00Z", getBody["birthday"])
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestGetInvalidNumericID executes a GET request with an invalid but still numeric ID for a single
// contact. It expects that the HTTP request is answered with the NOT FOUND status code.
func TestGetInvalidNumericID(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)
	mock.ExpectQuery("SELECT \\* FROM contacts WHERE id=?").
		WithArgs("9999").
		WillReturnRows(mock.NewRows([]string{"id", "name", "phone", "birthday"}))

	// Run test and compare results
	recorder := runTest(db, "GET", "/contacts/9999", nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestGetInvalidCharacterID executes a GET request with an invalid ID consisting of characters.
// It expects that the HTTP request is answered with the NOT FOUND status code. It also expects
// that we do not reach out to the database in the first place.
func TestGetInvalidCharacterID(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)

	// Run test and compare results
	recorder := runTest(db, "GET", "/contacts/INVALID", nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPost executes a POST request with a valid body. It expects that the HTTP request is
// answered with the CREATED status code and a body with the posted values.
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
		WillReturnResult(sqlmock.NewResult(42, 1))

	// Run test and compare results
	recorder := runTest(db, "POST", "/contacts", strings.NewReader(`
		{
			"name": "Erika Mustermann", 
			"phone": "+49 0815 4711", 
			"birthday": "1969-03-04T00:00:00Z"
		}
	`))
	assert.Equal(t, http.StatusCreated, recorder.Code)
	var postBody map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &postBody)
	assert.Equal(t, "Erika Mustermann", postBody["name"])
	assert.Equal(t, "+49 0815 4711", postBody["phone"])
	assert.Equal(t, "1969-03-04T00:00:00Z", postBody["birthday"])
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPostInvalidBodies executes POST requests with invalid bodies. It expects that the HTTP
// requests are all answered with the BAD REQUEST status code.
func TestPostInvalidBodies(t *testing.T) {
	invalidRequestBodies := []string{
		"",
		"not JSON",
		`{
			"name": "Erika Mustermann"
			"phone": "+49 0815 4711"
			"birthday": "1969-03-02T00:00:00Z"
		}`, // commas missing
	}
	for _, body := range invalidRequestBodies {
		db, mock := createMockObjects(t)
		defer db.Close()

		// Define expectations on SQL statements
		expectPreparedStatements(mock) // we expect that the call will fail before the SQL statements

		// Run test and compare results
		recorder := runTest(db, "POST", "/contacts", strings.NewReader(body))
		assert.Equal(t, http.StatusBadRequest, recorder.Code, "request body: "+body)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

// TestPostEmptyJSON executes a POST request with a valid ID but an empty body. It expects
// that the HTTP request is answered with the OK status code, and that a contact with all
// fields having nil/null values is created.
func TestPostEmptyJSON(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)
	mock.ExpectExec("INSERT INTO contacts").
		WithArgs(nil, nil, nil).
		WillReturnResult(sqlmock.NewResult(49, 1))

	// Run test and compare results
	recorder := runTest(db, "POST", "/contacts", strings.NewReader("{}"))
	assert.Equal(t, http.StatusCreated, recorder.Code)
	var postBody map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &postBody)
	assert.Equal(t, nil, postBody["name"])
	assert.Equal(t, nil, postBody["phone"])
	assert.Equal(t, nil, postBody["birthday"])
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPut executes a PUT request with a valid ID and body. It expects that the HTTP request is
// answered with the OK status code and a body with all values of the contact.
func TestPut(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)
	mock.ExpectExec("UPDATE contacts").
		WithArgs(
			"Rudi Völler",
			"+49 1234567890",
			time.Date(1960, time.April, 13, 0, 0, 0, 0, time.UTC),
			"17",
		).
		WillReturnResult(sqlmock.NewResult(-1, 1))
	expectSingleRowSelect(mock,
		17,
		"Rudi Völler",
		"+49 1234567890",
		time.Date(1960, time.April, 13, 0, 0, 0, 0, time.UTC),
	)

	// Run test and compare results
	recorder := runTest(db, "PUT", "/contacts/17", strings.NewReader(`
		{
			"name": "Rudi Völler", 
			"phone": "+49 1234567890", 
			"birthday": "1960-04-13T00:00:00Z"
		}
	`))
	assert.Equal(t, http.StatusOK, recorder.Code)
	var postBody map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &postBody)
	assert.Equal(t, 17.0, postBody["id"])
	assert.Equal(t, "Rudi Völler", postBody["name"])
	assert.Equal(t, "+49 1234567890", postBody["phone"])
	assert.Equal(t, "1960-04-13T00:00:00Z", postBody["birthday"])
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPutPartial executes a PUT request with a valid ID and a valid body that contains only a
// subset of new values. It expects that the HTTP request is answered with the OK status code and a
// body with all values of the contact.
func TestPutPartial(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)
	mock.ExpectExec("UPDATE contacts").
		WithArgs(
			time.Date(1950, time.April, 13, 0, 0, 0, 0, time.UTC),
			"35",
		).
		WillReturnResult(sqlmock.NewResult(-1, 1))
	expectSingleRowSelect(mock,
		35,
		"Rudi Völler",
		"+49 1234567890",
		time.Date(1950, time.April, 13, 0, 0, 0, 0, time.UTC),
	)

	// Run test and compare results
	recorder := runTest(db, "PUT", "/contacts/35", strings.NewReader(`
		{
			"birthday": "1950-04-13T00:00:00Z"
		}
	`))
	assert.Equal(t, http.StatusOK, recorder.Code)
	var postBody map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &postBody)
	assert.Equal(t, 35.0, postBody["id"])
	assert.Equal(t, "Rudi Völler", postBody["name"])
	assert.Equal(t, "+49 1234567890", postBody["phone"])
	assert.Equal(t, "1950-04-13T00:00:00Z", postBody["birthday"])
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPutInvalidNumericID executes a PUT request with an invalid burt still numeric ID and
// otherwise valid body for a single contact. It expects that the HTTP request is answered with the
// NOT FOUND status code.
func TestPutInvalidNumericID(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)
	mock.ExpectExec("UPDATE contacts").
		WithArgs("Rudi Völler", "9999").
		WillReturnResult(sqlmock.NewResult(-1, 0))

	// Run test and compare results
	recorder := runTest(db, "PUT", "/contacts/9999", strings.NewReader(`
		{
			"name": "Rudi Völler"
		}
	`))
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPutInvalidCharacterID executes a PUT request with an invalid ID consisting of characters.
// It expects that the HTTP request is answered with the NOT FOUND status code. It also expects
// that we do not reach out to the database in the first place.
func TestPutInvalidCharacterID(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)

	// Run test and compare results
	recorder := runTest(db, "PUT", "/contacts/INVALID", strings.NewReader(`
		{
			"name": "Rudi Völler"
		}
	`))
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestPutInvalidBodies executes PUT requests with valid IDs but invalid bodies. It expects
// that the HTTP requests are all answered with the BAD REQUEST status code.
func TestPutInvalidBodies(t *testing.T) {
	invalidRequestBodies := []string{
		"",
		"{}",
		"not JSON",
		`{
			"name": "Erika Mustermann"
			"phone": "+49 0815 4711"
			"birthday": "1969-03-02T00:00:00Z"
		}`, // commas missing
	}
	for _, body := range invalidRequestBodies {
		db, mock := createMockObjects(t)
		defer db.Close()

		// Define expectations on SQL statements
		expectPreparedStatements(mock) // we expect that the call will fail before the SQL statements

		// Run test and compare results
		recorder := runTest(db, "PUT", "/contacts/1", strings.NewReader(body))
		assert.Equal(t, http.StatusBadRequest, recorder.Code, "request body: "+body)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

// TestDelete executes a DELETE request for a single contact with a valid ID. It expects that the
// status OK is returned.
func TestDelete(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)
	mock.ExpectExec("DELETE FROM contacts").
		WithArgs("42").
		WillReturnResult(sqlmock.NewResult(-1, 1))

	// Run test and compare results
	recorder := runTest(db, "DELETE", "/contacts/42", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestDeleteInvalidNumericID executes a DELETE request with an invalid but still numeric ID for a
// single contact. It expects that the HTTP request is answered with the NOT FOUND status code.
func TestDeleteInvalidNumericID(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)
	mock.ExpectExec("DELETE FROM contacts").
		WithArgs("9999").
		WillReturnResult(sqlmock.NewResult(-1, 0))

	// Run test and compare results
	recorder := runTest(db, "DELETE", "/contacts/9999", nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestDeleteInvalidCharacterID executes a DELETE request with an invalid ID consisting of
// characters. It expects that the HTTP request is answered with the NOT FOUND status code. It also
// expects that we do not reach out to the database in the first place.
func TestDeleteInvalidCharacterID(t *testing.T) {
	db, mock := createMockObjects(t)
	defer db.Close()

	// Define expectations on SQL statements
	expectPreparedStatements(mock)

	// Run test and compare results
	recorder := runTest(db, "DELETE", "/contacts/INVALID", nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
