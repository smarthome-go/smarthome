package camera

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// ADD ERROR HANDLING
func TestImageProxy() {
	url := "https://mik-mueller.de/assets/Untitled.png"
	byt, _ := fetchImageBytes(url)
	img, err := convertBytesToPng(byt)
	if err != nil {
		log.Error(err.Error())
		return
	}
	ioutil.WriteFile("image.png", img, 0664)
}
