package main

import (
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


