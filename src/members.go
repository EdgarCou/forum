package forum

import (
	"context"
	"html/template"
	"net/http"
)

type Members struct { // Members structure
	Username string
	ProfilePicture string
}

type MembersData struct { // MembersData structure
	UserInfo UserInfo
	Members []Members
}

func MembersHandler(w http.ResponseWriter, r *http.Request) { // Function for handling the members
	db = OpenDb() // Open the database
	defer db.Close()

	tmpl, errReading7 := template.ParseFiles("templates/members.html")
	if errReading7 != nil {
		http.Error(w, "Error reading the HTML file : members.html", http.StatusInternalServerError)
		return
	}

	rows, errQuery9 := db.QueryContext(context.Background(), "SELECT username, profile_picture FROM utilisateurs")
	if errQuery9 != nil {
		http.Error(w, "Error while retrieving the members", http.StatusInternalServerError)
		return
	}

	var members []Members
	for rows.Next() { // Get the members
		var member Members
		errScan5 := rows.Scan(&member.Username, &member.ProfilePicture)
		if errScan5 != nil {
			http.Error(w, "Error while reading the members", http.StatusInternalServerError)
			return
		}
		members = append(members, member)
	}

	newData := MembersData{CheckUserInfo(w, r), members} // Get the data to be displayed
	tmpl.Execute(w, newData)
}