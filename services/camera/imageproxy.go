package camera

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// THIS FILE SHOULD FETCH AN IMAGE FROM A GIVEN URL IN ORDER TO RETURN A BYTE ARRAY AND AN IMAGE

// Todo: setup camera sql table(s)

// Fetches an image given an url, returns the image data as `[]byte`
// WIll return an error if connection or parsing problems occur
func fetchImageBytes(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		log.Error("Failed to fetch image through proxy: ", err.Error())
		return make([]byte, 0), err
	}
	if response.StatusCode != 200 {
		fmt.Printf("Received non 200 response code\n")
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return make([]byte, 0), err
	}
	return body, nil
}

// Converts the fetched bytes to a PNG image
// TODO: test and modify the code
func convertBytesToPng(imageBytes []byte) ([]byte, error) {
	contentType := http.DetectContentType(imageBytes)
	switch contentType {
	case "image/png":
		return imageBytes, nil
	case "image/jpeg":
		img, err := jpeg.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			return nil, errors.Wrap(err, "unable to decode jpeg")
		}
		buf := new(bytes.Buffer)
		if err := png.Encode(buf, img); err != nil {
			return nil, errors.Wrap(err, "unable to encode png")
		}

		return buf.Bytes(), nil
	}

	return nil, fmt.Errorf("unable to convert %#v to png", contentType)
}
