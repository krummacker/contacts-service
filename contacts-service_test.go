package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Usage via command line: DBUSER=dirk DBPWD=bullo92 GIN_MODE=release go test -v

// TestContactHappyPath tests a POST, GET, PUT, and DELETE with valid data.
func TestContactHappyPath(t *testing.T) {
	setupDatabase()
	router := setupHttpRouter()

	// test the endpoint for creating a contact
	postRecorder := httptest.NewRecorder()
	postRequest, _ := http.NewRequest("POST", "/contacts", strings.NewReader(`
		{
			"name": "Erika Mustermann", 
			"phone": "+49 0815 4711", 
			"birthday": "1969-03-02T00:00:00Z"
		}
	`))
	router.ServeHTTP(postRecorder, postRequest)
	assert.Equal(t, http.StatusCreated, postRecorder.Code)
	var postBody map[string]interface{}
	json.Unmarshal(postRecorder.Body.Bytes(), &postBody)
	assert.Equal(t, "Erika Mustermann", postBody["name"])
	assert.Equal(t, "+49 0815 4711", postBody["phone"])
	assert.Equal(t, "1969-03-02T00:00:00Z", postBody["birthday"])
	idAsFloat64 := postBody["id"]
	idAsString := fmt.Sprintf("%.0f", idAsFloat64)

	// test the endpoint for finding a contact
	getRecorder := httptest.NewRecorder()
	getRequest, _ := http.NewRequest("GET", "/contacts/"+idAsString, nil)
	router.ServeHTTP(getRecorder, getRequest)
	assert.Equal(t, http.StatusOK, getRecorder.Code)
	var getBody map[string]interface{}
	json.Unmarshal(getRecorder.Body.Bytes(), &getBody)
	assert.Equal(t, idAsFloat64, getBody["id"])
	assert.Equal(t, "Erika Mustermann", getBody["name"])
	assert.Equal(t, "+49 0815 4711", getBody["phone"])
	assert.Equal(t, "1969-03-02T00:00:00Z", getBody["birthday"])

	// test the endpoint for updating a contact
	putRecorder := httptest.NewRecorder()
	putRequest, _ := http.NewRequest("PUT", "/contacts/"+idAsString, strings.NewReader(`
		{
			"name": "Rudi Völler", 
			"phone": "+49 1234567890", 
			"birthday": "1960-04-13T00:00:00Z"
		}
	`))
	router.ServeHTTP(putRecorder, putRequest)
	assert.Equal(t, http.StatusOK, putRecorder.Code)
	var putBody map[string]interface{}
	json.Unmarshal(putRecorder.Body.Bytes(), &putBody)
	assert.Equal(t, idAsFloat64, putBody["id"])
	assert.Equal(t, "Rudi Völler", putBody["name"])
	assert.Equal(t, "+49 1234567890", putBody["phone"])
	assert.Equal(t, "1960-04-13T00:00:00Z", putBody["birthday"])

	// test if a subsequent lookup of the contact returns the updated values
	getAgainRecorder := httptest.NewRecorder()
	getAgainRequest, _ := http.NewRequest("GET", "/contacts/"+idAsString, nil)
	router.ServeHTTP(getAgainRecorder, getAgainRequest)
	assert.Equal(t, http.StatusOK, getAgainRecorder.Code)
	var getAgainBody map[string]interface{}
	json.Unmarshal(getAgainRecorder.Body.Bytes(), &getAgainBody)
	assert.Equal(t, idAsFloat64, getAgainBody["id"])
	assert.Equal(t, "Rudi Völler", getAgainBody["name"])
	assert.Equal(t, "+49 1234567890", getAgainBody["phone"])
	assert.Equal(t, "1960-04-13T00:00:00Z", getAgainBody["birthday"])

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
			"name": "Erika Mustermann"
			"phone": "+49 0815 4711"
			"birthday": "1969-03-02T00:00:00Z"
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
// fields having nil/null values.
func TestCreateContactEmptyJSON(t *testing.T) {
	setupDatabase()
	router := setupHttpRouter()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/contacts", strings.NewReader("{}"))
	router.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusCreated, recorder.Code)
	var body map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &body)
	assert.Nil(t, body["name"])
	assert.Nil(t, body["phone"])
	assert.Nil(t, body["birthday"])
}

