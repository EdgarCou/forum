package forum		


import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)


func OpenDb() *sql.DB{
	dbPath := "utilisateurs.db"
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	return db
}