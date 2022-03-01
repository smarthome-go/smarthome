package user

import (
	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// Will return <true / false> based on authentication validity
// <true> means valid authentication
// Can return an error if the database fails
func ValidateLogin(username string, password string) (bool, error) {
	users, err := database.ListUsers()
	if err != nil {
		log.Error("Could not validate login due to database error: ", err.Error())
		return false, err
	}
	for _, user := range users {
		if user.Username == username && user.Password == password {
			return true, nil
		}
	}
	return false, nil
}
