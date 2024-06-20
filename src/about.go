package forum

import (
	"html/template"
	"net/http"
)

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	tmpl, errReading := template.ParseFiles("templates/about.html")
	if errReading != nil {
		http.Error(w, "Error reading the HTML file : about.html", http.StatusInternalServerError)
		return
	}
	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w), DisplayCommments(w), DisplayTopics(w)}
	tmpl.Execute(w, newData)
}