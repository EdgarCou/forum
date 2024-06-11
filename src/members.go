package forum

import (
	"html/template"
	"net/http"
)

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