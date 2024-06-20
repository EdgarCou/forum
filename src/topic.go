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
		http.Error(w, "The fields cannot be empty", http.StatusBadRequest)
		return
	}

	errTopic := AddTopicInDb(topic)
	if errTopic != nil {
		http.Error(w, "Error while adding the topic", http.StatusInternalServerError)
		return
	}

	tmpl, errReading13 := template.ParseFiles("templates/forum.html")
	if errReading13 != nil {
		http.Error(w, "Error reading the HTML file : forum.html", http.StatusInternalServerError)
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
		_,errQuery19 := db.ExecContext(context.Background(), `INSERT INTO topics (title) VALUES (?)`, topic)
		if errQuery19 != nil {
			return errQuery19
		}
	}
	
	return nil
}

func DisplayTopics(w http.ResponseWriter) []Topics {
	db = OpenDb()
	rows, errQuery20 := db.QueryContext(context.Background(), "SELECT title, nbpost FROM topics")
	if errQuery20 != nil {
		http.Error(w, "Error while retrieving the topics", http.StatusInternalServerError)
		return nil
	}
	if rows != nil {
		defer rows.Close()
	}

	var topics []Topics
	for rows.Next() {
		var topic Topics
		errScan7:= rows.Scan(&topic.Title, &topic.NbPost)
		if errScan7 != nil {
			http.Error(w, "Error while reading the topics", http.StatusInternalServerError)
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
			_, errQuery21 := db.ExecContext(context.Background(), `INSERT INTO topics (title) VALUES (?)`, topic)
			if errQuery21 != nil {
				log.Println(errQuery21)
			}
		}
	}
}

func AlreadyInDb() []string {
	db = OpenDb()
	rows, errQuery22 := db.QueryContext(context.Background(), "SELECT title FROM topics")
	if errQuery22!= nil {
		log.Println(errQuery22)
		return nil
	}
	if rows != nil {
		defer rows.Close()
	}

	var topicsInDb []string
	for rows.Next() {
		var topic string
		errScan8 := rows.Scan(&topic)
		if errScan8 != nil {
			log.Println(errScan8)
			return nil
		}
		topicsInDb = append(topicsInDb, topic)
	}
	return topicsInDb
}


func AllTopicsHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	tmpl, errReading14 := template.ParseFiles("templates/topics.html")
	if errReading14 != nil {
		http.Error(w, "Error reading the HTML file : topic.html", http.StatusInternalServerError)
		return
	}
	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w), DisplayCommments(w), DisplayTopics(w)}
	tmpl.Execute(w, newData)
}

func ParticularDisplayTopics(w http.ResponseWriter, particularTopic string) Topics{
	db = OpenDb()
	rows, errQuery23 := db.QueryContext(context.Background(), "SELECT title, nbpost FROM topics WHERE title = ?", particularTopic)
	if errQuery23 != nil {
		http.Error(w, "Error while retrieving the topics", http.StatusInternalServerError)
		return Topics{}
	}
	if rows != nil {
		defer rows.Close()
	}

	var topics Topics
	for rows.Next() {
		errScan9 := rows.Scan(&topics.Title, &topics.NbPost)
		if errScan9 != nil {
			http.Error(w, "Error while reading the topics", http.StatusInternalServerError)
			return Topics{}
		}
	}

	return topics
}

func ParticularHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()

	topic := r.URL.Query().Get("topic")
	tmpl, errReading15 := template.ParseFiles("templates/particularTopic.html")
	if errReading15 != nil {
		http.Error(w, "Error reading the HTML file : particularTopic.html", http.StatusInternalServerError)
		return
	}
	newData := ParticularFinalData{CheckUserInfo(w, r), DisplayPost(w), DisplayCommments(w), ParticularDisplayTopics(w,topic)}
	finalPost := []Post{}
	for _, post := range newData.Posts {
		if(post.Topics != topic){
			continue
		} else {
			finalPost = append(finalPost, post)
		}
	}
	newData.Posts = finalPost
	tmpl.Execute(w, newData)
}

func UpdateTopics(w http.ResponseWriter, currentTopic string, topics string) {
	_,errQuery24 := db.ExecContext(context.Background(), `UPDATE topics SET nbpost = nbpost - 1 WHERE title = ?`, currentTopic)
		if errQuery24 != nil {
			http.Error(w, "Error while updating the post number", http.StatusInternalServerError)
			return
		}
	_,errQuery25 := db.ExecContext(context.Background(), `UPDATE topics SET nbpost = nbpost + 1 WHERE title = ?`, topics)
	if errQuery25 != nil {
		http.Error(w, "Error while updating the post number", http.StatusInternalServerError)
		return
	}
}