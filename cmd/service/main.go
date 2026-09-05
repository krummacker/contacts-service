package main

import (
	"fmt"
	"os"
	"strconv"

	"gitlab.com/dirk.krummacker/contacts-service/internal/service"
)

// Usage example on the command line:
// > PORT=8080 DBHOST=localhost DBUSER=dirk DBPWD=bullo92 GIN_MODE=release GIN_LOGGING=OFF go run main.go
func main() {
	if _, err := strconv.Atoi(os.Getenv("PORT")); err != nil {
		fmt.Println("could not parse PORT env variable", err)
		panic(err)
	}

	sqlDB := service.CreateDatabase()
	contactService, err := service.NewService(sqlDB)
	if err != nil {
		fmt.Println("could not initialize service", err)
		panic(err)
	}
	router := contactService.SetupHttpRouter()
	router.Run(":" + os.Getenv("PORT"))
}