// TestUpdateContactInvalidId tests a PUT with an invalid id.
func TestUpdateContactInvalidId(t *testing.T) {
	setupDatabase()
	router := setupHttpRouter()
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("PUT", "/contacts/invalid", strings.NewReader(`
		{
			"name": "Rudi Völler", 
			"phone": "+49 1234567890", 
			"birthday": "1960-04-13T00:00:00Z"
		}
	`))
	router.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

// TestUpdateContactInvalidBody tests a PUT with a valid id but an invalid request body.
func TestUpdateContactInvalidBody(t *testing.T) {
	setupDatabase()
	router := setupHttpRouter()

	postRecorder := httptest.NewRecorder()
	postRequest, _ := http.NewRequest("POST", "/contacts", strings.NewReader("{}"))
	router.ServeHTTP(postRecorder, postRequest)
	assert.Equal(t, http.StatusCreated, postRecorder.Code)
	var postBody map[string]interface{}
	json.Unmarshal(postRecorder.Body.Bytes(), &postBody)
	idAsFloat64 := postBody["id"]
	idAsString := fmt.Sprintf("%.0f", idAsFloat64)

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
		putRecorder := httptest.NewRecorder()
		putRequest, _ := http.NewRequest("PUT", "/contacts/"+idAsString, strings.NewReader(body))
		router.ServeHTTP(putRecorder, putRequest)
		assert.Equal(t, http.StatusBadRequest, putRecorder.Code)
	}
}

// TestUpdateContactPartially tests a PUT with only one field specified in the JSON. It verifies
// that the other fields are still nil.
func TestUpdateContactPartially(t *testing.T) {
	setupDatabase()
	router := setupHttpRouter()

	postRecorder := httptest.NewRecorder()
	postRequest, _ := http.NewRequest("POST", "/contacts", strings.NewReader("{}"))
	router.ServeHTTP(postRecorder, postRequest)
	assert.Equal(t, http.StatusCreated, postRecorder.Code)
	var postBody map[string]interface{}
	json.Unmarshal(postRecorder.Body.Bytes(), &postBody)
	idAsFloat64 := postBody["id"]
	idAsString := fmt.Sprintf("%.0f", idAsFloat64)

	putRecorder := httptest.NewRecorder()
	putRequest, _ := http.NewRequest("PUT", "/contacts/"+idAsString, strings.NewReader(`
		{
			"name": "Rudi Völler"
		}
	`))
	router.ServeHTTP(putRecorder, putRequest)
	assert.Equal(t, http.StatusOK, putRecorder.Code)
	var putBody map[string]interface{}
	json.Unmarshal(putRecorder.Body.Bytes(), &putBody)
	assert.Equal(t, idAsFloat64, putBody["id"])
	assert.Equal(t, "Rudi Völler", putBody["name"])
	assert.Nil(t, putBody["phone"])
	assert.Nil(t, putBody["birthday"])
}

// TestFindAllContacts retrieves all contacts and verifies that a previously created contact is
// among them.
func TestFindAllContacts(t *testing.T) {
	setupDatabase()
	router := setupHttpRouter()

	postRecorder := httptest.NewRecorder()
	postRequest, _ := http.NewRequest("POST", "/contacts", strings.NewReader(`
		{
			"name": "Julius Cäsar", 
			"phone": "+39 123 456 789", 
			"birthday": "0057-07-01T00:00:00Z"
		}
	`))
	router.ServeHTTP(postRecorder, postRequest)
	assert.Equal(t, http.StatusCreated, postRecorder.Code)
	var postBody map[string]interface{}
	json.Unmarshal(postRecorder.Body.Bytes(), &postBody)
	idFromPost := int64(math.Round(postBody["id"].(float64)))

	// test the endpoint for finding all contacts
	getRecorder := httptest.NewRecorder()
	getRequest, _ := http.NewRequest("GET", "/contacts", nil)
	router.ServeHTTP(getRecorder, getRequest)
	assert.Equal(t, http.StatusOK, getRecorder.Code)
	var contacts []Contact
	json.Unmarshal(getRecorder.Body.Bytes(), &contacts)
	var found bool
	for _, contact := range contacts {
		if contact.Id == idFromPost {
			assert.Equal(t, "Julius Cäsar", *contact.Name)
			assert.Equal(t, "+39 123 456 789", *contact.Phone)
			assert.Equal(t, time.Date(57, time.July, 1, 0, 0, 0, 0, time.UTC), *contact.Birthday)
			found = true
		}
	}
	assert.True(t, found, "could not find contact")
}

// TestFindContactInvalidId tests a GET with an invalid id.
func TestFindContactInvalidId(t *testing.T) {
	setupDatabase()
	router := setupHttpRouter()
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/contacts/invalid", nil)
	router.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}
