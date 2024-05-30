package forum

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)


func AjouterUtilisateur(username, email, password, profilePicture string) error {
	db = OpenDb()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `INSERT INTO utilisateurs (username, email, password, profile_picture) VALUES (?, ?, ?, ?)`,
		username, email, hashedPassword, profilePicture)
	if err != nil {
		return err
	}
	return nil
}

func VerifierUtilisateur(username, password string) error {
	db = OpenDb()
	var passwordDB string
	err := db.QueryRowContext(context.Background(), "SELECT password FROM utilisateurs WHERE username = ?", username).Scan(&passwordDB)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordDB), []byte(password))
	if err != nil {
		return fmt.Errorf("mot de passe incorrect")
	}
	return nil
}