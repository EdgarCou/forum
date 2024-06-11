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
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v. Message type: %v. Message: %v", err, messageType, string(p))
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
			err := row.Scan(&count)
			if err != nil {
				log.Printf("Error checking if user has already liked/disliked post: %v", err)
				return
			}
			var countInverse int
			// Si l'utilisateur n'a pas déjà aimé ou n'a pas aimé ce post avec le même type, insérez une nouvelle ligne
			if count == 0 {
				checkInverseQuery := "SELECT COUNT(*) FROM likedBy WHERE username = ? AND idpost = ? AND type = ?"
				row := db.QueryRow(checkInverseQuery, username, id, 1-likeType)
				err9 := row.Scan(&countInverse)
				if err9 != nil {
					log.Printf("Error checking if user has already liked/disliked post: %v", err9)
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
					_, err7 := db.Exec(query, id)
					if err7 != nil {
						log.Printf("Error updating likes/dislikes count: %v", err)
					}
				} else {
					id = strings.TrimPrefix(message, "like:")
					query = "UPDATE posts SET likes = likes + 1 WHERE id = ?"
					likeType = 1
					_, err8 := db.Exec(query, id)
					if err8 != nil {
						log.Printf("Error updating likes/dislikes count: %v", err)
					}
				}

				likeQuery := "INSERT INTO likedBy (username, idpost, type) VALUES (?, ?, ?)"
				_, err := db.Exec(likeQuery, username, id, likeType)
				if err != nil {
					log.Printf("Error liking/disliking post: %v", err)
				}

			} else {
				if strings.HasPrefix(message, "dislike:") {
					likeType = 0
				} else {
					likeType = 1
				}
				deleteQuery := "DELETE FROM likedBy WHERE username = ? AND idpost = ? AND type = ?"
				_, err := db.Exec(deleteQuery, username, id, likeType)
				if err != nil {
					log.Printf("Error unliking/undisliking post: %v", err)
				}

				// Mettez à jour le nombre de likes ou de dislikes dans la table posts
				var query string
				if likeType == 0 { // Si le type est 0 (dislike), décrémentez le nombre de dislikes
					query = "UPDATE posts SET dislikes = dislikes - 1 WHERE id = ?"
				} else { // Si le type est 1 (like), décrémentez le nombre de likes
					query = "UPDATE posts SET likes = likes - 1 WHERE id = ?"
				}
				_, err = db.Exec(query, id)
				if err != nil {
					log.Printf("Error updating likes/dislikes count: %v", err)
				}
			}

			// Get the new number of likes or dislikes
			var likes, dislikes int
			err = db.QueryRowContext(context.Background(), "SELECT likes, dislikes FROM posts WHERE id = ?", id).Scan(&likes, &dislikes)
			if err != nil {
				log.Println(err)
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
			err = conn.WriteMessage(websocket.TextMessage, []byte(response))
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func LikedHandler(w http.ResponseWriter, r *http.Request) {
    db = OpenDb()
    tmpl, err := template.ParseFiles("templates/likedPost.html")
    if err != nil {
        http.Error(w, "Erreur de lecture du fichier HTML 11", http.StatusInternalServerError)
        return
    }

    // Obtenir l'ID de l'utilisateur connecté
    user := CheckUserInfo(w, r)
    userName := user.Username

    // Interroger la base de données pour obtenir les posts aimés par l'utilisateur
    rows, err := db.Query(`SELECT posts.* FROM posts JOIN likedBy ON posts.id = likedBy.idpost WHERE likedBy.username = ? AND likedBy.type = 1 AND likedBy.idpost = posts.id`, userName)
    if err != nil {
        log.Println(err)
        return
    }



    // Créer une slice pour stocker les posts
    var likedPosts []Post
    for rows.Next() {
        var inter Post		
        err = rows.Scan(&inter.Id, &inter.Title, &inter.Content, &inter.Topics, &inter.Author, &inter.Likes, &inter.Dislikes, &inter.Date, &inter.Comments)
        if err != nil {
            log.Println(err)
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