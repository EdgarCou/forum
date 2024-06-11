package forum

import (
	"html/template"
	"net/http"
)

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