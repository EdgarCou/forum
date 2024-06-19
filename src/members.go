package forum

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
)

type Members struct {
	Username string
	Photo string
}

type MembersData struct {
	UserInfo UserInfo
	Members []Members
}

func MembersHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	defer db.Close()

	tmpl, err := template.ParseFiles("templates/members.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 7", http.StatusInternalServerError)
		return
	}

	rows, err := db.QueryContext(context.Background(), "SELECT username, profile_picture FROM utilisateurs")
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des membres", http.StatusInternalServerError)
		return
	}

	var members []Members
	for rows.Next() {
		var member Members
		err := rows.Scan(&member.Username, &member.Photo)
		if err != nil {
			http.Error(w, "Erreur lors de la lecture des membres", http.StatusInternalServerError)
			return
		}
		members = append(members, member)
	}

	infos := CheckUserInfo(w, r)
	fmt.Println(infos)

	newData := MembersData{CheckUserInfo(w, r), members}
	tmpl.Execute(w, newData)
}