package forum

import (
	"context"
	"html/template"
	"net/http"
	"time"
)




func CommentHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]

	if !ok {
		http.Error(w, "You must be connected to comment", http.StatusUnauthorized)
		return
	}

	id := r.FormValue("postId")
	content := r.FormValue("comment")

	if id == "" || content == "" {
		http.Error(w, "Fields cannot be empty", http.StatusBadRequest)
		return
	}

	errComment := AddCommentInDb(content, username.(string), id)
	if errComment != nil {
		http.Error(w, "Error while adding the comment", http.StatusInternalServerError)
		return
	}

	date := time.Now()
	date_string := date.Format("2006-01-02 15:04:05")

	_, errUpdate := db.ExecContext(context.Background(), `UPDATE posts SET date = ? WHERE id = ?`, date_string, id)
	if errUpdate != nil {
		http.Error(w, "Error while updating the data", http.StatusInternalServerError)
		return
	}

	tmpl, errReading2 := template.ParseFiles("templates/forum.html")
	if errReading2 != nil {
		http.Error(w, "Error reading the HTML file : forum.html", http.StatusInternalServerError)
		return
	}

	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w),DisplayCommments(w), DisplayTopics(w)}
	tmpl.Execute(w, newData)
}


func AddCommentInDb(content, author, id string) error{
	_, errInsert := db.ExecContext(context.Background(), "INSERT INTO comments (content, author, idpost) VALUES (?, ?, ?)", content, author, id)
	_,errUpdate2 := db.ExecContext(context.Background(), "UPDATE posts SET comments = comments + 1 WHERE id = ?", id)
	if errInsert != nil {
		return errInsert
	}
	if errUpdate2 != nil {
		return errUpdate2
	}

	return nil
}

func DisplayCommments(w http.ResponseWriter) []Comment {
	db = OpenDb()
	rows, errQuery := db.QueryContext(context.Background(), "SELECT content, author, idpost FROM comments")
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
		errScan := rows.Scan(&comment.Content, &comment.Author, &comment.Idpost)
		if errScan != nil {
			http.Error(w, "Error while reading the comments", http.StatusInternalServerError)
			return nil
		}
		comments = append(comments, comment)
	}
	
	for i := len(comments)/2-1; i >= 0; i-- {
        opp := len(comments)-1-i
        comments[i], comments[opp] = comments[opp], comments[i]
    }

	return comments
}

func CommentHandlerForMyPost(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]

	if !ok {
		http.Error(w, "You must be logged in to comment", http.StatusUnauthorized)
		return
	}

	id := r.FormValue("postId")
	content := r.FormValue("comment")

	if id == "" || content == "" {
		http.Error(w, "The fields cannot be empty", http.StatusBadRequest)
		return
	}

	errComment2 := AddCommentInDb(content, username.(string), id)
	if errComment2 != nil {
		http.Error(w, "Error while adding the comment", http.StatusInternalServerError)
		return
	}

	date := time.Now()
	date_string := date.Format("2006-01-02 15:04:05")

	_, errUpdate3 := db.ExecContext(context.Background(), `UPDATE posts SET date = ? WHERE id = ?`, date_string, id)
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

func CommentHandlerParticularTopic(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]

	if !ok {
		http.Error(w, "You must be logged in to comment", http.StatusUnauthorized)
		return
	}

	id := r.FormValue("postId")
	content := r.FormValue("comment")
	topic := r.FormValue("topic")

	if id == "" || content == "" {
		http.Error(w, "The fields cannot be empty", http.StatusBadRequest)
		return
	}

	errComment3 := AddCommentInDb(content, username.(string), id)
	if errComment3 != nil {
		http.Error(w, "Error while adding the comment", http.StatusInternalServerError)
		return
	}

	date := time.Now()
	date_string := date.Format("2006-01-02 15:04:05")

	_, errUpdate4 := db.ExecContext(context.Background(), `UPDATE posts SET date = ? WHERE id = ?`, date_string, id)
	if errUpdate4 != nil {
		http.Error(w, "Error while updating the date", http.StatusInternalServerError)
		return
	}

	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w),DisplayCommments(w), DisplayTopics(w)}


	http.Redirect(w, r, "/particular?topic="+topic, http.StatusSeeOther)

	tmpl, err := template.ParseFiles("templates/particularTopic.html")
	if err != nil {
		http.Error(w, "Error reading the HTML file : particularTopic.html", http.StatusInternalServerError)
		return
	}
	
	tmpl.Execute(w, newData)
}