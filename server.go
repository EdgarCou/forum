package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	//"io"
	"log"
	"net/http"
	//"os"

	//"os/user"
	//"path/filepath"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"

	//"golang.org/x/text/language/display"
	"github.com/gorilla/websocket"
	//"strconv"
	"strings"
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
	Tags     string
	Author   string
	Likes    int
	Dislikes int
}

type FinalData struct {
	UserInfo UserInfo
	Posts    []Post
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600, // La session expire lorsque le navigateur est fermé, ou au bout de une heure. 
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

	var err4 error
	_, err3 = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL
		)`)
	if err4 != nil {
		log.Fatal(err3)
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/forum", forumHandler)
	http.HandleFunc("/signup", registerHandler)
	http.HandleFunc("/members", membersHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/user", userHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/profile", userHandler)
	http.HandleFunc("/createPost", addNewPost)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/ws", wsHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started at :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	likeHandlerWs(conn, r)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
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
		newdata := FinalData{data, displayPost(w)}
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

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 2 ", http.StatusInternalServerError)
		return
	}
	post := displayPost(w)
	totalData := FinalData{data, post}
	tmpl.Execute(w, totalData)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		println(username, email, password)
		err := ajouterUtilisateur(username, email, password, "", "", "", "")
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><script>alert("Email already use, please find another one."); window.location="/signup";</script></body></html>`)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/signup.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 3", http.StatusInternalServerError)
		return
	}
	data := UserInfo{}
	newData := FinalData{data, displayPost(w)}
	tmpl.Execute(w, newData)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/login.html")
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		err := verifierUtilisateur(username, password)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><script>alert("Username or password incorrect"); window.location="/login";</script></body></html>`)
			return
		}

		session, _ := store.Get(r, "session")
		session.Values["username"] = username
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    username := ""
    if r.Method == "POST" {
        username = r.FormValue("username")
    } else {
        username = r.URL.Query().Get("username")
    }

    if username == "" {
        http.Error(w, "Utilisateur non spécifié", http.StatusBadRequest)
        return
    }

	if r.Method == "GET" {
		var email, profilePicture, firstname, lastname, birthdate string
		query := `SELECT email, profile_picture, firstname, lastname, birthdate FROM utilisateurs WHERE username = ?`
		err := db.QueryRowContext(context.Background(), query, username).Scan(&email, &profilePicture, &firstname, &lastname, &birthdate)
		if err != nil {
			http.Error(w, "Utilisateur non trouvé", http.StatusNotFound)
			return
		}

		var newData UserInfo
		newData.Username = username
		newData.Email = email
		newData.ProfilePicture = profilePicture
		newData.Firstname = firstname
		newData.Lastname = lastname
		newData.Birthdate = birthdate
		newData.IsLoggedIn = username != ""

		tmpl, err := template.ParseFiles("templates/user.html")
		if err != nil {
			http.Error(w, "Erreur de lecture du fichier HTML 4", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, newData)

	} else if r.Method == "POST" {

		firstname := r.FormValue("Firstname")
		lastname := r.FormValue("Lastname")
		birthdate := r.FormValue("birthdate")

		println(firstname, lastname, birthdate)

		/*file, handler, err := r.FormFile("profile_picture")
		if err != nil {
			http.Error(w, "Erreur lors du téléchargement du fichier"+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		os.MkdirAll("static/uploads", os.ModePerm)

		filePath := filepath.Join("static/uploads", handler.Filename)
		f, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Erreur lors de la sauvegarde du fichier", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		io.Copy(f, file)
		*/
		var err error
		updateSQL := `UPDATE utilisateurs SET firstname = ?, lastname = ?, birthdate = ?  WHERE username = ?`
result , err := db.ExecContext(context.Background(), updateSQL, firstname, lastname, birthdate, username, )
		if err != nil {
			http.Error(w, "Erreur lors de la mise à jour de la photo de profil", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			fmt.Println("Erreur lors de la récupération du nombre de lignes affectées :", err)
			return
		}

		if rowsAffected == 0 {
			fmt.Println("Aucune ligne n'a été mise à jour")
		} else {
			fmt.Println("Nombre de lignes mises à jour :", rowsAffected)
		}

		http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ajouterUtilisateur(username, email, password, profilePicture, lastname, firstname, birthdate string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `INSERT INTO utilisateurs (username, email, password, profile_picture, lastname, firstname, birthdate) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		username, email, hashedPassword, profilePicture, lastname, firstname, birthdate)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func verifierUtilisateur(username, password string) error {
	var passwordDB string
	err := db.QueryRowContext(context.Background(), "SELECT password FROM utilisateurs WHERE username = ?", username).Scan(&passwordDB)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordDB), []byte(password))
	if err != nil {
		return fmt.Errorf("mot de passe incorrect")
	}
	return nil
}

