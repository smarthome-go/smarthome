package user

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
)

// Generates a truly unique and random token and inserts it into the `userToken` table
func AddToken(
	username string,
	label string,
) (token string, err error) {
	// Generate a new random key as while it is taken
	for {
		// Generate a new random key
		token, err = generateRandomToken()
		if err != nil {
			return "", err
		}
		// Check if it already exists in the database
		_, found, err := database.GetUserTokenByToken(token)
		if err != nil {
			return "", err
		}
		// If the new token is not already present in the database, use it
		if !found {
			break
		}
		log.Warn("Random token already exists, generating new one...")
	}
	// After the token has been generated, insert it into the database
	if err := database.InsertUserToken(
		token,
		username,
		label,
	); err != nil {
		return "", err
	}
	log.Info(fmt.Sprintf("User `%s` added a new authentication token named `%s`", username, label))
	return token, nil
}

// Generates a new token without validating if it already exists
func generateRandomToken() (token string, err error) {
	seed := make([]byte, 64)
	if _, err := rand.Read(seed); err != nil {
		return "", err
	}
	hash := md5.Sum(seed)
	return hex.EncodeToString(hash[:]), nil
}
