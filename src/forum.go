package forum

import (
	"html/template"
	"net/http"
)

func ForumHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()

	tmpl, errReading4 := template.ParseFiles("templates/forum.html")
	if errReading4 != nil {
		http.Error(w, "Error reading the HTML file : forum.html", http.StatusInternalServerError)
		return
	}

	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w), DisplayCommments(w), DisplayTopics(w)}
	tmpl.Execute(w, newData)
}