package terminal

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

// Protocol represents a terminal image display protocol
type Protocol int

const (
	// ProtocolAuto detects the best available protocol
	ProtocolAuto Protocol = iota
	// ProtocolKitty uses the Kitty terminal graphics protocol
	ProtocolKitty
	// ProtocolITerm2 uses the iTerm2 inline image protocol
	ProtocolITerm2
	// ProtocolSixel uses the Sixel graphics format
	ProtocolSixel
	// ProtocolUnicode uses Unicode block characters for a fallback
	ProtocolUnicode
)

// Output handles terminal image output
type Output struct {
	protocol Protocol
}

// NewOutput creates a new terminal output handler
func NewOutput(protocol Protocol) *Output {
	return &Output{protocol: protocol}
}

// DetectProtocol detects the best available terminal protocol
func DetectProtocol() Protocol {
	// Check for Kitty
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return ProtocolKitty
	}

	// Check for iTerm2
	if os.Getenv("TERM_PROGRAM") == "iTerm.app" {
		return ProtocolITerm2
	}

	// Check for WezTerm (supports iTerm2 protocol)
	if os.Getenv("TERM_PROGRAM") == "WezTerm" {
		return ProtocolITerm2
	}

	// Check for Sixel support via environment variable
	if os.Getenv("SIXEL_SUPPORT") == "1" {
		return ProtocolSixel
	}

	// Default to Unicode fallback
	return ProtocolUnicode
}

// Display displays an image in the terminal using the specified protocol
func (o *Output) Display(img image.Image, result *models.ThumbnailResult) error {
	if img == nil {
		return fmt.Errorf("image is nil")
	}

	// Use detected protocol if auto
	protocol := o.protocol
	if protocol == ProtocolAuto {
		protocol = DetectProtocol()
	}

	switch protocol {
	case ProtocolKitty:
		return o.displayKitty(img)
	case ProtocolITerm2:
		return o.displayITerm2(img)
	case ProtocolSixel:
		return o.displaySixel(img)
	case ProtocolUnicode:
		return o.displayUnicode(img)
	default:
		return fmt.Errorf("unsupported protocol: %d", protocol)
	}
}

// displayKitty displays an image using the Kitty terminal graphics protocol
func (o *Output) displayKitty(img image.Image) error {
	// Encode image to PNG
	var buf []byte
	if err := encodePNG(img, &buf); err != nil {
		return err
	}

	// Kitty protocol uses base64 encoding
	encoded := encodeBase64(buf)

	// Split into chunks (Kitty has a max chunk size)
	chunkSize := 4096
	chunks := splitString(encoded, chunkSize)

	// Send the image data
	for i, chunk := range chunks {
		control := 0
		if i == 0 {
			control = 1 // Start of transmission
		}
		if i == len(chunks)-1 {
			control = 2 // End of transmission
		}
		if i > 0 && i < len(chunks)-1 {
			control = 0 // Middle
		}

		fmt.Printf("\033_Gf=100,t=d,s=%d,v=%d,a=T,c=%d;%s\033\\",
			img.Bounds().Dx(), img.Bounds().Dy(), control, chunk)
	}

	fmt.Println()
	return nil
}

// displayITerm2 displays an image using the iTerm2 inline image protocol
func (o *Output) displayITerm2(img image.Image) error {
	// Encode image to PNG
	var buf []byte
	if err := encodePNG(img, &buf); err != nil {
		return err
	}

	// iTerm2 protocol uses base64 encoding
	encoded := encodeBase64(buf)

	// Send the image data
	fmt.Printf("\033]1337;File=inline=1;size=%d:%s\a", len(buf), encoded)
	fmt.Println()
	return nil
}

// displaySixel displays an image using the Sixel graphics format
func (o *Output) displaySixel(img image.Image) error {
	// Sixel encoding is complex, so we'll use a simplified approach
	// In practice, you'd use a library like github.com/mattn/go-sixel

	// For now, fall back to Unicode
	return o.displayUnicode(img)
}

// displayUnicode displays an image using Unicode block characters
func (o *Output) displayUnicode(img image.Image) error {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Scale down to terminal-friendly size
	maxWidth := 80
	maxHeight := 24

	scaleX := float64(maxWidth) / float64(width)
	scaleY := float64(maxHeight) / float64(height)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}

	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale / 2) // Divide by 2 because characters are taller than wide

	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	// Create a new image with the scaled dimensions
	scaledImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Scale the image
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := bounds.Min.X + int(float64(x)/scale)
			srcY := bounds.Min.Y + int(float64(y)*2/scale)

			if srcX >= bounds.Max.X {
				srcX = bounds.Max.X - 1
			}
			if srcY >= bounds.Max.Y {
				srcY = bounds.Max.Y - 1
			}

			c := img.At(srcX, srcY)
			scaledImg.Set(x, y, c)
		}
	}

	// Display using Unicode block characters
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Get the color of the pixel
			c := scaledImg.RGBAAt(x, y)

			// Convert to ANSI 256-color
			r := int(c.R) * 5 / 255
			g := int(c.G) * 5 / 255
			b := int(c.B) * 5 / 255
			colorIndex := 16 + 36*r + 6*g + b

			// Use block character with color
			fmt.Printf("\033[38;5;%dm█\033[0m", colorIndex)
		}
		fmt.Println()
	}

	return nil
}

// encodePNG encodes an image to PNG format
func encodePNG(img image.Image, buf *[]byte) error {
	// Create a buffer
	f, err := os.CreateTemp("", "thumbnail-*.png")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	defer func() { _ = f.Close() }()

	// Encode the image
	if err := png.Encode(f, img); err != nil {
		return err
	}

	// Read the file back
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	*buf, err = os.ReadFile(f.Name())
	return err
}

// encodeBase64 encodes bytes to base64 string
func encodeBase64(data []byte) string {
	// Simple base64 encoding without external dependencies
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	var result []byte
	for i := 0; i < len(data); i += 3 {
		b0 := data[i]
		var b1, b2 byte
		if i+1 < len(data) {
			b1 = data[i+1]
		}
		if i+2 < len(data) {
			b2 = data[i+2]
		}

		result = append(result, chars[b0>>2])
		result = append(result, chars[((b0&0x3)<<4)|(b1>>4)])
		if i+1 < len(data) {
			result = append(result, chars[((b1&0xF)<<2)|(b2>>6)])
		} else {
			result = append(result, '=')
		}
		if i+2 < len(data) {
			result = append(result, chars[b2&0x3F])
		} else {
			result = append(result, '=')
		}
	}

	return string(result)
}

// splitString splits a string into chunks of the specified size
func splitString(s string, size int) []string {
	var chunks []string
	for len(s) > 0 {
		end := size
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[:end])
		s = s[end:]
	}
	return chunks
}

// GeneratePlaceholder generates a simple placeholder image
func GeneratePlaceholder(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a gradient background
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create a simple gradient
			r := uint8(float64(x) / float64(width) * 255)
			g := uint8(float64(y) / float64(height) * 255)
			b := uint8(128)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	return img
}

// init ensures the package is used
func init() {
	// Ensure these imports are used
	_ = draw.Draw
	_ = strings.TrimSpace
}
