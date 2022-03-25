package camera

func TestReturn() ([]byte, error) {
	url := "https://mik-mueller.de/assets/icon_1.png"
	byt, err := fetchImageBytes(url, 1)
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
