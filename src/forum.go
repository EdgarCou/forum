package forum

import (
	"html/template"
	"net/http"
)

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