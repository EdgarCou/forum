package forum

import (
	"context"
	"html/template"
	"net/http"
	"time"
)




func CommentHandler(w http.ResponseWriter, r *http.Request) { 
	db = OpenDb() // Open the database
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"] // Get the username from the session

	if !ok {
		http.Error(w, "You must be connected to comment", http.StatusUnauthorized)
		return
	}

	id := r.FormValue("postId") // Get the post id from the form
	content := r.FormValue("comment")// Get the comment content from the form

	if id == "" || content == "" {
		http.Error(w, "Fields cannot be empty", http.StatusBadRequest)
		return
	}

	errComment := AddCommentInDb(content, username.(string), id) // Add the comment in the database
	if errComment != nil {
		http.Error(w, "Error while adding the comment", http.StatusInternalServerError)
		return
	}

	date := time.Now() // Get the current date
	date_string := date.Format("2006-01-02 15:04:05")

	_, errUpdate := db.ExecContext(context.Background(), `UPDATE posts SET date = ? WHERE id = ?`, date_string, id) // Update the date of the post
	if errUpdate != nil {
		http.Error(w, "Error while updating the data", http.StatusInternalServerError)
		return
	}

	tmpl, errReading2 := template.ParseFiles("templates/forum.html")
	if errReading2 != nil {
		http.Error(w, "Error reading the HTML file : forum.html", http.StatusInternalServerError)
		return
	}

	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w),DisplayCommments(w), DisplayTopics(w)} // Get the data to be displayed
	tmpl.Execute(w, newData)
}


func AddCommentInDb(content, author, id string) error{ // Function for adding a comment in the database
	_, errInsert := db.ExecContext(context.Background(), "INSERT INTO comments (content, author, idpost) VALUES (?, ?, ?)", content, author, id) // Insert the comment in the database
	_,errUpdate2 := db.ExecContext(context.Background(), "UPDATE posts SET comments = comments + 1 WHERE id = ?", id) // Update the number of comments of the post
	if errInsert != nil {
		return errInsert
	}
	if errUpdate2 != nil {
		return errUpdate2
	}

	return nil
}

func DisplayCommments(w http.ResponseWriter) []Comment { // Function for displaying the comments
	db = OpenDb() // Open the database
	rows, errQuery := db.QueryContext(context.Background(), "SELECT content, author, idpost FROM comments") // Get the comments from the database
	if errQuery != nil {
		http.Error(w, "Error while retrieving the comments", http.StatusInternalServerError)
		return nil
	}

	if rows != nil {
		defer rows.Close()
	} 

	var comments []Comment
	for rows.Next() {
		var comment Comment
		errScan := rows.Scan(&comment.Content, &comment.Author, &comment.Idpost) // Get the comments
		if errScan != nil {
			http.Error(w, "Error while reading the comments", http.StatusInternalServerError)
			return nil
		}
		comments = append(comments, comment)
	}
	
	for i := len(comments)/2-1; i >= 0; i-- { // Reverse the comments order
        opp := len(comments)-1-i
        comments[i], comments[opp] = comments[opp], comments[i]
    }

	return comments
}

func CommentHandlerForMyPost(w http.ResponseWriter, r *http.Request) { // Function for adding a comment in the user's post page
	db = OpenDb() // Open the database
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"] // Get the username from the session

	if !ok {
		http.Error(w, "You must be logged in to comment", http.StatusUnauthorized)
		return
	}

	id := r.FormValue("postId") // Get the post id from the form
	content := r.FormValue("comment") // Get the comment content from the form

	if id == "" || content == "" {
		http.Error(w, "The fields cannot be empty", http.StatusBadRequest)
		return
	}

	errComment2 := AddCommentInDb(content, username.(string), id) // Add the comment in the database
	if errComment2 != nil {
		http.Error(w, "Error while adding the comment", http.StatusInternalServerError)
		return
	}

	date := time.Now() // Get the current date
	date_string := date.Format("2006-01-02 15:04:05")

	_, errUpdate3 := db.ExecContext(context.Background(), `UPDATE posts SET date = ? WHERE id = ?`, date_string, id) // Update the date of the post
	if errUpdate3 != nil {
		http.Error(w, "Error while updating the date", http.StatusInternalServerError)
		return
	}

	tmpl, errReading3 := template.ParseFiles("templates/myPost.html")
	if errReading3 != nil {
		http.Error(w, "Error reading the HTML file : myPost.html", http.StatusInternalServerError)
		return
	}

	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w),DisplayCommments(w), DisplayTopics(w)}
	tmpl.Execute(w, newData)
}

func CommentHandlerParticularTopic(w http.ResponseWriter, r *http.Request) { // Function for adding a comment in the particular topic page
	db = OpenDb()
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"] // Get the username from the session

	if !ok {
		http.Error(w, "You must be logged in to comment", http.StatusUnauthorized)
		return
	}

	id := r.FormValue("postId") // Get the post id from the form
	content := r.FormValue("comment") // Get the comment content from the form
	topic := r.FormValue("topic") // Get the topic from the form

	if id == "" || content == "" {
		http.Error(w, "The fields cannot be empty", http.StatusBadRequest)
		return
	}

	errComment3 := AddCommentInDb(content, username.(string), id)
	if errComment3 != nil {
		http.Error(w, "Error while adding the comment", http.StatusInternalServerError)
		return
	}

	date := time.Now() // Get the current date
	date_string := date.Format("2006-01-02 15:04:05")

	_, errUpdate4 := db.ExecContext(context.Background(), `UPDATE posts SET date = ? WHERE id = ?`, date_string, id) // Update the date of the post
	if errUpdate4 != nil {
		http.Error(w, "Error while updating the date", http.StatusInternalServerError)
		return
	}

	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w),DisplayCommments(w), DisplayTopics(w)} // Get the data to be displayed


	http.Redirect(w, r, "/particular?topic="+topic, http.StatusSeeOther) // Redirect to the particular topic page

	tmpl, err := template.ParseFiles("templates/particularTopic.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file : particularTopic.html", http.StatusInternalServerError)
		return
	}
	
	tmpl.Execute(w, newData)
}