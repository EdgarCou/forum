package forum

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strings"
)



func AddTopicHandler(w http.ResponseWriter, r *http.Request) { // Function to handle the addition of a topic
	db = OpenDb() // Open the database

	topic := r.FormValue("topic") // Get the topic name

	if topic == "" {
		http.Error(w, "The fields cannot be empty", http.StatusBadRequest)
		return
	}

	errTopic := AddTopicInDb(topic) // Add the topic in the database
	if errTopic != nil {
		http.Error(w, "Error while adding the topic", http.StatusInternalServerError)
		return
	}

	tmpl, errReading13 := template.ParseFiles("templates/forum.html")
	if errReading13 != nil {
		http.Error(w, "Error reading the HTML file : forum.html", http.StatusInternalServerError)
		return
	}

	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w),DisplayCommments(w), DisplayTopics(w)} // Get the data to be displayed
	tmpl.Execute(w, newData)
}


func AddTopicInDb(topic string) error { // Function to add a topic in the database
	db = OpenDb() // Open the database

	topicInDb := AlreadyInDb() // Check if the topic is already in the database
	found := false

	for _, t := range topicInDb {
		if strings.EqualFold(t, topic) {
			found = true
		}
	}

	if !found {
		_,errQuery19 := db.ExecContext(context.Background(), `INSERT INTO topics (title) VALUES (?)`, topic) // Insert the topic in the database
		if errQuery19 != nil {
			return errQuery19
		}
	}
	
	return nil
}

func DisplayTopics(w http.ResponseWriter) []Topics { // Function to display the topics
	db = OpenDb()
	rows, errQuery20 := db.QueryContext(context.Background(), "SELECT title, nbpost FROM topics") // Get the topics from the database
	if errQuery20 != nil {
		http.Error(w, "Error while retrieving the topics", http.StatusInternalServerError)
		return nil
	}
	if rows != nil {
		defer rows.Close()
	}

	var topics []Topics
	for rows.Next() { // Get the topics
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

func InitTopics() { // Function to initialize some topics
	db = OpenDb() // Open the database
	topics := []string{"Sport", "Music", "Cinema", "Science", "Technology", "Politics", "Economy", "Art", "Literature", "History", "Travel", "Cooking"} // Topics automatically added

	topicsInDb := AlreadyInDb() // Check if the topics are already in the database
	
	if topicsInDb == nil {
		for _, topic := range topics {
			_, errQuery21 := db.ExecContext(context.Background(), `INSERT INTO topics (title) VALUES (?)`, topic) // Insert the topics in the database
			if errQuery21 != nil {
				log.Println(errQuery21)
			}
		}
	}
}

func AlreadyInDb() []string { // Function to check if the topics are already in the database
	db = OpenDb()
	rows, errQuery22 := db.QueryContext(context.Background(), "SELECT title FROM topics") // Get all the topics from the database
	if errQuery22!= nil {
		log.Println(errQuery22)
		return nil
	}
	if rows != nil {
		defer rows.Close()
	}

	var topicsInDb []string
	for rows.Next() { // Get the topics
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


func AllTopicsHandler(w http.ResponseWriter, r *http.Request) { // Function to handle all the topics
	db = OpenDb()
	tmpl, errReading14 := template.ParseFiles("templates/topics.html")
	if errReading14 != nil {
		http.Error(w, "Error reading the HTML file : topic.html", http.StatusInternalServerError)
		return
	}
	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w), DisplayCommments(w), DisplayTopics(w)} // Get the data to be displayed
	tmpl.Execute(w, newData)
}

func ParticularDisplayTopics(w http.ResponseWriter, particularTopic string) Topics{ // Function to display a particular topic
	db = OpenDb() // Open the database
	rows, errQuery23 := db.QueryContext(context.Background(), "SELECT title, nbpost FROM topics WHERE title = ?", particularTopic) // Get the topics from the database
	if errQuery23 != nil {
		http.Error(w, "Error while retrieving the topics", http.StatusInternalServerError)
		return Topics{}
	}
	if rows != nil {
		defer rows.Close()
	}

	var topics Topics
	for rows.Next() { // Get the topics
		errScan9 := rows.Scan(&topics.Title, &topics.NbPost)
		if errScan9 != nil {
			http.Error(w, "Error while reading the topics", http.StatusInternalServerError)
			return Topics{}
		}
	}

	return topics
}

func ParticularHandler(w http.ResponseWriter, r *http.Request) { // Function to handle a particular topic
	db = OpenDb() // Open the database

	topic := r.URL.Query().Get("topic") // Get the topic name
	tmpl, errReading15 := template.ParseFiles("templates/particularTopic.html")
	if errReading15 != nil {
		http.Error(w, "Error reading the HTML file : particularTopic.html", http.StatusInternalServerError)
		return
	}
	newData := ParticularFinalData{CheckUserInfo(w, r), DisplayPost(w), DisplayCommments(w), ParticularDisplayTopics(w,topic)} // Get the data to be displayed
	finalPost := []Post{}
	for _, post := range newData.Posts { // Get the posts of the topic
		if(post.Topics != topic){
			continue
		} else {
			finalPost = append(finalPost, post)
		}
	}
	newData.Posts = finalPost
	tmpl.Execute(w, newData)
}

func UpdateTopics(w http.ResponseWriter, currentTopic string, topics string) { // Function to update the topics
	_,errQuery24 := db.ExecContext(context.Background(), `UPDATE topics SET nbpost = nbpost - 1 WHERE title = ?`, currentTopic) // Update the number of posts of the topic
		if errQuery24 != nil {
			http.Error(w, "Error while updating the post number", http.StatusInternalServerError)
			return
		}
	_,errQuery25 := db.ExecContext(context.Background(), `UPDATE topics SET nbpost = nbpost + 1 WHERE title = ?`, topics) // Update the number of posts of the topic
	if errQuery25 != nil {
		http.Error(w, "Error while updating the post number", http.StatusInternalServerError)
		return
	}
}