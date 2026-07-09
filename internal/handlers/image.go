package handlers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

// ImageHandler handles image file thumbnails
type ImageHandler struct{}

// CanHandle returns true if this handler can process the given file type
func (h *ImageHandler) CanHandle(info *models.FileInfo) bool {
	return info.FileType == models.FileTypeImage
}

// Generate creates a thumbnail from an image file
func (h *ImageHandler) Generate(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Open and decode the image
	img, err := LoadImage(info.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Resize the image
	resized := ResizeImage(img, opts.Width, opts.Height)

	// Composite onto background if needed (handles transparency)
	result := compositeOnBackground(resized, opts.Background)

	return &models.ThumbnailResult{
		Image:    result,
		MimeType: getOutputMimeType(opts.Format),
		Width:    opts.Width,
		Height:   opts.Height,
	}, nil
}

// LoadImage loads an image from the specified path.
// For formats Go's stdlib can't decode (HEIC/HEIF/AVIF/ICNS), it shells out
// to external tools or uses pure-Go parsers.
func LoadImage(path string) (image.Image, error) {
	ext := strings.ToLower(filepath.Ext(path))

	// Try Go's built-in decoders first (PNG, JPEG, GIF, WebP, BMP, TIFF)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(f)
	f.Close()
	if err == nil {
		return img, nil
	}

	// Fall back to format-specific handlers for Apple formats
	switch ext {
	case ".heic", ".heif":
		return loadHEIF(path)
	case ".avif":
		return loadAVIF(path)
	case ".icns":
		return loadICNS(path)
	}

	return nil, err
}

// loadHEIF decodes HEIC/HEIF images using heif-convert (libheif-examples)
func loadHEIF(path string) (image.Image, error) {
	tmpDir, err := os.MkdirTemp("", "thumbforge-heif-*")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	outPath := filepath.Join(tmpDir, "output.png")

	cmd := exec.Command("heif-convert", path, outPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("heif-convert failed: %w: %s", err, stderr.String())
	}

	// heif-convert may name the file differently
	entries, _ := os.ReadDir(tmpDir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".png") {
			outPath = filepath.Join(tmpDir, e.Name())
			break
		}
	}

	return loadPNG(outPath)
}

// loadAVIF decodes AVIF images using heif-convert (libheif supports AVIF)
func loadAVIF(path string) (image.Image, error) {
	return loadHEIF(path) // same tool handles AVIF
}

// loadICNS decodes Apple ICNS icon files using a pure Go parser.
// ICNS files contain embedded PNG or JPEG data in icon entries.
func loadICNS(path string) (image.Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Verify ICNS magic
	if len(data) < 8 || string(data[0:4]) != "icns" {
		return nil, fmt.Errorf("invalid ICNS file: missing 'icns' magic")
	}

	// Parse icon entries. Header: 4 bytes magic + 4 bytes file size.
	// Each entry: 4 bytes type + 4 bytes entry size + data.
	offset := 8
	for offset+8 <= len(data) {
		entryType := string(data[offset : offset+4])
		entrySize := binary.BigEndian.Uint32(data[offset+4 : offset+8])

		if offset+int(entrySize) > len(data) {
			break
		}

		entryData := data[offset+8 : offset+int(entrySize)]

		// Try to decode the icon data as PNG or JPEG
		// ICNS types: ic07=128x128 PNG, ic08=256x256 PNG, ic09=512x512 PNG,
		// ic10=1024x1024 PNG, ic11=32x32 PNG, ic12=64x64 PNG, ic13=256x256 JPEG2000,
		// ic14=512x512 JPEG2000, ic04=16x16, ic05=32x32
		img, _, err := image.Decode(bytes.NewReader(entryData))
		if err == nil {
			return img, nil
		}

		// Try JPEG2000 via image.Decode fallback (some ICNS use raw JPEG)
		if len(entryData) > 2 && entryData[0] == 0xFF && entryData[1] == 0xD8 {
			img, _, err := image.Decode(bytes.NewReader(entryData))
			if err == nil {
				return img, nil
			}
		}

		// Try ARGB raw pixel data (some older ICNS types)
		// Types like ih32 (16x16), il32 (32x32), is32 (16x16) use raw ARGB
		switch entryType {
		case "ih32", "il32", "is32":
			// These are masked pixel data, skip for now
		}

		offset += int(entrySize)
		// Align to 4 bytes
		if offset%4 != 0 {
			offset += 4 - offset%4
		}
	}

	return nil, fmt.Errorf("no decodable image data found in ICNS file")
}

// loadPNG loads a PNG file directly
func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

// ResizeImage resizes an image to fit within the specified dimensions while maintaining aspect ratio
func ResizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	// Get original dimensions
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Calculate scale factor to fit within bounds
	scaleX := float64(maxWidth) / float64(origWidth)
	scaleY := float64(maxHeight) / float64(origHeight)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Don't upscale
	if scale > 1.0 {
		scale = 1.0
	}

	// Calculate new dimensions
	newWidth := int(float64(origWidth) * scale)
	newHeight := int(float64(origHeight) * scale)

	// Resize using Lanczos filter for high quality
	return imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
}

// compositeOnBackground composites an image onto a background color
func compositeOnBackground(img image.Image, bgColor color.Color) image.Image {
	bounds := img.Bounds()

	// Check if image has alpha channel
	if _, hasAlpha := img.(*image.NRGBA); !hasAlpha {
		if _, hasAlpha := img.(*image.RGBA); !hasAlpha {
			// No alpha channel, return as-is
			return img
		}
	}

	// Create background image
	bg := image.NewRGBA(bounds)
	draw.Draw(bg, bounds, &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Composite source onto background
	draw.Draw(bg, bounds, img, bounds.Min, draw.Over)

	return bg
}

// getOutputMimeType returns the MIME type for the output format
func getOutputMimeType(format string) string {
	switch format {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

// init registers the image format decoders
func init() {
	// Register additional format decoders if needed
	// WebP is registered via blank import above
}
