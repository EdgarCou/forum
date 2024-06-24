package forum

import (
	"context"
	"database/sql"
	//"fmt"
	"html/template"
	"net/http"
	"sort"
	"time"

	//"github.com/google/pprof/profile"
	"github.com/gorilla/sessions"
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
	Topics     string
	Author   string
	Likes    int
	Dislikes int
	Date     string
	Comments int
	ProfilePicture string
}

type Comment struct {
	Content string
	Author  string
	Idpost  int
}

type Topics struct {
	Title string
	NbPost int
}

type FinalData struct {
	UserInfo UserInfo
	Posts    []Post
	Comments []Comment
	Topics   []Topics
}

type ParticularFinalData struct {
	UserInfo UserInfo
	Posts    []Post
	Comments []Comment
	Topics   Topics
}


func AddNewPost(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()

	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/index.html")
	} else if r.Method == "POST" {
		title := r.FormValue("title")
		content := r.FormValue("content")
		topics := r.FormValue("topics")
		session, _ := store.Get(r, "session")
		author := session.Values["username"].(string)


		rows, errQuery19 := db.QueryContext(context.Background(), "SELECT profile_picture FROM utilisateurs WHERE username = ?", author)

		if errQuery19 != nil {
			http.Error(w, "Error while retrieving the profile picture", http.StatusInternalServerError)
			return
		}

		if rows != nil {
			defer rows.Close()
		}

		var profilePicture string

		for rows.Next() {
			errQuery20 := rows.Scan(&profilePicture)
			if errQuery20 != nil {
				http.Error(w, "Error while reading the profile picture", http.StatusInternalServerError)
				return
			}
		}



		errPost := AddPostInDb(title, content, topics, author,profilePicture)
		if errPost != nil {
			http.Error(w, "Error while adding the post", http.StatusInternalServerError)
		} else {
			posts := DisplayPost(w)
			tmpl, errReading8 := template.ParseFiles("templates/forum.html")
			if errReading8 != nil {
				http.Error(w, "Error reading the HTML file : forum.html", http.StatusInternalServerError)
				return
			}
			username, ok := session.Values["username"]

			var data UserInfo
			data.IsLoggedIn = ok
			if !ok {
				tmpl, errReading9 := template.ParseFiles("templates/index.html")
				if errReading9 != nil {
					http.Error(w, "Error reading the HTML file : index.html", http.StatusInternalServerError)
					return
				}
				newdata := FinalData{data, DisplayPost(w),DisplayCommments(w), DisplayTopics(w)}
				tmpl.Execute(w, newdata)
				return
			} else if ok {
				var profilePicture string
				errQuery10 := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM utilisateurs WHERE username = ?", username).Scan(&profilePicture)
				if errQuery10 != nil && errQuery10 != sql.ErrNoRows {
					http.Error(w, "Error while retrieving the profile picture", http.StatusInternalServerError)
					return
				}
				data.Username = username.(string)
				data.ProfilePicture = profilePicture
			}

			newData := FinalData{data, posts,DisplayCommments(w), DisplayTopics(w)}
			tmpl.Execute(w, newData)
		}
	}
}

func AddPostInDb(title string, content string, topics string, author string, profilePicture string) error {
	db = OpenDb()
	date := time.Now()
	_, errQuery11 := db.ExecContext(context.Background(), `INSERT INTO posts (title,content,topics,author,date,profile_picture) VALUES (?, ?, ?, ?, ?, ?)`,
		title, content, topics, author, date, profilePicture)
	if errQuery11 != nil {
		return errQuery11
	}

	_,errQuery12 := db.ExecContext(context.Background(), `UPDATE topics SET nbpost = nbpost + 1 WHERE title = ?`, topics)
	if errQuery12 != nil {
		return errQuery12
	}
	return nil
}

func DisplayPost(w http.ResponseWriter) []Post {
	db = OpenDb()
	rows, errQuery13 := db.QueryContext(context.Background(), "SELECT id,title, content, topics, author, likes, dislikes, date, comments, profile_picture FROM posts")
	if errQuery13 != nil {
		http.Error(w, "Error while retrieving the posts", http.StatusInternalServerError)
		return nil
	}
	if rows != nil {
		defer rows.Close()
	}

	var posts []Post
	for rows.Next() {
		var inter Post
		errScan6 := rows.Scan(&inter.Id, &inter.Title, &inter.Content, &inter.Topics, &inter.Author, &inter.Likes, &inter.Dislikes, &inter.Date, &inter.Comments, &inter.ProfilePicture)
		if errScan6 != nil {
			http.Error(w, "Error while reading the", http.StatusInternalServerError)
			return nil
		}
		inter.Date = inter.Date[:16]
		posts = append(posts, inter)
	}
	if posts == nil {
		date := time.Now()
		date_string := date.Format("01-02-2024 15:04")
		posts = append(posts, Post{Id: -1, Title: "No title", Content: "No content", Topics: "No topic", Author: "No author", Likes: 0, Dislikes: 0, Date: date_string, Comments: 0, ProfilePicture: "No profile picture"})
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date > posts[j].Date
	})

	return posts
}


func MyPostHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	tmpl, errReading10 := template.ParseFiles("templates/myPost.html")
	if errReading10 != nil {
		http.Error(w, "Error reading the HTML file : myPost.html", http.StatusInternalServerError)
		return
	}
	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w), DisplayCommments(w), DisplayTopics(w)}
	tmpl.Execute(w, newData)
}


func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	id := r.URL.Query().Get("postid")
	topics := r.URL.Query().Get("topics")

	_, errQuery14 := db.ExecContext(context.Background(), "DELETE FROM posts WHERE id = ?", id)
	if errQuery14 != nil {
		http.Error(w, "Error while deleting the post", http.StatusInternalServerError)
		return
	}

	_,errQuery15 := db.ExecContext(context.Background(), `UPDATE topics SET nbpost = nbpost - 1 WHERE title = ?`, topics)
	if errQuery15 != nil {
		http.Error(w, "Error while updating the number of posts", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/myPosts", http.StatusSeeOther)
}

func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	id := r.URL.Query().Get("postid")

	rows, errQuery16 := db.QueryContext(context.Background(), "SELECT topics FROM posts WHERE id = ?", id)
	if errQuery16 != nil {
		http.Error(w, "Error while retrieving the post", http.StatusInternalServerError)
		return
	}

	var currentTopic string

	for rows.Next() {
		errQuery17 := rows.Scan(&currentTopic)
		if errQuery17 != nil {
			http.Error(w, "Error while reading the post", http.StatusInternalServerError)
			return
		}
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	topics := r.FormValue("topics")

	if (topics != currentTopic) {
		UpdateTopics(w, currentTopic, topics)
	}

	_, errQuery18 := db.ExecContext(context.Background(), "UPDATE posts SET title = ?, content = ?, topics = ? WHERE id = ?", title, content, topics, id)
	if errQuery18 != nil {
		http.Error(w, "Error while updating the post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/myPosts", http.StatusSeeOther)
}

