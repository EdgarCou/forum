package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"forum/src"
)

var db *sql.DB


var store = sessions.NewCookieStore([]byte("something-very-secret"))

type UserInfo struct {
	IsLoggedIn     bool
	Email          string
	Username       string
	ProfilePicture string
}

type Post struct {
	Id       int
	Title    string
	Content  string
	Tags     string
	Author   string
	Likes    int
	Dislikes int
}

type FinalData struct {
	UserInfo UserInfo
	Posts    []Post
}

func main() {

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600, // La session expire lorsque le navigateur est fermé
		HttpOnly: true,
	}

	dbPath := "utilisateurs.db"
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS utilisateurs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		profile_picture TEXT
		)`)
	if err != nil {
		log.Fatal(err)
	}

	var err2 error
	_, err2 = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		tags TEXT,
		author TEXT NOT NULL,
		likes INTEGER DEFAULT 0,
		dislikes INTEGER DEFAULT 0,
		FOREIGN KEY (author) REFERENCES utilisateurs(username)
		ON DELETE CASCADE
    	ON UPDATE CASCADE
		)`)
	if err2 != nil {
		log.Fatal(err2)
	}

	var err3 error
	_, err3 = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS likedBy (
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

	if err3 != nil {
		log.Fatal(err3)
	}

	http.HandleFunc("/", forum.HomeHandler)
	http.HandleFunc("/forum", forum.ForumHandler)
	http.HandleFunc("/signup", forum.RegisterHandler)
	http.HandleFunc("/members", forum.MembersHandler)
	http.HandleFunc("/login", forum.LoginHandler)
	http.HandleFunc("/user", forum.UserHandler)
	http.HandleFunc("/logout", forum.LogoutHandler)
	http.HandleFunc("/profile", forum.UserHandler)
	http.HandleFunc("/createPost", forum.AddNewPost)
	http.HandleFunc("/about", forum.AboutHandler)
	http.HandleFunc("/ws", forum.WsHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server is listening on port 8080")
	http.ListenAndServe(":8080", nil)
}

