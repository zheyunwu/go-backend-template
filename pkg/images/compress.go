package images

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"strings"

	// Register decoders for image formats
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

const (
	// DefaultMaxWidth is the default maximum width for compressed images
	DefaultMaxWidth = 1200
	// DefaultJPEGQuality is the quality setting for JPEG compression (0-100)
	DefaultJPEGQuality = 85
	// DefaultMaxFileSize is the threshold file size for compression (1MB)
	DefaultMaxFileSize = 1 * 1024 * 1024
)

// CompressOptions contains options for image compression
type CompressOptions struct {
	MaxWidth    int  // Maximum width in pixels
	JPEGQuality int  // JPEG quality (0-100)
	ForceJPEG   bool // Force conversion to JPEG
}

// DefaultCompressOptions returns the default compression options
func DefaultCompressOptions() *CompressOptions {
	return &CompressOptions{
		MaxWidth:    DefaultMaxWidth,
		JPEGQuality: DefaultJPEGQuality,
		ForceJPEG:   false,
	}
}

// CompressImage compresses an image using default options
func CompressImage(imageData []byte) ([]byte, string, error) {
	return CompressImageWithOptions(imageData, DefaultCompressOptions())
}

// CompressImageWithOptions compresses an image with custom options
func CompressImageWithOptions(imageData []byte, options *CompressOptions) ([]byte, string, error) {
	// If image is smaller than 1MB and no force option, return as is
	if len(imageData) < DefaultMaxFileSize && !options.ForceJPEG {
		// Still need to determine the format
		format, err := detectImageFormat(imageData)
		if err != nil {
			return nil, "", err
		}
		return imageData, format, nil
	}

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Get the original dimensions
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// If image is already smaller than max width and not forcing JPEG, return as is
	if origWidth <= options.MaxWidth && !options.ForceJPEG {
		return imageData, format, nil
	}

	// Calculate new dimensions while preserving aspect ratio
	var newWidth, newHeight int
	if origWidth > options.MaxWidth {
		newWidth = options.MaxWidth
		newHeight = int(float64(origHeight) * float64(newWidth) / float64(origWidth))
	} else {
		newWidth = origWidth
		newHeight = origHeight
	}

	// Create a new RGBA image
	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Use bilinear scaling to resize
	bilinearScale(img, newImg, origWidth, origHeight, newWidth, newHeight)

	// Buffer to store the encoded image
	var buf bytes.Buffer

	// Determine output format (convert to JPEG if forced or original is not PNG/JPEG)
	outFormat := format
	if options.ForceJPEG || (format != "jpeg" && format != "png") {
		outFormat = "jpeg"
	}

	// Encode in the chosen format
	switch outFormat {
	case "jpeg":
		err = jpeg.Encode(&buf, newImg, &jpeg.Options{Quality: options.JPEGQuality})
	case "png":
		err = png.Encode(&buf, newImg)
	default:
		// Fallback to JPEG if format is unknown
		err = jpeg.Encode(&buf, newImg, &jpeg.Options{Quality: options.JPEGQuality})
		outFormat = "jpeg"
	}

	if err != nil {
		return nil, "", fmt.Errorf("failed to encode compressed image: %w", err)
	}

	return buf.Bytes(), outFormat, nil
}

// IsCompressible checks if the given data appears to be a compressible image
func IsCompressible(data []byte) bool {
	format, err := detectImageFormat(data)
	return err == nil && (format == "jpeg" || format == "png" || format == "gif")
}

// detectImageFormat detects the format of an image from its bytes
func detectImageFormat(data []byte) (string, error) {
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to detect image format: %w", err)
	}
	return format, nil
}

// FormatBytes returns a human-readable string representing bytes
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetContentType returns the content type based on the format
func GetContentType(format string) string {
	format = strings.ToLower(format)
	switch format {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

// bilinearScale implements a simple bilinear scaling algorithm
func bilinearScale(src image.Image, dst draw.Image, srcWidth, srcHeight, dstWidth, dstHeight int) {
	// For each pixel in the destination image
	for y := 0; y < dstHeight; y++ {
		for x := 0; x < dstWidth; x++ {
			// Calculate the corresponding position in the source image
			srcX := float64(x) * float64(srcWidth) / float64(dstWidth)
			srcY := float64(y) * float64(srcHeight) / float64(dstHeight)

			// Calculate the four nearest pixels in the source image
			x0, y0 := int(math.Floor(srcX)), int(math.Floor(srcY))
			x1, y1 := min(x0+1, srcWidth-1), min(y0+1, srcHeight-1)

			// Calculate the interpolation weights
			wx := srcX - float64(x0)
			wy := srcY - float64(y0)

			// Get the four nearest pixels
			c00 := src.At(x0, y0)
			c01 := src.At(x0, y1)
			c10 := src.At(x1, y0)
			c11 := src.At(x1, y1)

			// Convert to RGBA
			r00, g00, b00, a00 := c00.RGBA()
			r01, g01, b01, a01 := c01.RGBA()
			r10, g10, b10, a10 := c10.RGBA()
			r11, g11, b11, a11 := c11.RGBA()

			// Bilinear interpolation
			r := bilinearInterp(r00, r01, r10, r11, wx, wy)
			g := bilinearInterp(g00, g01, g10, g11, wx, wy)
			b := bilinearInterp(b00, b01, b10, b11, wx, wy)
			a := bilinearInterp(a00, a01, a10, a11, wx, wy)

			// Set the pixel in the destination image
			dst.Set(x, y, color.RGBA{r, g, b, a})
		}
	}
}

// bilinearInterp performs bilinear interpolation on a single channel
func bilinearInterp(c00, c01, c10, c11 uint32, wx, wy float64) uint8 {
	// Interpolate along the x-axis
	c0 := float64(c00)*(1-wx) + float64(c10)*wx
	c1 := float64(c01)*(1-wx) + float64(c11)*wx

	// Interpolate along the y-axis
	c := c0*(1-wy) + c1*wy

	// Convert back to uint8 (0-255)
	return uint8(c)
}
