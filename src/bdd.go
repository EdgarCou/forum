package forum

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func OpenDb() *sql.DB { // Function to open the database
	dbPath := "forum.db"
	db, errOpenBDD := sql.Open("sqlite3", dbPath)
	if errOpenBDD != nil {
		log.Fatal(errOpenBDD)
	}
	return db
}
