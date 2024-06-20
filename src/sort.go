package forum

import (
	"html/template"
	"net/http"
	"sort"
)

func SortHandler(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	tmpl, errReading11 := template.ParseFiles("templates/forum.html")
	if errReading11 != nil {
		http.Error(w, "Error reading the HTML file : forum.html", http.StatusInternalServerError)
		return
	}

	sortType := r.FormValue("sort")

	var posts []Post
	posts = sortPost(sortType, w, posts)
	newData := FinalData{CheckUserInfo(w, r), posts, DisplayCommments(w), DisplayTopics(w)}

	tmpl.Execute(w, newData)
}


func SortHandlerMyPost(w http.ResponseWriter, r *http.Request) {
	db = OpenDb()
	tmpl, errReading12 := template.ParseFiles("templates/myPost.html")
	if errReading12 != nil {
		http.Error(w, "Error reading the HTML file : myPost.html", http.StatusInternalServerError)
		return
	}

	sortType := r.FormValue("sort")
	var posts []Post
	posts = sortPost(sortType, w, posts)

	newData := FinalData{CheckUserInfo(w, r), posts, DisplayCommments(w), DisplayTopics(w)}

	tmpl.Execute(w, newData)
}


func sortPost(sortType string, w http.ResponseWriter, posts []Post) []Post  {
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
		http.Error(w, "Invalid sort ", http.StatusBadRequest)
		return posts
	}

	return posts

}