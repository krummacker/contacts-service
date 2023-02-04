package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {

	// Create mock database and mock object
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Define expectations on SQL statements
	mock.ExpectPrepare("INSERT INTO contacts")
	mock.ExpectPrepare("SELECT \\* FROM contacts")
	mock.ExpectPrepare("SELECT \\* FROM contacts WHERE id=?")
	mock.ExpectPrepare("DELETE FROM contacts WHERE id=?")
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

	// Initialize contacts service
	setupDatabaseWrapper(db)
	gin.SetMode(gin.ReleaseMode)
	router := setupHttpRouter()

	// Run test
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/contacts/1", nil)
	router.ServeHTTP(recorder, request)

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
