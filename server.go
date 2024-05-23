package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	//"os/user"
	"path/filepath"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	//"golang.org/x/text/language/display"
)


var db *sql.DB
var store = sessions.NewCookieStore([]byte("Edd-Key"))

type UserInfo struct {
	IsLoggedIn     bool
	Email          string
	Username       string
	ProfilePicture string
}

type Post struct {
	Title   string
	Content string
	Tags    string
}

type FinalData struct {
	UserInfo UserInfo
	Posts []Post
}


func main() {
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
		tags TEXT
		)`)
	if err2 != nil {
		log.Fatal(err2)
	}
	

	http.HandleFunc("/",homeHandler)
	http.HandleFunc("/signup", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/user", userHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/profile", userHandler)
	http.HandleFunc("/createPost", addNewPost)	
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    username, ok := session.Values["username"]

	var data UserInfo
	if !ok {
        tmpl, err := template.ParseFiles("templates/index.html")
		log.Println(err)
        if err != nil {
            http.Error(w, "Erreur de lecture du fichier HTML 1", http.StatusInternalServerError)
            return
        }
		newdata := FinalData{data,displayPost(w,r)}
        tmpl.Execute(w, newdata)
        return
    }else if ok {
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
	post := displayPost(w,r)
	totalData := FinalData{data,post}
    tmpl.Execute(w,totalData)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
			username := r.FormValue("username")
			email := r.FormValue("email")
			password := r.FormValue("password")

			println(username, email, password)
		err := ajouterUtilisateur(username, email, password, "")
		if err != nil {
			http.Error(w, "Erreur lors de l'inscription" + err.Error(), http.StatusInternalServerError)
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
	newData := FinalData{data,displayPost(w,r)}
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
			http.Error(w, "Nom d'utilisateur ou mot de passe incorrect", http.StatusUnauthorized)
			return
		}

		session, _ := store.Get(r, "session")
		session.Values["username"] = username
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Utilisateur non spécifié", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		var email, profilePicture string
		query := `SELECT email, profile_picture FROM utilisateurs WHERE username = ?`
		err := db.QueryRowContext(context.Background(), query, username).Scan(&email, &profilePicture)
		if err != nil {
			http.Error(w, "Utilisateur non trouvé", http.StatusNotFound)
			return
		}

		var newData UserInfo
		newData.Username = username
		newData.Email = email	
		newData.ProfilePicture = profilePicture
		newData.IsLoggedIn = username != ""

		tmpl, err := template.ParseFiles("templates/user.html")
		if err != nil {
			http.Error(w, "Erreur de lecture du fichier HTML 4", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, newData)
	} else if r.Method == "POST" {
		file, handler, err := r.FormFile("profile_picture")
		if err != nil {
			http.Error(w, "Erreur lors du téléchargement du fichier", http.StatusInternalServerError)
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

		updateSQL := `UPDATE utilisateurs SET profile_picture = ? WHERE username = ?`
		_, err = db.ExecContext(context.Background(), updateSQL, "/static/uploads/"+handler.Filename, username)
		if err != nil {
			http.Error(w, "Erreur lors de la mise à jour de la photo de profil", http.StatusInternalServerError)
			return
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

func ajouterUtilisateur(username, email, motDePasse, profilePicture string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(motDePasse), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `INSERT INTO utilisateurs (username, email, password, profile_picture) VALUES (?, ?, ?, ?)`,
		username, email, hashedPassword, profilePicture)
	if err != nil {
		return err
	}
	return nil
}

func verifierUtilisateur(username, motDePasse string) error {
	var motDePasseDB string
	err := db.QueryRowContext(context.Background(), "SELECT password FROM utilisateurs WHERE username = ?", username).Scan(&motDePasseDB)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(motDePasseDB), []byte(motDePasse))
	if err != nil {
		return fmt.Errorf("mot de passe incorrect")
	}
	return nil
}

func addNewPost(w http.ResponseWriter, r *http.Request) {
	println("addNewPost")
	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/index.html")
	} else if r.Method == "POST" {
		title := r.FormValue("title")
		content := r.FormValue("content")
		tags := r.FormValue("tags")

		err := ajouterPostinDb(title, content, tags)
		if err != nil {
    		println(err.Error())
		} else {
    		println("Post added successfully")
			posts := displayPost(w,r)
			tmpl, err := template.ParseFiles("templates/index.html")
			if err != nil {
				http.Error(w, "Erreur de lecture du fichier HTML 5", http.StatusInternalServerError)
				return
			}
			data := UserInfo{}
			newData := FinalData{data,posts}
			tmpl.Execute(w, newData)
		}		
	}
}

func ajouterPostinDb(title string,content string, tags string) error {
	_, err := db.ExecContext(context.Background(), `INSERT INTO posts (title,content,tags) VALUES (?, ?, ?)`,
		title, content, tags)
	if err != nil {
		return err
	}
	return nil
}

func displayPost(w http.ResponseWriter, r *http.Request) []Post {
	rows, err := db.QueryContext(context.Background(), "SELECT title, content, tags FROM posts")
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
		return nil
	}
	defer rows.Close()


	var posts []Post
	for rows.Next() {
		var inter Post
		err := rows.Scan(&inter.Title, &inter.Content, &inter.Tags)
		if err != nil {
			http.Error(w, "Erreur lors de la lecture des posts", http.StatusInternalServerError)
			return nil
		}
		posts = append(posts, inter)
	}

	return posts
}
