package forum

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	//"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func AddUser(username, email, password, profilePicture, lastname, firstname, birthdate string) error {
	db = OpenDb()
	hashedPassword, errCrypting := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if errCrypting != nil {
		return errCrypting
	}

	_, errQuery26 := db.ExecContext(context.Background(), `INSERT INTO utilisateurs (username, email, password, profile_picture, lastname, firstname, birthdate) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		username, email, hashedPassword, profilePicture, lastname, firstname, birthdate)
	if errQuery26 != nil {
		return errQuery26
	}
	return nil
}

func VerifierUtilisateur(username, password string) error {
	db = OpenDb()
	var passwordDB string
	errQuery27 := db.QueryRowContext(context.Background(), "SELECT password FROM utilisateurs WHERE username = ?", username).Scan(&passwordDB)
	if errQuery27 != nil {
		return errQuery27
	}

	errCrypting2 := bcrypt.CompareHashAndPassword([]byte(passwordDB), []byte(password))
	if errCrypting2 != nil {
		return fmt.Errorf("incorrect password")
	}
	return nil
}


func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		println(username, email, password)
		errUser := AddUser(username, email, password, "", "", "", "")
		if errUser != nil {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><script>alert("Email already use, please find another one."); window.location="/signup";</script></body></html>`)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl, errReading16 := template.ParseFiles("templates/signup.html")
	if errReading16 != nil {
		http.Error(w, "Error reading the HTML file : signup.html", http.StatusInternalServerError)
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

		errUser2 := VerifierUtilisateur(username, password)
		if errUser2 != nil {
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
		http.Error(w, "User not specified", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		var email, profilePicture, firstname, lastname, birthdate string
		query := `SELECT email, profile_picture, firstname, lastname, birthdate FROM utilisateurs WHERE username = ?`
		errQuery28 := db.QueryRowContext(context.Background(), query, username).Scan(&email, &profilePicture, &firstname, &lastname, &birthdate)
		if errQuery28 != nil {
			http.Error(w, "User not found", http.StatusNotFound)
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

		tmpl, errReading17 := template.ParseFiles("templates/user.html")
		if errReading17 != nil {
			http.Error(w, "Error reading the HTML file : user.html", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, newData)

	} else if r.Method == "POST" {

		firstname := r.FormValue("Firstname")
		lastname := r.FormValue("Lastname")
		birthdate := r.FormValue("birthdate")

		file, handler, errUpload := r.FormFile("profilepicture")
        if errUpload != nil {
            http.Error(w, "Error during file upload", http.StatusInternalServerError)
            return
        }
        defer file.Close()

        os.MkdirAll("static/uploads", os.ModePerm)

        filePath := filepath.Join("static/uploads", handler.Filename)
        f, errSave := os.Create(filePath)
        if errSave != nil {
            http.Error(w, "Error saving the file", http.StatusInternalServerError)
            return
        }
        defer f.Close()
        io.Copy(f, file)

        updateSQL := `UPDATE utilisateurs SET profile_picture = ?  WHERE username = ?`
        _, errQuery29 := db.ExecContext(context.Background(), updateSQL, "./static/uploads/"+handler.Filename, username)
        if errQuery29 != nil {
            http.Error(w, "Error updating the profile picture", http.StatusInternalServerError)
            return
        }


		session, _ := store.Get(r, "session")
		session.Values["username"] = username

		_,errQuery30 := db.ExecContext(context.Background(), `UPDATE posts SET profile_picture = ? WHERE author = ?`, "/static/uploads/"+handler.Filename, username)
		if errQuery30 != nil {
			http.Error(w, "Error updating the profile picture", http.StatusInternalServerError)
			return
		}		

        http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther)
	
		updateSQL = `UPDATE utilisateurs SET firstname = ?, lastname = ?, birthdate = ?  WHERE username = ?`
		result, errQuery30 := db.ExecContext(context.Background(), updateSQL, firstname, lastname, birthdate, username)
		if errQuery30 != nil {
			http.Error(w, "Error updating the profile picture", http.StatusInternalServerError)
			return
		}

		_, errScan10 := result.RowsAffected()
		if errScan10 != nil {
			fmt.Println("Error while retrieving the number of affected rows:", errScan10)
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

func CheckUserInfo(w http.ResponseWriter, r *http.Request) UserInfo {
    session, _ := store.Get(r, "session")
    username, ok := session.Values["username"]

    var data UserInfo
    data.IsLoggedIn = ok
    if ok {
        var profilePicture string
        errQuery31 := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM utilisateurs WHERE username = ?", username).Scan(&profilePicture)
        if errQuery31 != nil && errQuery31 != sql.ErrNoRows {
            http.Error(w, "Error while retrieving the profile picture", http.StatusInternalServerError)
            return data
        }

        data.Username = username.(string)
        data.ProfilePicture = profilePicture
    }

    return data
}

func RGPDHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, errReading18 := template.ParseFiles("templates/RGPD.html")
	if errReading18 != nil {
		http.Error(w, "Error reading the HTML file : RGPD.html", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}