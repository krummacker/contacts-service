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
	sqlDB := service.CreateDatabase()
	service.SetupDatabaseWrapper(sqlDB)
	router := service.SetupHttpRouter()
	_, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		fmt.Println("could not parse PORT env variable", err)
		panic(err)
	}
	router.Run(":" + os.Getenv("PORT"))
}
