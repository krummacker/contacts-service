package main

import (
	"bufio"
	"flag"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"gitlab.com/dirk.krummacker/contacts-service/internal/service"
)

// Usage example on the command line:
// > DBHOST=localhost DBUSER=dirk DBPWD=bullo92 go run main.go -file=../../scripts/database.sql
func main() {
	sqlDB := service.CreateDatabase()
	db := sqlx.NewDb(sqlDB, "mysql")
	defer db.Close()

	filePtr := flag.String("file", "database.sql", "the sql file to execute")
	flag.Parse()

	readFile, err := os.Open(*filePtr) // nosemgrep
	if err != nil {
		panic(err)
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	builder := strings.Builder{}
	for fileScanner.Scan() {
		line := fileScanner.Text()
		builder.WriteString(line)
		builder.WriteString(" ")
		if strings.Contains(line, ";") {
			sql := builder.String()
			db.MustExec(sql)
			builder = strings.Builder{}
		}
	}
}
