package forum

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"sort"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func LikeHandlerWs(conn *websocket.Conn, r *http.Request) {
	session, _ := store.Get(r, "session")
	username:= session.Values["username"]
	db = OpenDb()

	for {
		messageType, p, errReadMessage := conn.ReadMessage()
		if errReadMessage != nil {
			log.Printf("Error reading message: %v. Message type: %v. Message: %v", errReadMessage, messageType, string(p))
			return
		}

		if messageType == websocket.TextMessage {
			message := string(p)
			var id string
			var query string
			var likeType int

			if strings.HasPrefix(message, "dislike:") {
				likeType = 0
				id = strings.TrimPrefix(message, "dislike:")
			} else {
				likeType = 1
				id = strings.TrimPrefix(message, "like:")
			}

			checkQuery := "SELECT COUNT(*) FROM likedBy WHERE username = ? AND idpost = ? AND type = ?"
			row := db.QueryRow(checkQuery, username, id, likeType)
			var count int
			errScan2 := row.Scan(&count)
			if errScan2 != nil {
				log.Printf("Error checking if user has already liked/disliked post: %v", errScan2)
				return
			}
			var countInverse int

			if count == 0 {
				checkInverseQuery := "SELECT COUNT(*) FROM likedBy WHERE username = ? AND idpost = ? AND type = ?"
				row := db.QueryRow(checkInverseQuery, username, id, 1-likeType)
				errScan3 := row.Scan(&countInverse)
				if errScan3 != nil {
					log.Printf("Error checking if user has already liked/disliked post: %v", errScan3)
					return
				}
				if countInverse > 0 {
					var queryRemove string
					var queryDelete string
					if likeType == 0 {
						queryRemove = "UPDATE posts SET likes = likes - 1 WHERE id = ?"
						queryDelete = "DELETE FROM likedBy WHERE username = ? AND idpost = ? AND type = ?"
					} else {
						queryRemove = "UPDATE posts SET dislikes = dislikes - 1 WHERE id = ?"
						queryDelete = "DELETE FROM likedBy WHERE username = ? AND idpost = ? AND type = ?"
					}
					_, errRemove := db.Exec(queryRemove, id)
					if errRemove != nil {
						log.Printf("Error updating likes/dislikes count: %v", errRemove)
					}
					_, errDelete := db.Exec(queryDelete, username, id, 1-likeType)
					if errDelete != nil {
						log.Printf("Error deleting row: %v", errDelete)
					}
				}

				if strings.HasPrefix(message, "dislike:") {
					id = strings.TrimPrefix(message, "dislike:")
					query = "UPDATE posts SET dislikes = dislikes + 1 WHERE id = ?"
					likeType = 0
					_, errQuery2 := db.Exec(query, id)
					if errQuery2 != nil {
						log.Printf("Error updating likes/dislikes count: %v", errQuery2)
					}
				} else {
					id = strings.TrimPrefix(message, "like:")
					query = "UPDATE posts SET likes = likes + 1 WHERE id = ?"
					likeType = 1
					_, errQuery3 := db.Exec(query, id)
					if errQuery3 != nil {
						log.Printf("Error updating likes/dislikes count: %v", errQuery3)
					}
				}

				likeQuery := "INSERT INTO likedBy (username, idpost, type) VALUES (?, ?, ?)"
				_, errQuery4 := db.Exec(likeQuery, username, id, likeType)
				if errQuery4 != nil {
					log.Printf("Error liking/disliking post: %v", errQuery4)
				}

			} else {
				if strings.HasPrefix(message, "dislike:") {
					likeType = 0
				} else {
					likeType = 1
				}
				deleteQuery := "DELETE FROM likedBy WHERE username = ? AND idpost = ? AND type = ?"
				_, errQuery5 := db.Exec(deleteQuery, username, id, likeType)
				if errQuery5 != nil {
					log.Printf("Error liking/disliking post: %v", errQuery5)
				}

				var query string
				if likeType == 0 { 
					query = "UPDATE posts SET dislikes = dislikes - 1 WHERE id = ?"
				} else { 
					query = "UPDATE posts SET likes = likes - 1 WHERE id = ?"
				}
				_, errQuery6 := db.Exec(query, id)
				if errQuery6 != nil {
					log.Printf("Error updating likes/dislikes count: %v", errQuery6)
				}
			}

			var likes, dislikes int
			errQuery7 := db.QueryRowContext(context.Background(), "SELECT likes, dislikes FROM posts WHERE id = ?", id).Scan(&likes, &dislikes)
			if errQuery7 != nil {
				log.Println(errQuery7)
				return
			}

			var response string
			if countInverse > 0 && likeType == 0 {
				response = fmt.Sprintf("dislikes:%s:%d:likes:%s:%d", id, dislikes, id, likes)
			} else if countInverse > 0 && likeType == 1 {
				response = fmt.Sprintf("likes:%s:%d:dislikes:%s:%d", id, likes, id, dislikes)
			} else {
				if likeType == 0 {
					response = fmt.Sprintf("dislikes:%s:%d", id, dislikes)
				} else {
					response = fmt.Sprintf("likes:%s:%d", id, likes)
				}
			}
			errWs := conn.WriteMessage(websocket.TextMessage, []byte(response))
			if errWs != nil {
				log.Println(errWs)
				return
			}
		}
	}
}

