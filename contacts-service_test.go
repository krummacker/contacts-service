package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestContactHappyPath tests a POST, GET, PUT, and DELETE with valid data.
//
// Usage: DBUSER=dirk DBPWD=bullo92 GIN_MODE=release go test
func TestContactHappyPath(t *testing.T) {
	setupDatabase()
	router := setupHttpRouter()

	// test the endpoint for creating a contact
	postRecorder := httptest.NewRecorder()
	postRequest, _ := http.NewRequest("POST", "/contacts", strings.NewReader(`
		{
			"Name": "Erika Mustermann", 
			"Phone": "+49 0815 4711", 
			"Birthday": "1969-03-02T00:00:00Z"
		}
	`))
	router.ServeHTTP(postRecorder, postRequest)
	assert.Equal(t, http.StatusCreated, postRecorder.Code)
	var postBody map[string]interface{}
	json.Unmarshal(postRecorder.Body.Bytes(), &postBody)
	assert.Equal(t, "Erika Mustermann", postBody["Name"])
	assert.Equal(t, "+49 0815 4711", postBody["Phone"])
	assert.Equal(t, "1969-03-02T00:00:00Z", postBody["Birthday"])
	idAsFloat64 := postBody["Id"]
	idAsString := fmt.Sprintf("%.0f", idAsFloat64)

	// test the endpoint for finding a contact
	getRecorder := httptest.NewRecorder()
	getRequest, _ := http.NewRequest("GET", "/contacts/"+idAsString, nil)
	router.ServeHTTP(getRecorder, getRequest)
	assert.Equal(t, http.StatusOK, getRecorder.Code)
	var getBody map[string]interface{}
	json.Unmarshal(getRecorder.Body.Bytes(), &getBody)
	assert.Equal(t, idAsFloat64, getBody["Id"])
	assert.Equal(t, "Erika Mustermann", getBody["Name"])
	assert.Equal(t, "+49 0815 4711", getBody["Phone"])
	assert.Equal(t, "1969-03-02T00:00:00Z", getBody["Birthday"])

	// test the endpoint for updating a contact
	putRecorder := httptest.NewRecorder()
	putRequest, _ := http.NewRequest("PUT", "/contacts/"+idAsString, strings.NewReader(`
		{
			"Name": "Rudi Völler", 
			"Phone": "+49 1234567890", 
			"Birthday": "1960-04-13T00:00:00Z"
		}
	`))
	router.ServeHTTP(putRecorder, putRequest)
	assert.Equal(t, http.StatusOK, putRecorder.Code)
	var putBody map[string]interface{}
	json.Unmarshal(putRecorder.Body.Bytes(), &putBody)
	assert.Equal(t, idAsFloat64, putBody["Id"])
	assert.Equal(t, "Rudi Völler", putBody["Name"])
	assert.Equal(t, "+49 1234567890", putBody["Phone"])
	assert.Equal(t, "1960-04-13T00:00:00Z", putBody["Birthday"])

	// test if a subsequent lookup of the contact returns the updated values
	getAgainRecorder := httptest.NewRecorder()
	getAgainRequest, _ := http.NewRequest("GET", "/contacts/"+idAsString, nil)
	router.ServeHTTP(getAgainRecorder, getAgainRequest)
	assert.Equal(t, http.StatusOK, getAgainRecorder.Code)
	var getAgainBody map[string]interface{}
	json.Unmarshal(getAgainRecorder.Body.Bytes(), &getAgainBody)
	assert.Equal(t, idAsFloat64, getAgainBody["Id"])
	assert.Equal(t, "Rudi Völler", getAgainBody["Name"])
	assert.Equal(t, "+49 1234567890", getAgainBody["Phone"])
	assert.Equal(t, "1960-04-13T00:00:00Z", getAgainBody["Birthday"])

	// test the endpoint for deleting a contact
	deleteRecorder := httptest.NewRecorder()
	deleteRequest, _ := http.NewRequest("DELETE", "/contacts/"+idAsString, nil)
	router.ServeHTTP(deleteRecorder, deleteRequest)
	assert.Equal(t, http.StatusOK, deleteRecorder.Code)

	// test if a final lookup of the contact will correctly not find it
	getFinalRecorder := httptest.NewRecorder()
	getFinalRequest, _ := http.NewRequest("GET", "/contacts/"+idAsString, nil)
	router.ServeHTTP(getFinalRecorder, getFinalRequest)
	assert.Equal(t, http.StatusNotFound, getFinalRecorder.Code)
}

// TestCreateContactInvalidBody tests a POST with different forms of invalid request body data.
func TestCreateContactInvalidBody(t *testing.T) {
	invalidRequestBodies := []string{
		"",
		"not JSON",
		`{
			"Name": "Erika Mustermann"
			"Phone": "+49 0815 4711"
			"Birthday": "1969-03-02T00:00:00Z"
		}`, // commas missing
	}

	setupDatabase()
	router := setupHttpRouter()
	for _, body := range invalidRequestBodies {
		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest("POST", "/contacts", strings.NewReader(body))
		router.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusBadRequest, recorder.Code, "request body: "+body)
	}
}

// TestCreateContactEmptyJSON tests a POST with an empty JSON which must create a contact with all
// fields having the default values.
func TestCreateContactEmptyJSON(t *testing.T) {
	setupDatabase()
	router := setupHttpRouter()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/contacts", strings.NewReader("{}"))
	router.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusCreated, recorder.Code)
	var body map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &body)
	assert.Equal(t, "", body["Name"])
	assert.Equal(t, "", body["Phone"])
	assert.Equal(t, "0001-01-01T00:00:01Z", body["Birthday"])
}
