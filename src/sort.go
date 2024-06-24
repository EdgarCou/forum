package forum

import (
	"html/template"
	"net/http"
	"sort"
)

func SortHandler(w http.ResponseWriter, r *http.Request) { // Function to handle the sorting
	db = OpenDb() // Open the database
	tmpl, errReading11 := template.ParseFiles("templates/forum.html")
	if errReading11 != nil {
		http.Error(w, "Error reading the HTML file : forum.html", http.StatusInternalServerError)
		return
	}

	sortType := r.FormValue("sort") // Get the sort type

	var posts []Post
	posts = sortPost(sortType, w, posts) // Sort the posts
	newData := FinalData{CheckUserInfo(w, r), posts, DisplayCommments(w), DisplayTopics(w)} // Get the data to be displayed

	tmpl.Execute(w, newData)
}


func SortHandlerMyPost(w http.ResponseWriter, r *http.Request) { // Function to handle the sorting of the user posts
	db = OpenDb() // Open the database
	tmpl, errReading12 := template.ParseFiles("templates/myPost.html")
	if errReading12 != nil {
		http.Error(w, "Error reading the HTML file : myPost.html", http.StatusInternalServerError)
		return
	}

	sortType := r.FormValue("sort") // Get the sort type
	var posts []Post
	posts = sortPost(sortType, w, posts) // Sort the posts

	newData := FinalData{CheckUserInfo(w, r), posts, DisplayCommments(w), DisplayTopics(w)} // Get the data to be displayed

	tmpl.Execute(w, newData)
}


func sortPost(sortType string, w http.ResponseWriter, posts []Post) []Post  { // Function to sort the posts
	if sortType == "mostLiked" { // Sort by the most liked posts
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool { 
			if (posts[i].Likes == posts[j].Likes) { // If the likes are the same
				return posts[i].Dislikes < posts[j].Dislikes // Sort by the less disliked
			}
			return posts[i].Likes > posts[j].Likes
		})
	} else if sortType == "mostDisliked" { // Sort by the most disliked posts
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			if (posts[i].Dislikes == posts[j].Dislikes) { // If the dislikes are the same
				return posts[i].Likes < posts[j].Likes // Sort by the less liked
			}
			return posts[i].Dislikes > posts[j].Dislikes
		})
	} else if sortType == "newest" { // Sort by the newest posts
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Date > posts[j].Date
		})
	} else if sortType == "oldest" { // Sort by the oldest posts
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Date < posts[j].Date
		})
	}else if sortType == "A-Z" { // Sort by the title A-Z
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool { 
			return posts[i].Title < posts[j].Title
		})
	}else if sortType == "Z-A" { // Sort by the title Z-A
		posts = DisplayPost(w)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Title > posts[j].Title
		}) 
	} else {
		http.Error(w, "Invalid sort ", http.StatusBadRequest) // If the sort type is invalid
		return posts
	}

	return posts

}