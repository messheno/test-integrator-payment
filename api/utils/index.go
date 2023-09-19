package utils

import (
	"mime/multipart"
	"net/http"
	"time"
)

func GetFileContentType(ouput multipart.File) (string, error) {
	// to sniff the content type only the first
	// 512 bytes are used.

	buf := make([]byte, 512)

	_, err := ouput.Read(buf)

	if err != nil {
		return "", err
	}

	// the function that actually does the trick
	contentType := http.DetectContentType(buf)

	return contentType, nil
}

func StringToDate(dateString string, layouts []string) (bool, time.Time) {
	for _, layout := range layouts {
		t, err := time.Parse(layout, dateString)
		if err == nil {
			return true, t
		}
	}

	return false, time.Time{}
}
