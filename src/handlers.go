package forum

import (
	"context"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"sort"
)

var db *sql.DB

func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	LikeHandlerWs(conn, r)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()

	_, err := db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS utilisateurs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE CHECK(length(username) >= 3 AND length(username) <= 20),
		email TEXT NOT NULL UNIQUE CHECK(length(email) >= 3 AND length(email) <= 30),
		password TEXT NOT NULL CHECK(length(password) >= 8),
		profile_picture TEXT,
		firstname TEXT,
		lastname TEXT,
		birthdate TEXT
		)`)
	if err != nil {
		log.Fatal(err)
	}

	var err2 error
	_, err2 = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		topics TEXT,
		author TEXT NOT NULL,
		likes INTEGER DEFAULT 0,
		dislikes INTEGER DEFAULT 0,
		date TEXT,
		comments INTEGER DEFAULT 0,
		FOREIGN KEY (author) REFERENCES utilisateurs(username)
		ON DELETE CASCADE
    	ON UPDATE CASCADE
		FOREIGN KEY (topics) REFERENCES topics(title)
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

	var err4 error
	_, err4 = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		nbpost INTEGER DEFAULT 0
	)`)

	InitTopics()


	
	if err4 != nil {
		log.Fatal(err4)
	}

	var err5 error
	_, err5 = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS comments (
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
	if err5 != nil {
		log.Fatal(err5)
	}

	data := CheckUserInfo(w, r)

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 2 ", http.StatusInternalServerError)
		return
	}
	totalData := FinalData{data, DisplayPost(w), DisplayCommments(w), DisplayTopics(w)}
	tmpl.Execute(w, totalData)
}

func ForumHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	tmpl, err := template.ParseFiles("templates/forum.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 6", http.StatusInternalServerError)
		return
	}
	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w), DisplayCommments(w), DisplayTopics(w)}
	tmpl.Execute(w, newData)
}

func MembersHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	tmpl, err := template.ParseFiles("templates/members.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 7", http.StatusInternalServerError)
		return
	}
	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w), DisplayCommments(w), DisplayTopics(w)}
	tmpl.Execute(w, newData)
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	tmpl, err := template.ParseFiles("templates/about.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 8", http.StatusInternalServerError)
		return
	}
	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w), DisplayCommments(w), DisplayTopics(w)}
	tmpl.Execute(w, newData)
}

func SortHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	tmpl, err := template.ParseFiles("templates/forum.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 10", http.StatusInternalServerError)
		return
	}

	sortType := r.FormValue("sort")

	var rows *sql.Rows
	var posts []Post

	if sortType == "mostLiked" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Likes > posts[j].Likes
		})
	} else if sortType == "mostDisliked" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Dislikes > posts[j].Dislikes
		})
	} else if sortType == "newest" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Date > posts[j].Date
		})
	} else if sortType == "oldest" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Date < posts[j].Date
		})
	}else if sortType == "A-Z" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Title < posts[j].Title
		})
	}else if sortType == "Z-A" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Title > posts[j].Title
		}) 
	} else {
		http.Error(w, "Tri invalide", http.StatusBadRequest)
		return
	}

	if rows != nil {
		defer rows.Close()
	}

	newData := FinalData{CheckUserInfo(w, r), posts, DisplayCommments(w), DisplayTopics(w)}

	tmpl.Execute(w, newData)
}


func SortHandlerMyPost(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	tmpl, err := template.ParseFiles("templates/myPost.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 10", http.StatusInternalServerError)
		return
	}

	sortType := r.FormValue("sort")

	var rows *sql.Rows
	var posts []Post

	if sortType == "mostLiked" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Likes > posts[j].Likes
		})
	} else if sortType == "mostDisliked" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Dislikes > posts[j].Dislikes
		})
	} else if sortType == "newest" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Date > posts[j].Date
		})
	} else if sortType == "oldest" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Date < posts[j].Date
		})
	}else if sortType == "A-Z" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Title < posts[j].Title
		})
	}else if sortType == "Z-A" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Title > posts[j].Title
		}) 
	} else {
		http.Error(w, "Tri invalide", http.StatusBadRequest)
		return
	}

	if rows != nil {
		defer rows.Close()
	}

	newData := FinalData{CheckUserInfo(w, r), posts, DisplayCommments(w), DisplayTopics(w)}

	tmpl.Execute(w, newData)
}

func RGPDHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/RGPD.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 9", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

