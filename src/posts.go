package forum

import (
	"context"
	"database/sql"

	//"fmt"
	"html/template"
	"log"
	"net/http"

	"sort"
	"time"

	"github.com/gorilla/sessions"
	//"golang.org/x/text/date"
)

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
	Tags     string
	Author   string
	Likes    int
	Dislikes int
	Date     string
}

type Comment struct {
	Content string
	Author  string
	Idpost  int
}

type FinalData struct {
	UserInfo UserInfo
	Posts    []Post
	Comments []Comment
}


func AddNewPost(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	//println("addNewPost")
	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/index.html")
	} else if r.Method == "POST" {
		title := r.FormValue("title")
		content := r.FormValue("content")
		tags := r.FormValue("tags")
		session, _ := store.Get(r, "session")
		author := session.Values["username"].(string)
		println((author))
		err := AddPostInDb(title, content, tags, author)
		if err != nil {
			println(err.Error())
		} else {
			println("Post added successfully")
			posts := DisplayPost(w)
			tmpl, err := template.ParseFiles("templates/forum.html")
			if err != nil {
				http.Error(w, "Erreur de lecture du fichier HTML 5", http.StatusInternalServerError)
				return
			}
			username, ok := session.Values["username"]

			var data UserInfo
			data.IsLoggedIn = ok
			if !ok {
				tmpl, err := template.ParseFiles("templates/index.html")
				log.Println(err)
				if err != nil {
					http.Error(w, "Erreur de lecture du fichier HTML 1", http.StatusInternalServerError)
					return
				}
				newdata := FinalData{data, DisplayPost(w),DisplayCommments(w)}
				tmpl.Execute(w, newdata)
				return
			} else if ok {
				var profilePicture string
				err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM utilisateurs WHERE username = ?", username).Scan(&profilePicture)
				if err != nil && err != sql.ErrNoRows {
					http.Error(w, "Erreur lors de la récupération de la photo de profil", http.StatusInternalServerError)
					return
				}

				data.Username = username.(string)
				data.ProfilePicture = profilePicture
			}

			newData := FinalData{data, posts,DisplayCommments(w)}
			tmpl.Execute(w, newData)
		}
	}
}

func AddPostInDb(title string, content string, tags string, author string) error {
	db = OpenDb()
	date := time.Now()
	_, err := db.ExecContext(context.Background(), `INSERT INTO posts (title,content,tags,author,date) VALUES (?, ?, ?, ?, ?)`,
		title, content, tags, author, date)
	if err != nil {
		return err
	}
	return nil
}

func DisplayPost(w http.ResponseWriter) []Post {
	db = OpenDb()
	rows, err := db.QueryContext(context.Background(), "SELECT id,title, content, tags, author, likes, dislikes, date FROM posts")
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
		return nil
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var inter Post
		err := rows.Scan(&inter.Id, &inter.Title, &inter.Content, &inter.Tags, &inter.Author, &inter.Likes, &inter.Dislikes, &inter.Date)
		if err != nil {
			http.Error(w, "Erreur lors de la lecture des posts", http.StatusInternalServerError)
			return nil
		}
		inter.Date = inter.Date[:16]
		posts = append(posts, inter)
	}
	if posts == nil {
		date := time.Now()
		date_string := date.Format("01-02-2024 15:04")
		posts = append(posts, Post{Id: -1, Title: "Aucun post", Content: "Aucun post", Tags: "Aucun post", Author: "Aucun post", Likes: 0, Dislikes: 0, Date: date_string})
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date > posts[j].Date
	})

	return posts
}
