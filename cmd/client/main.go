package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"gitlab.com/dirk.krummacker/contacts-service/internal/model"
	"gitlab.com/dirk.krummacker/contacts-service/internal/randomgen"
)

// serverPort is the port of the server that this client will contact.
var serverPort int

// Usage example on the command line:
// > PORT=8080 go run main.go
func main() {
	var err error
	serverPort, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		fmt.Println("could not parse PORT env variable", err)
		panic(err)
	}
	fmt.Println()
	fmt.Println("  Elements      POST       PUT       GET     FIRST      LAST      BOTH  BIRTHDAY    DELETE ")
	fmt.Println("-------------------------------------------------------------------------------------------")
	sizes := []int{1000, 5000, 10000, 50000}
	for _, loops := range sizes {
		firstID, _ := sendPostRequest(bytes.NewReader(CreateRandomContactJson()))
		fmt.Printf("%10d", loops)
		{
			// POST requests
			var duration int64
			for i := 0; i < loops; i++ {
				jsonBody := CreateRandomContactJson()
				_, d := sendPostRequest(bytes.NewReader(jsonBody))
				duration += d
			}
			fmt.Printf("%10d", duration/int64(loops*1000))
		}
		{
			// PUT requests
			f := func(id int64) int64 {
				jsonBody := CreateRandomContactJson()
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
			// GET requests specifying the beginning of the first name.
			// The result set is limited to 20 rows to not account for data transfer costs.
			var duration int64
			for i := 0; i < loops/1000; i++ {
				firstNameStart := randomgen.PickFirstName()[:3]
				requestURL := fmt.Sprintf("http://localhost:%d/contacts/?firstname=%s&limit=20", serverPort, firstNameStart)
				_, d := sendRequest(http.MethodGet, requestURL, nil)
				duration += d
			}
			fmt.Printf("%10d", duration/int64(loops))
		}
		{
			// GET requests specifying the beginning of the last name.
			// The result set is limited to 20 rows to not account for data transfer costs.
			var duration int64
			for i := 0; i < loops/1000; i++ {
				lastNameStart := randomgen.PickLastName()[:3]
				requestURL := fmt.Sprintf("http://localhost:%d/contacts/?lastname=%s&limit=20", serverPort, lastNameStart)
				_, d := sendRequest(http.MethodGet, requestURL, nil)
				duration += d
			}
			fmt.Printf("%10d", duration/int64(loops))
		}
		{
			// GET requests specifying the beginning of both the first and the last name.
			// The result set is limited to 20 rows to not account for data transfer costs.
			var duration int64
			for i := 0; i < loops/1000; i++ {
				firstNameStart := randomgen.PickFirstName()[:3]
				lastNameStart := randomgen.PickLastName()[:3]
				requestURL := fmt.Sprintf("http://localhost:%d/contacts/?lastname=%s&firstname=%s&limit=20", serverPort, lastNameStart, firstNameStart)
				_, d := sendRequest(http.MethodGet, requestURL, nil)
				duration += d
			}
			fmt.Printf("%10d", duration/int64(loops))
		}
		{
			// GET requests specifying a random birthday.
			// The result set is limited to 20 rows to not account for data transfer costs.
			var duration int64
			for i := 0; i < loops/1000; i++ {
				month := rand.Intn(12) + 1
				day := rand.Intn(28) + 1
				requestURL := fmt.Sprintf("http://localhost:%d/contacts/?birthday%d-%d&limit=20", serverPort, month, day)
				_, d := sendRequest(http.MethodGet, requestURL, nil)
				duration += d
			}
			fmt.Printf("%10d", duration/int64(loops))
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

func CreateRandomContactJson() []byte {
	result := []byte(`{
		"firstname": "` + randomgen.PickFirstName() + `",
		"lastname": "` + randomgen.PickLastName() + `",
		"phone": "` + randomgen.PickPhoneNumber("+1") + `",
		"birthday": "` + randomgen.PickBirthDate() + `"
	}`)
	return result
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
	var contact model.Contact
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
