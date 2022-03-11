package camera

import "io/ioutil"

func TestImageProxy() {
	url := "https://mik-mueller.de/assets/icon_1.png"
	byt, _ := fetchImageBytes(url)
	img, err := convertBytesToPng(byt)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err := ioutil.WriteFile("112312312312312.png", img, 0664); err != nil {
		log.Error("Failed to write test image to disk: ", err.Error())
	}
}

func TestReturn() ([]byte, error) {
	url := "https://mik-mueller.de/assets/icon_1.png"
	byt, err := fetchImageBytes(url)
	if err != nil {
		return nil, err
	}
	img, err := convertBytesToPng(byt)
	if err != nil {
		log.Error(err.Error())
		return nil, nil
	}
	return img, nil
}
