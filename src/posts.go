package forum

import (
	"context"
	"database/sql"
	"html/template"
	"net/http"
	"sort"
	"time"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("something-very-secret")) // Create a new cookie store

type UserInfo struct { // UserInfo structure
	IsLoggedIn     bool
	Email          string
	Username       string
	ProfilePicture string
	Firstname      string
	Lastname       string
	Birthdate      string
}

type Post struct { // Post structure
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

type Comment struct { // Comment structure
	Content string
	Author  string
	Idpost  int
}

type Topics struct { // Topics structure
	Title string
	NbPost int
}

type FinalData struct { // FinalData structure
	UserInfo UserInfo
	Posts    []Post
	Comments []Comment
	Topics   []Topics
}

type ParticularFinalData struct { // ParticularFinalData structure
	UserInfo UserInfo
	Posts    []Post
	Comments []Comment
	Topics   Topics
}


func AddNewPost(w http.ResponseWriter, r *http.Request) { // Function for adding a new post
	db = OpenDb() // Open the database

	if r.Method == "GET" { // If the method is GET
		http.ServeFile(w, r, "templates/index.html")
	} else if r.Method == "POST" { // If the method is POST
		title := r.FormValue("title") // Get the title
		content := r.FormValue("content") // Get the content
		topics := r.FormValue("topics") // Get the topics
		session, _ := store.Get(r, "session") // Get the session
		author := session.Values["username"].(string) // Get the author


		rows, errQuery19 := db.QueryContext(context.Background(), "SELECT profile_picture FROM utilisateurs WHERE username = ?", author) // Get the profile picture

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

		errPost := AddPostInDb(title, content, topics, author,profilePicture) // Add the post in the database
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
			if !ok { // If the user is not logged in
				tmpl, errReading9 := template.ParseFiles("templates/index.html")
				if errReading9 != nil {
					http.Error(w, "Error reading the HTML file : index.html", http.StatusInternalServerError)
					return
				}
				newdata := FinalData{data, DisplayPost(w),DisplayCommments(w), DisplayTopics(w)} // Get the data to be displayed
				tmpl.Execute(w, newdata)
				return
			} else if ok { // If the user is logged in
				var profilePicture string
				errQuery10 := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM utilisateurs WHERE username = ?", username).Scan(&profilePicture) // Get the profile picture
				if errQuery10 != nil && errQuery10 != sql.ErrNoRows {
					http.Error(w, "Error while retrieving the profile picture", http.StatusInternalServerError)
					return
				}
				data.Username = username.(string)
				data.ProfilePicture = profilePicture
			}

			newData := FinalData{data, posts,DisplayCommments(w), DisplayTopics(w)} // Get the data to be displayed
			tmpl.Execute(w, newData) // Execute the template
		}
	}
}

func AddPostInDb(title string, content string, topics string, author string, profilePicture string) error { // Function for adding a post in the database
	db = OpenDb()
	date := time.Now()
	_, errQuery11 := db.ExecContext(context.Background(), `INSERT INTO posts (title,content,topics,author,date,profile_picture) VALUES (?, ?, ?, ?, ?, ?)`, // Insert the post in the database
		title, content, topics, author, date, profilePicture)
	if errQuery11 != nil {
		return errQuery11
	}

	_,errQuery12 := db.ExecContext(context.Background(), `UPDATE topics SET nbpost = nbpost + 1 WHERE title = ?`, topics) // Update the number of posts of the topic
	if errQuery12 != nil {
		return errQuery12
	}
	return nil
}

func DisplayPost(w http.ResponseWriter) []Post { // Function for displaying the posts
	db = OpenDb()
	rows, errQuery13 := db.QueryContext(context.Background(), "SELECT id,title, content, topics, author, likes, dislikes, date, comments, profile_picture FROM posts") // Get the posts from the database
	if errQuery13 != nil {
		http.Error(w, "Error while retrieving the posts", http.StatusInternalServerError)
		return nil
	}
	if rows != nil {
		defer rows.Close()
	}

	var posts []Post
	for rows.Next() { // Get the posts
		var inter Post
		errScan6 := rows.Scan(&inter.Id, &inter.Title, &inter.Content, &inter.Topics, &inter.Author, &inter.Likes, &inter.Dislikes, &inter.Date, &inter.Comments, &inter.ProfilePicture)
		if errScan6 != nil {
			http.Error(w, "Error while reading the", http.StatusInternalServerError)
			return nil
		}
		inter.Date = inter.Date[:16] // Get the good format for the date
		posts = append(posts, inter)
	}
	if posts == nil { // If there is no post
		date := time.Now()
		date_string := date.Format("01-02-2024 15:04")
		posts = append(posts, Post{Id: -1, Title: "No title", Content: "No content", Topics: "No topic", Author: "No author", Likes: 0, Dislikes: 0, Date: date_string, Comments: 0, ProfilePicture: "No profile picture"}) // Create an empty post with the id -1
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date > posts[j].Date // Sort the posts by the date
	})

	return posts
}


func MyPostHandler(w http.ResponseWriter, r *http.Request) { // Function for handling the posts of the user
	db = OpenDb()
	tmpl, errReading10 := template.ParseFiles("templates/myPost.html")
	if errReading10 != nil {
		http.Error(w, "Error reading the HTML file : myPost.html", http.StatusInternalServerError)
		return
	}
	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w), DisplayCommments(w), DisplayTopics(w)} // Get the data to be displayed
	tmpl.Execute(w, newData)
}


func DeleteHandler(w http.ResponseWriter, r *http.Request) { // Function for deleting a post
	db = OpenDb() // Open the database
	id := r.URL.Query().Get("postid") // Get the post id
	topics := r.URL.Query().Get("topics") // Get the topics

	_, errQuery14 := db.ExecContext(context.Background(), "DELETE FROM posts WHERE id = ?", id) // Delete the post from the database
	if errQuery14 != nil {
		http.Error(w, "Error while deleting the post", http.StatusInternalServerError)
		return
	}

	_,errQuery15 := db.ExecContext(context.Background(), `UPDATE topics SET nbpost = nbpost - 1 WHERE title = ?`, topics) // Update the number of posts of the topic
	if errQuery15 != nil {
		http.Error(w, "Error while updating the number of posts", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/myPosts", http.StatusSeeOther) // Redirect to the user posts
}

func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb() // Open the database
	id := r.URL.Query().Get("postid") // Get the post id

	rows, errQuery16 := db.QueryContext(context.Background(), "SELECT topics FROM posts WHERE id = ?", id) // Get the topics of the post
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

	title := r.FormValue("title") // Get the title
	content := r.FormValue("content") // Get the content
	topics := r.FormValue("topics") // Get the topics

	if (topics != currentTopic) { // If the topic is different
		UpdateTopics(w, currentTopic, topics) // Update the topic
	}

	_, errQuery18 := db.ExecContext(context.Background(), "UPDATE posts SET title = ?, content = ?, topics = ? WHERE id = ?", title, content, topics, id) // Update the post
	if errQuery18 != nil {
		http.Error(w, "Error while updating the post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/myPosts", http.StatusSeeOther) // Redirect to the user posts page
}

