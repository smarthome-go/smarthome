package camera

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func TestImageProxy() {
	url := "https://mik-mueller.de/assets/Untitled.png"
	byt, _ := fetchImageBytes(url)
	img, err := convertBytesToPng(byt)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := ioutil.WriteFile("image.png", img, 0664); err != nil {
		log.Error("Failed to write test image to disk: ", err.Error())
	}
}

func TestReturn() ([]byte, error) {
	url := "https://mik-mueller.de/assets/Untitled.png"
	byt, _ := fetchImageBytes(url)
	img, err := convertBytesToPng(byt)
	if err != nil {
		log.Error(err.Error())
		return nil, nil
	}
	return img, nil
}
