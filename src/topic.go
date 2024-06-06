package forum

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strings"
)



func AddTopicHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()

	topic := r.FormValue("topic")

	if topic == "" {
		http.Error(w, "Les champs ne peuvent pas être vides", http.StatusBadRequest)
		return
	}

	err := AddTopicInDb(topic)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout du topic", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/forum.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 9", http.StatusInternalServerError)
		return
	}

	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w),DisplayCommments(w), DisplayTopics(w)}
	tmpl.Execute(w, newData)
}


func AddTopicInDb(topic string) error {
	db = OpenDb()

	topicInDb := AlreadyInDb()
	found := false

	for _, t := range topicInDb {
		if strings.EqualFold(t, topic) {
			found = true
		}
	}

	if !found {
		_,err := db.ExecContext(context.Background(), `INSERT INTO topics (title) VALUES (?)`, topic)
		if err != nil {
			return err
		}
	}
	
	return nil
}

func DisplayTopics(w http.ResponseWriter) []Topics {
	db = OpenDb()
	rows, err := db.QueryContext(context.Background(), "SELECT title FROM topics")
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des topics", http.StatusInternalServerError)
		return nil
	}
	defer rows.Close()

	var topics []Topics
	for rows.Next() {
		var topic Topics
		err := rows.Scan(&topic.Title)
		if err != nil {
			http.Error(w, "Erreur lors de la lecture des topics", http.StatusInternalServerError)
			return nil
		}
		topics = append(topics, topic)
	}
	return topics
}

func InitTopics() {
	db = OpenDb()
	topics := []string{"Sport", "Music", "Cinema", "Science", "Technology", "Politics", "Economy", "Art", "Literature", "History", "Travel", "Cooking"}

	topicsInDb := AlreadyInDb()
	
	if topicsInDb == nil {
		for _, topic := range topics {
			_, err := db.ExecContext(context.Background(), `INSERT INTO topics (title) VALUES (?)`, topic)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func AlreadyInDb() []string {
	db = OpenDb()
	rows, err := db.QueryContext(context.Background(), "SELECT title FROM topics")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	var topicsInDb []string
	for rows.Next() {
		var topic string
		err := rows.Scan(&topic)
		if err != nil {
			log.Println(err)
			return nil
		}
		topicsInDb = append(topicsInDb, topic)
	}
	return topicsInDb
}