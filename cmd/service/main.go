package main

import (
	"gitlab.com/dirk.krummacker/contacts-service/internal/service"
)

// Usage example on the command line:
// > DBUSER=dirk DBPWD=bullo92 GIN_MODE=release GIN_LOGGING=OFF go run main.go
func main() {
	sqlDB := service.CreateDatabase()
	service.SetupDatabaseWrapper(sqlDB)
	router := service.SetupHttpRouter()
	router.Run(":8080")
}
