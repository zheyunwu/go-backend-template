package images

import (
	"encoding/base64"
	"strings"
)

// NormalizeBase64Image validates and normalizes base64 encoded image data.
// It handles both raw base64 strings and data URIs (e.g., "data:image/png;base64,...").
func NormalizeBase64Image(imageData string) (string, error) {
	// Handle data URI format
	if strings.HasPrefix(imageData, "data:image") {
		parts := strings.Split(imageData, ",")
		if len(parts) == 2 {
			imageData = parts[1]
		}
	}

	// Validate base64 data
	_, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return "", err
	}

	return imageData, nil
}
