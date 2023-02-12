package main

import (
	_ "github.com/go-sql-driver/mysql"
)

// Usage example on the command line:
// > DBUSER=dirk DBPWD=bullo92 GIN_LOGGING=OFF go run contacts-service.go
func main() {
	sqlDB := createDatabase()
	setupDatabaseWrapper(sqlDB)
	router := setupHttpRouter()
	router.Run(":8080")
}
