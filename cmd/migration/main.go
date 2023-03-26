package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// CreateDatabase initializes and returns a database connection. The connection parameters are
// taken from the system's environment variables.
func CreateDatabase() *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(mysql)/test?parseTime=true", os.Getenv("DBUSER"), os.Getenv("DBPWD"))
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	return sqlDB
}

// Usage example on the command line:
// > DBUSER=dirk DBPWD=bullo92 go run main.go
func main() {
	sqlDB := CreateDatabase()
	db := sqlx.NewDb(sqlDB, "mysql")
	defer db.Close()
	db.MustExec(`
		CREATE TABLE contacts (
			id          INT AUTO_INCREMENT PRIMARY KEY,
			firstname   VARCHAR(50),
			lastname    VARCHAR(50),
			phone       VARCHAR(50),
			birthday    DATE
		)
	`)
	db.MustExec(`
		CREATE INDEX contacts_firstname
		ON contacts (firstname)
	`)
	db.MustExec(`
		CREATE INDEX contacts_lastname
		ON contacts (lastname)
	`)
}
