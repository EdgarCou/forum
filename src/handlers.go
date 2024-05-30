package forum


import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
		newdata := FinalData{data, DisplayPost(w)}
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
	post := DisplayPost(w)
	totalData := FinalData{data, post}
	tmpl.Execute(w, totalData)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		println(username, email, password)
		err := AjouterUtilisateur(username, email, password, "")
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
	newData := FinalData{data, DisplayPost(w)}
	tmpl.Execute(w, newData)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/login.html")
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		err := VerifierUtilisateur(username, password)
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

func UserHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
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

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}


func ForumHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
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
		newdata := FinalData{data, DisplayPost(w)}
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
	posts := DisplayPost(w)
	tmpl, err := template.ParseFiles("templates/forum.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 6", http.StatusInternalServerError)
		return
	}
	newData := FinalData{data, posts}
	tmpl.Execute(w, newData)
}

func MembersHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
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
		newdata := FinalData{data, DisplayPost(w)}
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
	newData := FinalData{data, DisplayPost(w)}
	tmpl.Execute(w, newData)
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
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
		newdata := FinalData{data, DisplayPost(w)}
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
	newData := FinalData{data, DisplayPost(w)}
	tmpl.Execute(w, newData)
}

