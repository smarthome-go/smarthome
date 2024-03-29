package camera

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Fetches an image given an url, returns the image data as `[]byte`
// WIil return an error if connection or parsing problems occur
func fetchImageBytes(imgURL string, timeout int) ([]byte, error) {
	start := time.Now()
	imgURLStruct, err := url.Parse(imgURL)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to fetch image from %s: Could not parse URL: %s", imgURL, err.Error()))
		return nil, err
	}
	switch imgURLStruct.Scheme {
	case "http", "https":
		break
	case "":
		log.Warn("Omitting the camera feed's request method is not recommended, using default 'http'")

		// If no protocol was provided, try prefixing the url with `http`
		imgURL = "http://" + imgURL

		// Parse the URL again in order to display the (maybe) current hostname
		imgURLStructTemp, err := url.Parse(imgURL)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to fetch image from %s: Could not parse URL: %s", imgURL, err.Error()))
			return nil, err
		}
		imgURLStruct = imgURLStructTemp
	default:
		return nil, fmt.Errorf("Unsupported protocol error: Protocol: '%s' can not be used to fetch images. Please use 'http' or 'https' instead.", imgURLStruct.Scheme)
	}
	log.Trace(fmt.Sprintf("Initiating image fetching from: '%s'", imgURLStruct.Host))
	client := http.Client{Timeout: time.Second * time.Duration(timeout)}
	response, err := client.Get(imgURL)
	if err != nil {
		log.Error("Failed to fetch image through proxy: ", err.Error())
		return nil, err
	}
	if response.StatusCode != 200 {
		log.Error("Received non 200 response code\n")
	}
	imageData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	imageMegabytes := len(imageData) / 1024 / 1024
	if imageMegabytes > int(maxImageSize) {
		log.Warn("Failed to fetch image with a size greater than allowed limit")
		return nil, fmt.Errorf("failed to fetch image: size to large (max: %d MB; current %d MB)", maxImageSize, imageMegabytes)
	}
	log.Trace(fmt.Sprintf("Finished image fetching. (size: %d MB) in %v", imageMegabytes, time.Since(start)))
	return imageData, nil
}

// [DEPRECATED] Uses a C library which causes programs for cross compilation
// func compressConvert(dataIn []byte) ([]byte, error) {
// 	fmt.Println("converting...")
// 	start := time.Now()
//
// 	// Convert to WEBP
// 	converted, err := bimg.NewImage(dataIn).Convert(bimg.WEBP)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Compress to lower quality
// 	compressed, err := bimg.NewImage(converted).Process(bimg.Options{Quality: int(imageQualityPercent)})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	fmt.Printf("TIME %v | bef: %d | af: %d\n", time.Since(start), len(dataIn), len(compressed))
//
// 	return compressed, nil
// }

// Validates an arbitrary set of bytes to match any of the formats below
// Supported formats: webp, png or jpeg/jpg
// Used to validate result when fetching a camera's video preview
func ensureValidFormat(data []byte) (valid bool) {
	supportedTypes := []string{"image/png", "image/webp", "image/jpeg"}
	contentType := http.DetectContentType(data)
	for _, format := range supportedTypes {
		if format == contentType {
			return true
		}
	}
	return false
}

// Converts the fetched bytes (any of the following formats) to PNG bytes, supported media types are `png, jpeg, and webp`
// [DEPRECATED] this function is temporarily deprecated due to extreme delays (initial * 1000) when converting jpeg to png

// func convertBytesToPng(imageBytes []byte) ([]byte, error) {
// 	log.Trace("Converting image to PNG")
// 	contentType := http.DetectContentType(imageBytes)
// 	var intermediateImage image.Image
// 	switch contentType {
// 	case "image/png":
// 		log.Trace("Image is already in target format:(`image/png`)")
// 		return imageBytes, nil
// 	case "image/webp":
// 		log.Trace("Image is not in target format:(`image/webp`)")
// 		intermediateImageTemp, err := webp.Decode(bytes.NewReader(imageBytes))
// 		if err != nil {
// 			log.Error("Could not decode image: ", err.Error())
// 			return nil, errors.Wrap(err, "unable to decode webp")
// 		}
// 		intermediateImage = intermediateImageTemp
// 	case "image/jpeg":
// 		log.Trace("Image is not in target format:(`image/jpeg`)")
// 		intermediateImageTemp, err := jpeg.Decode(bytes.NewReader(imageBytes))
// 		if err != nil {
// 			log.Error("Could not decode image: ", err.Error())
// 			return nil, errors.Wrap(err, "unable to decode jpeg")
// 		}
// 		intermediateImage = intermediateImageTemp
// 	default:
// 		log.Error(fmt.Sprintf("Unable to convert `%#v` to `image/png`: incompatible media type: ", contentType))
// 		return nil, errors.New("Unable to convert to `image/png`: incompatible source media type")
// 	}
// 	buffer := new(bytes.Buffer)
// 	if err := png.Encode(buffer, intermediateImage); err != nil {
// 		log.Error("Could not decode image: ", err.Error())
// 		return nil, errors.Wrap(err, "unable to encode png")
// 	}
// 	log.Trace("Successfully converted image to PNG")
// 	return buffer.Bytes(), nil
// }
