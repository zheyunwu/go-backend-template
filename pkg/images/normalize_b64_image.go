package images

import (
	"encoding/base64"
	"strings"
)

// NormalizeBase64Image 校验并规范化base64编码的图片数据
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
