package user

import (
	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// Will return <true / false> based on authentication validity
// <true> means valid authentication
// Can return an error if the database fails to return a valid result, meaning service downtime
func ValidateCredentials(username string, password string) (bool, error) {
	users, err := database.ListUsers()
	if err != nil {
		log.Error("Could not validate login due to database error: ", err.Error())
		return false, err
	}
	for _, user := range users {
		if user.Username == username {
			// Check hash equality
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err == nil {
				return true, nil
			}
			if err.Error() != "crypto/bcrypt: hashedPassword is not the hash of the given password" {
				log.Error("failed to check password: ", err.Error())
				return false, err
			}
			log.Trace("password check using bcrypt failed: passwords do not match")
		}
	}
	return false, nil
}
