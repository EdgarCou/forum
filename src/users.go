package forum

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"golang.org/x/crypto/bcrypt"
)

func AjouterUtilisateur(username, email, password, profilePicture, lastname, firstname, birthdate string) error {
	db = OpenDb()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `INSERT INTO utilisateurs (username, email, password, profile_picture, lastname, firstname, birthdate) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		username, email, hashedPassword, profilePicture, lastname, firstname, birthdate)
	if err != nil {
		return err
	}
	return nil
}

func VerifierUtilisateur(username, password string) error {
	db = OpenDb()
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
		var err error
		updateSQL := `UPDATE utilisateurs SET firstname = ?, lastname = ?, birthdate = ?  WHERE username = ?`
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

func RGPDHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/RGPD.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 9", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}