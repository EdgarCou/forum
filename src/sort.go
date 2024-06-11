package forum

import (
	"html/template"
	"net/http"
	"sort"
	"database/sql"
)

func SortHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	tmpl, err := template.ParseFiles("templates/forum.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 10", http.StatusInternalServerError)
		return
	}

	sortType := r.FormValue("sort")

	var rows *sql.Rows
	var posts []Post

	if sortType == "mostLiked" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			if (posts[i].Likes == posts[j].Likes) {
				return posts[i].Dislikes < posts[j].Dislikes
			}
			return posts[i].Likes > posts[j].Likes
		})
	} else if sortType == "mostDisliked" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			if (posts[i].Dislikes == posts[j].Dislikes) {
				return posts[i].Likes < posts[j].Likes
			}
			return posts[i].Dislikes > posts[j].Dislikes
		})
	} else if sortType == "newest" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Date > posts[j].Date
		})
	} else if sortType == "oldest" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Date < posts[j].Date
		})
	}else if sortType == "A-Z" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Title < posts[j].Title
		})
	}else if sortType == "Z-A" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Title > posts[j].Title
		}) 
	} else {
		http.Error(w, "Tri invalide", http.StatusBadRequest)
		return
	}

	if rows != nil {
		defer rows.Close()
	}

	newData := FinalData{CheckUserInfo(w, r), posts, DisplayCommments(w), DisplayTopics(w)}

	tmpl.Execute(w, newData)
}


func SortHandlerMyPost(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	tmpl, err := template.ParseFiles("templates/myPost.html")
	if err != nil {
		http.Error(w, "Erreur de lecture du fichier HTML 10", http.StatusInternalServerError)
		return
	}

	sortType := r.FormValue("sort")

	var rows *sql.Rows
	var posts []Post

	if sortType == "mostLiked" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Likes > posts[j].Likes
		})
	} else if sortType == "mostDisliked" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Dislikes > posts[j].Dislikes
		})
	} else if sortType == "newest" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Date > posts[j].Date
		})
	} else if sortType == "oldest" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Date < posts[j].Date
		})
	}else if sortType == "A-Z" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Title < posts[j].Title
		})
	}else if sortType == "Z-A" {
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Title > posts[j].Title
		}) 
	} else {
		http.Error(w, "Tri invalide", http.StatusBadRequest)
		return
	}

	if rows != nil {
		defer rows.Close()
	}

	newData := FinalData{CheckUserInfo(w, r), posts, DisplayCommments(w), DisplayTopics(w)}

	tmpl.Execute(w, newData)
}