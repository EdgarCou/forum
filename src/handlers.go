package forum

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	//"time"
	"io"
	"log"
	"net/http"

	//"os"
	//"path/filepath"
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
		tags TEXT,
		author TEXT NOT NULL,
		likes INTEGER DEFAULT 0,
		dislikes INTEGER DEFAULT 0,
		date TEXT,
		comments INTEGER DEFAULT 0,
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
	_, err4 = db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL
	)`)

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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		println(username, email, password)
		err := AjouterUtilisateur(username, email, password, "", "", "", "")
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
	newData := FinalData{data, DisplayPost(w), DisplayCommments(w), DisplayTopics(w)}
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

		file, handler, err := r.FormFile("profilepicture")
        if err != nil {
            http.Error(w, "Error during file upload", http.StatusInternalServerError)
            return
        }
        defer file.Close()

        os.MkdirAll("static/uploads", os.ModePerm)

        filePath := filepath.Join("static/uploads", handler.Filename)
        f, err := os.Create(filePath)
        if err != nil {
            http.Error(w, "Error saving the file", http.StatusInternalServerError)
            return
        }
        defer f.Close()
        io.Copy(f, file)

        updateSQL := `UPDATE utilisateurs SET profile_picture = ?  WHERE username = ?`
        _, err = db.ExecContext(context.Background(), updateSQL, "/static/uploads/"+handler.Filename, username)
        if err != nil {
            http.Error(w, "Error updating the profile picture", http.StatusInternalServerError)
            return
        }

        http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
	
		updateSQL = `UPDATE utilisateurs SET firstname = ?, lastname = ?, birthdate = ?  WHERE username = ?`
		result, err := db.ExecContext(context.Background(), updateSQL, firstname, lastname, birthdate, username)
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

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
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

func CheckUserInfo(w http.ResponseWriter, r *http.Request) UserInfo {
    session, _ := store.Get(r, "session")
    username, ok := session.Values["username"]

    var data UserInfo
    data.IsLoggedIn = ok
    if ok {
        var profilePicture string
        err := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM utilisateurs WHERE username = ?", username).Scan(&profilePicture)
        if err != nil && err != sql.ErrNoRows {
            http.Error(w, "Erreur lors de la récupération de la photo de profil", http.StatusInternalServerError)
            return data
        }

        data.Username = username.(string)
        data.ProfilePicture = profilePicture
    }

    return data
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

func RGPDHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/RGPD.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 9", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