func addNewPost(w http.ResponseWriter, r *http.Request) {
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
		err := ajouterPostinDb(title, content, tags, author)
		if err != nil {
			println(err.Error())
		} else {
			println("Post added successfully")
			posts := displayPost(w)
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
				newdata := FinalData{data, displayPost(w)}
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

			newData := FinalData{data, posts}
			tmpl.Execute(w, newData)
		}
	}
}

func ajouterPostinDb(title string, content string, tags string, author string) error {
	_, err := db.ExecContext(context.Background(), `INSERT INTO posts (title,content,tags,author) VALUES (?, ?, ?, ?)`,
		title, content, tags, author)
	if err != nil {
		return err
	}
	return nil
}


func displayPost(w http.ResponseWriter) []Post {
	rows, err := db.QueryContext(context.Background(), "SELECT id,title, content, tags, author, likes, dislikes FROM posts")
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
		return nil
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var inter Post
		err := rows.Scan(&inter.Id, &inter.Title, &inter.Content, &inter.Tags, &inter.Author, &inter.Likes, &inter.Dislikes)
		if err != nil {
			http.Error(w, "Erreur lors de la lecture des posts", http.StatusInternalServerError)
			return nil
		}
		posts = append(posts, inter)
	}
	if posts == nil {
		posts = append(posts, Post{Id: -1, Title: "Aucun post", Content: "Aucun post", Tags: "Aucun post", Author: "Aucun post", Likes: 0, Dislikes: 0})
	}

	return posts
}

func forumHandler(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]

	var data UserInfo
	data.IsLoggedIn = ok
	if !ok {
		tmpl, err := template.ParseFiles("templates/forum.html")
		log.Println(err)
		if err != nil {
			http.Error(w, "Erreur de lecture du fichier HTML 1", http.StatusInternalServerError)
			return
		}
		newdata := FinalData{data, displayPost(w)}
		fmt.Println(newdata.UserInfo.IsLoggedIn)
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
	posts := displayPost(w)
	tmpl, err := template.ParseFiles("templates/forum.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 6", http.StatusInternalServerError)
		return
	}
	newData := FinalData{data, posts}
	tmpl.Execute(w, newData)
}

func membersHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]

	var data UserInfo
	data.IsLoggedIn = ok
	if !ok {
		tmpl, err := template.ParseFiles("templates/members.html")
		log.Println(err)
		if err != nil {
			http.Error(w, "Erreur de lecture du fichier HTML 1", http.StatusInternalServerError)
			return
		}
		newdata := FinalData{data, displayPost(w)}
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
	tmpl, err := template.ParseFiles("templates/members.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 7", http.StatusInternalServerError)
		return
	}
	newData := FinalData{data, displayPost(w)}
	tmpl.Execute(w, newData)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]

	var data UserInfo
	data.IsLoggedIn = ok
	if !ok {
		tmpl, err := template.ParseFiles("templates/about.html")
		log.Println(err)
		if err != nil {
			http.Error(w, "Erreur de lecture du fichier HTML 1", http.StatusInternalServerError)
			return
		}
		newdata := FinalData{data, displayPost(w)}
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
	tmpl, err := template.ParseFiles("templates/about.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 8", http.StatusInternalServerError)
		return
	}
	newData := FinalData{data, displayPost(w)}
	tmpl.Execute(w, newData)
}

