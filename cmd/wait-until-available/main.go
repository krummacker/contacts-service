package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	for {
		res, err := http.Get("http://localhost:8080/contacts/")
		if err == nil && res.StatusCode == http.StatusOK {
			break
		}
		fmt.Println("Waiting 2 seconds")
		time.Sleep(2 * time.Second)
	}
}
