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
		if err == nil {
			if res.StatusCode == http.StatusOK {
				fmt.Println(res)
				break
			} else {
				fmt.Println(res)
			}
		} else {
			fmt.Println(err)
		}
		totalWaitTime += 5
		fmt.Printf("Waiting %d seconds", totalWaitTime)
		fmt.Println()
		time.Sleep(5 * time.Second)
	}
}
