package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Usage example on the command line:
// > PORT=8080 go run main.go
func main() {
	serverPort, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		fmt.Println("could not parse PORT env variable", err)
		panic(err)
	}
	requestURL := fmt.Sprintf("http://localhost:%d/contacts", serverPort)
	totalWaitTime := 0
	for {
		res, err := http.Get(requestURL) // nosemgrep
		if err != nil {
			fmt.Printf("Received error: " + err.Error())
			fmt.Println()
		} else {
			fmt.Printf("Received status code: %d", res.StatusCode)
			fmt.Println()
			if res.StatusCode == http.StatusNotFound {
				break
			}
		}
		totalWaitTime += 5
		fmt.Printf("Waiting in total %d seconds...", totalWaitTime)
		fmt.Println()
		time.Sleep(5 * time.Second)
	}
}
