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
	"golang.org/x/crypto/bcrypt"
)

func AddUser(username, email, password, profilePicture, lastname, firstname, birthdate string) error { // Function to add a user
	db = OpenDb() // Open the database
	hashedPassword, errCrypting := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // Hash the password
	if errCrypting != nil {
		return errCrypting
	}

	_, errQuery26 := db.ExecContext(context.Background(), `INSERT INTO utilisateurs (username, email, password, profile_picture, lastname, firstname, birthdate) VALUES (?, ?, ?, ?, ?, ?, ?)`, // Insert the user in the database
		username, email, hashedPassword, profilePicture, lastname, firstname, birthdate)
	if errQuery26 != nil {
		return errQuery26
	}
	return nil
}

func VerifierUtilisateur(username, password string) error { // Function to verify the user
	db = OpenDb()
	var passwordDB string
	errQuery27 := db.QueryRowContext(context.Background(), "SELECT password FROM utilisateurs WHERE username = ?", username).Scan(&passwordDB) // Get the password from the database
	if errQuery27 != nil {
		return errQuery27
	}

	errCrypting2 := bcrypt.CompareHashAndPassword([]byte(passwordDB), []byte(password)) // Compare the password
	if errCrypting2 != nil {
		return fmt.Errorf("incorrect password")
	}
	return nil
}


func RegisterHandler(w http.ResponseWriter, r *http.Request) { // Function to handle the registration
	if r.Method == http.MethodPost {
		username := r.FormValue("username") // Get the username
		email := r.FormValue("email") // Get the email
		password := r.FormValue("password") // Get the password

		errUser := AddUser(username, email, password, "", "", "", "") // Add the user
		if errUser != nil {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><script>alert("Email already use, please find another one."); window.location="/signup";</script></body></html>`) // Alert the user if the email is already used
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
	newData := FinalData{data, DisplayPost(w), DisplayCommments(w), DisplayTopics(w)} // Get the data to be displayed
	tmpl.Execute(w, newData)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) { // Function to handle the login
	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/login.html")
	} else if r.Method == "POST" {
		username := r.FormValue("username") // Get the username
		password := r.FormValue("password") // Get the password

		errUser2 := VerifierUtilisateur(username, password) // Verify the user
		if errUser2 != nil {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><script>alert("Username or password incorrect"); window.location="/login";</script></body></html>`) // Alert the user if the username or password is incorrect
			return
		}

		session, _ := store.Get(r, "session") // Get the session
		session.Values["username"] = username // Get the username
		session.Save(r, w) // Save the session

		http.Redirect(w, r, "/", http.StatusSeeOther) // Redirect the user to the home page
	}
}

func UserHandler(w http.ResponseWriter, r *http.Request) { // Function to handle the user
	db = OpenDb() // Open the database
	username := ""
	if r.Method == "POST" {
		username = r.FormValue("username") // Get the username
	} else {
		username = r.URL.Query().Get("username")
	}

	if username == "" {
		http.Error(w, "User not specified", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		var email, profilePicture, firstname, lastname, birthdate string
		query := `SELECT email, profile_picture, firstname, lastname, birthdate FROM utilisateurs WHERE username = ?` // Get the user information
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

		firstname := r.FormValue("Firstname") // Get the firstname
		lastname := r.FormValue("Lastname")	// Get the lastname
		birthdate := r.FormValue("birthdate") // Get the birthdate

		file, handler, errUpload := r.FormFile("profilepicture") // Get the profile picture
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

        updateSQL := `UPDATE utilisateurs SET profile_picture = ?  WHERE username = ?` // Update the profile picture
        _, errQuery29 := db.ExecContext(context.Background(), updateSQL, "./static/uploads/"+handler.Filename, username)
        if errQuery29 != nil {
            http.Error(w, "Error updating the profile picture", http.StatusInternalServerError)
            return
        }


		session, _ := store.Get(r, "session")
		session.Values["username"] = username // Get the username

		_,errQuery30 := db.ExecContext(context.Background(), `UPDATE posts SET profile_picture = ? WHERE author = ?`, "/static/uploads/"+handler.Filename, username) // Update the profile picture
		if errQuery30 != nil {
			http.Error(w, "Error updating the profile picture", http.StatusInternalServerError)
			return
		}		

        http.Redirect(w, r, fmt.Sprintf("/user?username=%s", username), http.StatusSeeOther) // Redirect the user to the user page
	
		updateSQL = `UPDATE utilisateurs SET firstname = ?, lastname = ?, birthdate = ?  WHERE username = ?` // Update the user information
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

func LogoutHandler(w http.ResponseWriter, r *http.Request) { // Function to handle the logout
	session, _ := store.Get(r, "session") // Get the session
	session.Options.MaxAge = -1 
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther) // Redirect the user to the home page
}

func CheckUserInfo(w http.ResponseWriter, r *http.Request) UserInfo { // Function to check the user information
    session, _ := store.Get(r, "session") // Get the session
    username, ok := session.Values["username"] // Get the username

    var data UserInfo
    data.IsLoggedIn = ok
    if ok { // If the user is logged in
        var profilePicture string
        errQuery31 := db.QueryRowContext(context.Background(), "SELECT profile_picture FROM utilisateurs WHERE username = ?", username).Scan(&profilePicture) // Get the profile picture
        if errQuery31 != nil && errQuery31 != sql.ErrNoRows {
            http.Error(w, "Error while retrieving the profile picture", http.StatusInternalServerError)
            return data
        }

        data.Username = username.(string)
        data.ProfilePicture = profilePicture
    }

    return data
}

func RGPDHandler(w http.ResponseWriter, r *http.Request) { // Function to handle the RGPD
	tmpl, errReading18 := template.ParseFiles("templates/RGPD.html")
	if errReading18 != nil {
		http.Error(w, "Error reading the HTML file : RGPD.html", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}