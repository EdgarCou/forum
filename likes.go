package forum

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"github.com/gorilla/websocket"
)

func LikeHandlerWs(conn *websocket.Conn, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, _ := session.Values["username"]

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