func LikedHandler(w http.ResponseWriter, r *http.Request) {
    db = OpenDb()
    tmpl, errReading6 := template.ParseFiles("templates/likedPost.html")
    if errReading6 != nil {
        http.Error(w, "Error reading the HTML file : likedPost.html", http.StatusInternalServerError)
        return
    }

    user := CheckUserInfo(w, r)
    userName := user.Username

    rows, errQuery8 := db.Query(`SELECT posts.* FROM posts JOIN likedBy ON posts.id = likedBy.idpost WHERE likedBy.username = ? AND likedBy.type = 1 AND likedBy.idpost = posts.id`, userName)
    if errQuery8 != nil {
        log.Println(errQuery8)
        return
    }

    var likedPosts []Post
    for rows.Next() {
        var inter Post		
        errScan4 := rows.Scan(&inter.Id, &inter.Title, &inter.Content, &inter.Topics, &inter.Author, &inter.Likes, &inter.Dislikes, &inter.Date, &inter.Comments)
        if errScan4 != nil {
            log.Println(errScan4)
            return
        }
		inter.Date = inter.Date[:16]
        likedPosts = append(likedPosts, inter)
    }

	likedPosts = SortLikedPost(likedPosts, w, r)
    newData := FinalData{user, likedPosts, DisplayCommments(w), DisplayTopics(w)}

    tmpl.Execute(w, newData)
}

func SortLikedPost(likedPosts []Post, w http.ResponseWriter, r *http.Request) []Post {

	sortType := r.FormValue("sort")

	if sortType == "mostLiked" {
		sort.Slice(likedPosts, func(i, j int) bool {
			return likedPosts[i].Likes > likedPosts[j].Likes
		})
	} else if sortType == "mostDisliked" {
		sort.Slice(likedPosts, func(i, j int) bool {
			return likedPosts[i].Dislikes > likedPosts[j].Dislikes
		})
	} else if sortType == "newest" {
		sort.Slice(likedPosts, func(i, j int) bool {
			return likedPosts[i].Date > likedPosts[j].Date
		})
	} else if sortType == "oldest" {
		sort.Slice(likedPosts, func(i, j int) bool {
			return likedPosts[i].Date < likedPosts[j].Date
		})
	}else if sortType == "A-Z" {
		sort.Slice(likedPosts, func(i, j int) bool {
			return likedPosts[i].Title < likedPosts[j].Title
		})
	}else if sortType == "Z-A" {
		sort.Slice(likedPosts, func(i, j int) bool {
			return likedPosts[i].Title > likedPosts[j].Title
		}) 
	} else {
		return likedPosts
	}

	return likedPosts
}