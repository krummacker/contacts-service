package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	totalWaitTime := 0
	for {
		res, err := http.Get("http://localhost:8080/contacts/")
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
