package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const serverPort = 8080

type Contact struct {
	Id       int64      `json:"id"                 db:"id"`
	Name     *string    `json:"name,omitempty"     db:"name"`
	Phone    *string    `json:"phone,omitempty"    db:"phone"`
	Birthday *time.Time `json:"birthday,omitempty" db:"birthday"`
}

// Usage example on the command line:
// > go run main.go
func main() {
	fmt.Println()
	fmt.Println("  Elements      POST       PUT       GET    DELETE ")
	fmt.Println("---------------------------------------------------")
	sizes := []int{1000, 5000, 10000, 50000, 100000}
	jsonBody := []byte(`{
		"name": "Marcus Antonius",
		"phone": "+39 999 777 555",
		"birthday": "0027-11-09T00:00:00Z"
	}`)
	for _, loops := range sizes {
		firstID, _ := sendPostRequest(bytes.NewReader(jsonBody))
		fmt.Printf("%10d", loops)
		{
			// POST requests
			var duration int64
			for i := 0; i < loops; i++ {
				_, d := sendPostRequest(bytes.NewReader(jsonBody))
				duration += d
			}
			fmt.Printf("%10d", duration/int64(loops*1000))
		}
		{
			// PUT requests
			f := func(id int64) int64 {
				return sendPutGetDeleteRequest(id, http.MethodPut, bytes.NewReader(jsonBody))
			}
			callInLoop(firstID, loops, f)
		}
		{
			// GET requests
			f := func(id int64) int64 {
				return sendPutGetDeleteRequest(id, http.MethodGet, nil)
			}
			callInLoop(firstID, loops, f)
		}
		{
			// DELETE requests
			f := func(id int64) int64 {
				return sendPutGetDeleteRequest(id, http.MethodDelete, nil)
			}
			callInLoop(firstID, loops, f)
		}
		sendPutGetDeleteRequest(firstID, http.MethodDelete, nil)
		fmt.Println()
	}
}

func callInLoop(firstID int64, loops int, f func(id int64) int64) {
	ids := createRandomSliceWithIDs(firstID+1, loops)
	var duration int64
	for _, id := range ids {
		d := f(id)
		duration += d
	}
	fmt.Printf("%10d", duration/int64(loops*1000))
}

func createRandomSliceWithIDs(firstID int64, loops int) []int64 {
	ids := make([]int64, 0, loops)
	for i := 0; i < loops; i++ {
		ids = append(ids, firstID+int64(i))
	}
	rand.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})
	return ids
}

func sendPostRequest(bodyReader io.Reader) (int64, int64) {
	requestURL := fmt.Sprintf("http://localhost:%d/contacts", serverPort)
	resBody, duration := sendRequest(http.MethodPost, requestURL, bodyReader)
	var contact Contact
	err := json.Unmarshal(resBody, &contact)
	if err != nil {
		fmt.Println("could not unmarshal JSON", err)
		panic(err)
	}
	return contact.Id, duration
}

func sendPutGetDeleteRequest(id int64, method string, bodyReader io.Reader) int64 {
	requestURL := fmt.Sprintf("http://localhost:%d/contacts/%d", serverPort, id)
	_, duration := sendRequest(method, requestURL, bodyReader)
	return duration
}

func sendRequest(method string, requestURL string, bodyReader io.Reader) ([]byte, int64) {
	req, err := http.NewRequest(method, requestURL, bodyReader)
	if err != nil {
		fmt.Println("could not create request", err)
		panic(err)
	}
	before := time.Now().UnixNano()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("error making http request", err)
		panic(err)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("could not read response body", err)
		panic(err)
	}
	after := time.Now().UnixNano()
	return resBody, after - before
}
