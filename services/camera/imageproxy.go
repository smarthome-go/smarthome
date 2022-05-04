package camera

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/image/webp"
)

// Todo: setup camera sql table(s)

// Fetches an image given an url, returns the image data as `[]byte`
// WIll return an error if connection or parsing problems occur
func fetchImageBytes(url string, timeout int) ([]byte, error) {
	start := time.Now()
	log.Trace(fmt.Sprintf("Initiating image fetching from: '%s'", url))
	client := http.Client{Timeout: time.Second * time.Duration(timeout)}
	response, err := client.Get(url)
	if err != nil {
		log.Error("Failed to fetch image through proxy: ", err.Error())
		return make([]byte, 0), err
	}
	if response.StatusCode != 200 {
		log.Error("Received non 200 response code\n")
	}
	imageData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return make([]byte, 0), err
	}
	imageMegabytes := len(imageData) / 1024 / 1024
	if imageMegabytes > int(maxImageSize) {
		log.Warn("Failed to fetch image with a size greater than allowed limit")
		return nil, errors.New("failed to fetch image: size to large")
	}
	log.Trace(fmt.Sprintf("Finished image fetching. (size: %d MB) in %v", imageMegabytes, time.Since(start)))
	return imageData, nil
}

// Converts the fetched bytes (any of the following formats) to PNG bytes, supported media types are `png, jpeg, and webp`
// [DEPRECATED] this function is temporarely depreacted due to extreme delays (* 1000) when converting jpeg to png
func convertBytesToPng(imageBytes []byte) ([]byte, error) {
	log.Trace("Converting image to PNG")
	contentType := http.DetectContentType(imageBytes)
	var intermediateImage image.Image
	switch contentType {
	case "image/png":
		log.Trace("Image is already in target format:(`image/png`)")
		return imageBytes, nil
	case "image/webp":
		log.Trace("Image is not in target format:(`image/webp`)")
		intermediateImageTemp, err := webp.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			log.Error("Could not decode image: ", err.Error())
			return nil, errors.Wrap(err, "unable to decode webp")
		}
		intermediateImage = intermediateImageTemp
	case "image/jpeg":
		log.Trace("Image is not in target format:(`image/jpeg`)")
		intermediateImageTemp, err := jpeg.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			log.Error("Could not decode image: ", err.Error())
			return nil, errors.Wrap(err, "unable to decode jpeg")
		}
		intermediateImage = intermediateImageTemp
	default:
		log.Error(fmt.Sprintf("Unable to convert `%#v` to `image/png`: incompatible media type: ", contentType))
		return nil, errors.New("Unable to convert to `image/png`: incompatible source media type")
	}
	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, intermediateImage); err != nil {
		log.Error("Could not decode image: ", err.Error())
		return nil, errors.Wrap(err, "unable to encode png")
	}
	log.Trace("Successfully converted image to PNG")
	return buffer.Bytes(), nil
}

// Validates an arbitrary set of bytes to match any of the formats below
// Supported formates: webp, png or jpeg/jpg
// Used to validate result when fetching a camera's video preview
func ensureValidFormat(data []byte) (valid bool) {
	supportedTypes := []string{"image/png", "image/webp", "image/jpeg"}
	for _, format := range supportedTypes {
		if format == http.DetectContentType(data) {
			return true
		}
	}
	return false
}
