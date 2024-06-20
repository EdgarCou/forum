package forum

import (
	"context"
	"html/template"
	"net/http"
)

type Members struct {
	Username string
	ProfilePicture string
}

type MembersData struct {
	UserInfo UserInfo
	Members []Members
}

func MembersHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
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
	for rows.Next() {
		var member Members
		errScan5 := rows.Scan(&member.Username, &member.ProfilePicture)
		if errScan5 != nil {
			http.Error(w, "Error while reading the members", http.StatusInternalServerError)
			return
		}
		members = append(members, member)
	}

	newData := MembersData{CheckUserInfo(w, r), members}
	tmpl.Execute(w, newData)
}