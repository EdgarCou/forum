package forum

import (
	"context"
	//"slices"
	//"fmt"
	"html/template"
	"net/http"
	"time"
	//"fmt"
)




func CommentHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"]

	if !ok {
		http.Error(w, "Vous devez être connecté pour commenter", http.StatusUnauthorized)
		return
	}


	id := r.FormValue("postId")
	println(id)
	content := r.FormValue("comment")

	if id == "" || content == "" {
		http.Error(w, "Les champs ne peuvent pas être vides", http.StatusBadRequest)
		return
	}

	err := AddCommentInDb(content, username.(string), id)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout du commentaire", http.StatusInternalServerError)
		return
	}

	date := time.Now()
	date_string := date.Format("2006-01-02 15:04:05")
	_, err = db.ExecContext(context.Background(), `UPDATE posts SET date = ? WHERE id = ?`, date_string, id)
	if err != nil {
		http.Error(w, "Erreur lors de la mise à jour de la date", http.StatusInternalServerError)
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


func AddCommentInDb(content, author, id string) error{
	_, err := db.ExecContext(context.Background(), "INSERT INTO comments (content, author, idpost) VALUES (?, ?, ?)", content, author, id)
	_,err2 := db.ExecContext(context.Background(), "UPDATE posts SET comments = comments + 1 WHERE id = ?", id)
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}

	return nil
}

func DisplayCommments(w http.ResponseWriter) []Comment {
	db = OpenDb()
	rows, err := db.QueryContext(context.Background(), "SELECT content, author, idpost FROM comments")
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des commentaires", http.StatusInternalServerError)
		return nil
	}
	if rows != nil {
		defer rows.Close()
	} 

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.Content, &comment.Author, &comment.Idpost)
		if err != nil {
			http.Error(w, "Erreur lors de la lecture des commentaires", http.StatusInternalServerError)
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
		http.Error(w, "Vous devez être connecté pour commenter", http.StatusUnauthorized)
		return
	}


	id := r.FormValue("postId")
	println(id)
	content := r.FormValue("comment")

	if id == "" || content == "" {
		http.Error(w, "Les champs ne peuvent pas être vides", http.StatusBadRequest)
		return
	}

	err := AddCommentInDb(content, username.(string), id)
	if err != nil {
		http.Error(w, "Erreur lors de l'ajout du commentaire", http.StatusInternalServerError)
		return
	}

	date := time.Now()
	date_string := date.Format("2006-01-02 15:04:05")
	_, err = db.ExecContext(context.Background(), `UPDATE posts SET date = ? WHERE id = ?`, date_string, id)
	if err != nil {
		http.Error(w, "Erreur lors de la mise à jour de la date", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/myPost.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 9", http.StatusInternalServerError)
		return
	}

	newData := FinalData{CheckUserInfo(w, r), DisplayPost(w),DisplayCommments(w), DisplayTopics(w)}
	tmpl.Execute(w, newData)
}