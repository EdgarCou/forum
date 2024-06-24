package main

import (
	"database/sql"
	"forum/src"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	//"time"
)

var db *sql.DB

var store = sessions.NewCookieStore([]byte("something-very-secret"))

type UserInfo struct {
	IsLoggedIn     bool
	Email          string
	Username       string
	ProfilePicture string
	Firstname      string
	Lastname       string
	Birthdate      string
}

type Post struct {
	Id       int
	Title    string
	Content  string
	Topics   string
	Author   string
	Likes    int
	Dislikes int
	Date     string
}

type FinalData struct {
	UserInfo UserInfo
	Posts    []Post
}

func main() {

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600, // La session expire lorsque le navigateur est fermé, ou au bout de une heure.
		HttpOnly: true,
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
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
	http.HandleFunc("/CreateComment", forum.CommentHandler)
	http.HandleFunc("/CreateCommentForMyPost", forum.CommentHandlerForMyPost)
	http.HandleFunc("/sort", forum.SortHandler)
	http.HandleFunc("/sortMyPost", forum.SortHandlerMyPost)
	http.HandleFunc("/RGPD", forum.RGPDHandler)
	http.HandleFunc("/addTopic", forum.AddTopicHandler)
	http.HandleFunc("/allTopics", forum.AllTopicsHandler)
	http.HandleFunc("/myPosts", forum.MyPostHandler)
	http.HandleFunc("/particular", forum.ParticularHandler)
	http.HandleFunc("/liked", forum.LikedHandler)
	http.HandleFunc("/delete", forum.DeleteHandler)
	http.HandleFunc("/editPost", forum.EditPostHandler)
	http.HandleFunc("/createCommentParticularTopic", forum.CommentHandlerParticularTopic)
	log.Println("Server is listening on port 8080")
	http.ListenAndServe(":8080", nil)
}


//clé du docker := fb21b8d0c4cbc93c61a398a2b2af3fc2d7e08f9ade215fd94beb41a1e6418ffc