func likeHandlerWs(conn *websocket.Conn, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, _ := session.Values["username"]

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v. Message type: %v. Message: %v", err, messageType, string(p))
			return
		}

		if messageType == websocket.TextMessage {
			message := string(p)
			var id string
			var query string
			var likeType int

			if strings.HasPrefix(message, "dislike:") {
				likeType = 0
				id = strings.TrimPrefix(message, "dislike:")
			} else {
				likeType = 1
				id = strings.TrimPrefix(message, "like:")
			}

			checkQuery := "SELECT COUNT(*) FROM likedBy WHERE username = ? AND idpost = ? AND type = ?"
			row := db.QueryRow(checkQuery, username, id, likeType)
			var count int
			err := row.Scan(&count)
			if err != nil {
				log.Printf("Error checking if user has already liked/disliked post: %v", err)
				return
			}
			var countInverse int
			// Si l'utilisateur n'a pas déjà aimé ou n'a pas aimé ce post avec le même type, insérez une nouvelle ligne
			if count == 0 {
				checkInverseQuery := "SELECT COUNT(*) FROM likedBy WHERE username = ? AND idpost = ? AND type = ?"
				row := db.QueryRow(checkInverseQuery, username, id, 1-likeType)
				err9 := row.Scan(&countInverse)
				if err9 != nil {
					log.Printf("Error checking if user has already liked/disliked post: %v", err9)
					return
				}
				if countInverse > 0 {
					var queryRemove string
					var queryDelete string
					if likeType == 0 {
						queryRemove = "UPDATE posts SET likes = likes - 1 WHERE id = ?"
						queryDelete = "DELETE FROM likedBy WHERE username = ? AND idpost = ? AND type = ?"
					} else {
						queryRemove = "UPDATE posts SET dislikes = dislikes - 1 WHERE id = ?"
						queryDelete = "DELETE FROM likedBy WHERE username = ? AND idpost = ? AND type = ?"
					}
					_, errRemove := db.Exec(queryRemove, id)
					if errRemove != nil {
						log.Printf("Error updating likes/dislikes count: %v", errRemove)
					}
					_, errDelete := db.Exec(queryDelete, username, id, 1-likeType)
					if errDelete != nil {
						log.Printf("Error deleting row: %v", errDelete)
					}
				}

				if strings.HasPrefix(message, "dislike:") {
					id = strings.TrimPrefix(message, "dislike:")
					query = "UPDATE posts SET dislikes = dislikes + 1 WHERE id = ?"
					likeType = 0
					_, err7 := db.Exec(query, id)
					if err7 != nil {
						log.Printf("Error updating likes/dislikes count: %v", err)
					}
				} else {
					id = strings.TrimPrefix(message, "like:")
					query = "UPDATE posts SET likes = likes + 1 WHERE id = ?"
					likeType = 1
					_, err8 := db.Exec(query, id)
					if err8 != nil {
						log.Printf("Error updating likes/dislikes count: %v", err)
					}
				}

				likeQuery := "INSERT INTO likedBy (username, idpost, type) VALUES (?, ?, ?)"
				_, err := db.Exec(likeQuery, username, id, likeType)
				if err != nil {
					log.Printf("Error liking/disliking post: %v", err)
				}

			} else {
				if strings.HasPrefix(message, "dislike:") {
					likeType = 0
				} else {
					likeType = 1
				}
				deleteQuery := "DELETE FROM likedBy WHERE username = ? AND idpost = ? AND type = ?"
				_, err := db.Exec(deleteQuery, username, id, likeType)
				if err != nil {
					log.Printf("Error unliking/undisliking post: %v", err)
				}

				// Mettez à jour le nombre de likes ou de dislikes dans la table posts
				var query string
				if likeType == 0 { // Si le type est 0 (dislike), décrémentez le nombre de dislikes
					query = "UPDATE posts SET dislikes = dislikes - 1 WHERE id = ?"
				} else { // Si le type est 1 (like), décrémentez le nombre de likes
					query = "UPDATE posts SET likes = likes - 1 WHERE id = ?"
				}
				_, err = db.Exec(query, id)
				if err != nil {
					log.Printf("Error updating likes/dislikes count: %v", err)
				}
			}

			// Get the new number of likes or dislikes
			var likes, dislikes int
			err = db.QueryRowContext(context.Background(), "SELECT likes, dislikes FROM posts WHERE id = ?", id).Scan(&likes, &dislikes)
			if err != nil {
				log.Println(err)
				return
			}

			var response string
			if countInverse > 0 && likeType == 0 {
				response = fmt.Sprintf("dislikes:%s:%d:likes:%s:%d", id, dislikes, id, likes)
			} else if countInverse > 0 && likeType == 1 {
				response = fmt.Sprintf("likes:%s:%d:dislikes:%s:%d", id, likes, id, dislikes)
			} else {
				if likeType == 0 {
					response = fmt.Sprintf("dislikes:%s:%d", id, dislikes)
				} else {
					response = fmt.Sprintf("likes:%s:%d", id, likes)
				}
			}
			err = conn.WriteMessage(websocket.TextMessage, []byte(response))
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}
