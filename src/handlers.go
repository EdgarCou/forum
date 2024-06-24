package forum

import (
	"context"
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

var db *sql.DB

func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, errWs := upgrader.Upgrade(w, r, nil)
	if errWs != nil {
		log.Println(errWs)
		return
	}

	defer conn.Close()
	LikeHandlerWs(conn, r)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()

	_, errBdd := db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS utilisateurs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE CHECK(length(username) >= 3 AND length(username) <= 20),
		email TEXT NOT NULL UNIQUE CHECK(length(email) >= 3 AND length(email) <= 30),
		password TEXT NOT NULL CHECK(length(password) >= 8),
		profile_picture TEXT,
		firstname TEXT,
		lastname TEXT,
		birthdate TEXT
		)`)
	if errBdd != nil {
		log.Fatal(errBdd)
	}

	_, errBdd2 := db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		topics TEXT,
		author TEXT NOT NULL,
		likes INTEGER DEFAULT 0,
		dislikes INTEGER DEFAULT 0,
		date TEXT,
		comments INTEGER DEFAULT 0,
		profile_picture TEXT,
		FOREIGN KEY (author) REFERENCES utilisateurs(username)
		ON DELETE CASCADE
    	ON UPDATE CASCADE
		FOREIGN KEY (topics) REFERENCES topics(title)
		ON DELETE CASCADE
		ON UPDATE CASCADE	
		FOREIGN KEY (profile_picture) REFERENCES utilisateurs(profile_picture)
		ON DELETE CASCADE
		ON UPDATE CASCADE	
		)`)
	if errBdd2 != nil {
		log.Fatal(errBdd2)
	}

	_, errBdd3 := db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS likedBy (
		username TEXT,
		idpost INTEGER,
		type BOOLEAN,
		PRIMARY KEY (username, idpost, type),
		FOREIGN KEY (username) REFERENCES utilisateurs(username)
		ON DELETE CASCADE
		ON UPDATE CASCADE,
		FOREIGN KEY (idpost) REFERENCES posts(id)
		ON DELETE CASCADE
		ON UPDATE CASCADE
		)`)

	if errBdd3 != nil {
		log.Fatal(errBdd3)
	}

	_, errBdd4 := db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		nbpost INTEGER DEFAULT 0
	)`)

	InitTopics()

	if errBdd4 != nil {
		log.Fatal(errBdd4)
	}

	_, errBdd5 := db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS comments (
		content TEXT NOT NULL,
		author TEXT NOT NULL,
		idpost INTEGER,
		FOREIGN KEY (author) REFERENCES utilisateurs(username)
		ON DELETE CASCADE
		ON UPDATE CASCADE,
		FOREIGN KEY (idpost) REFERENCES posts(id)
		ON DELETE CASCADE
		ON UPDATE CASCADE
	)`)
	if errBdd5 != nil {
		log.Fatal(errBdd5)
	}

	data := CheckUserInfo(w, r)

	tmpl, errReading5 := template.ParseFiles("templates/index.html")
	if errReading5 != nil {
		http.Error(w, "Error reading the HTML file : index.html", http.StatusInternalServerError)
		return
	}

	totalData := FinalData{data, DisplayPost(w), DisplayCommments(w), DisplayTopics(w)}
	tmpl.Execute(w, totalData)
}

