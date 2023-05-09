package camera

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
)

// Can be adjusted to define a maximum image size
const maxImageSize uint8 = 10 // Size is in megabytes
const imageQualityPercent uint8 = 25

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// Returns the current image (from feed) of the given camera
// Returns an error if fetching or encoding data fails
// Uses the arguments (camera-id and fetch timeout in seconds)
func GetCameraFeed(id string, timeoutSecs int) (data []byte, err error) {
	camera, found, err := database.GetCameraById(id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("no such camera exists")
	}
	byteData, err := fetchImageBytes(camera.Url, timeoutSecs)
	if err != nil {
		log.Error("Failed to fetch camera feed: ", err.Error())
		return nil, err
	}
	// [DEPRECATED] for the same reason the convertBytesToPng function is deprecated
	// img, err := convertBytesToPng(byteData)
	// if err != nil {
	// 	log.Error("Failed to fetch camera feed: could not convert bytes to image: ", err.Error())
	// 	return nil, err
	// }
	// return img, nil

	// [INSTEAD], the fetched data is just validated to match modern browser's requirements
	if !ensureValidFormat(byteData) {
		log.Warn("invalid content-type of fetched bytes: not a supported image type")
		return nil, errors.New("content-type of fetched bytes not supported")
	}

	// [DEPRECATED] Uses a C library which causes programs for cross compilation
	// Convert to WEBP & compress image
	// return compressConvert(byteData)

	return byteData, nil
}
