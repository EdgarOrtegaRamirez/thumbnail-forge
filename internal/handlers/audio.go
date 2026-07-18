package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	"github.com/disintegration/imaging"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

// AudioHandler handles audio file thumbnails
type AudioHandler struct{}

// CanHandle returns true if this handler can process the given file type
func (h *AudioHandler) CanHandle(info *models.FileInfo) bool {
	return info.FileType == models.FileTypeAudio
}

// Generate creates a thumbnail from an audio file
func (h *AudioHandler) Generate(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Try to extract album art first
	img, err := extractAlbumArt(info.Path)
	if err == nil && img != nil {
		// Successfully extracted album art
		resized := ResizeImage(img, opts.Width, opts.Height)
		result := compositeOnBackground(resized, opts.Background)

		return &models.ThumbnailResult{
			Image:    result,
			MimeType: "image/png",
			Width:    opts.Width,
			Height:   opts.Height,
		}, nil
	}

	// Fall back to waveform visualization if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err == nil {
		return h.generateWaveform(info, opts)
	}

	// If neither album art nor ffmpeg is available, generate a placeholder
	return h.generatePlaceholder(info, opts)
}

// extractAlbumArt extracts album art from an audio file
func extractAlbumArt(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	// Get the tag metadata
	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}

	// Check if there's album art
	pic := m.Picture()
	if pic == nil {
		return nil, fmt.Errorf("no album art found")
	}

	// Decode the album art image
	img, _, err := image.Decode(bytes.NewReader(pic.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode album art: %w", err)
	}

	return img, nil
}

// generateWaveform generates a waveform visualization using ffmpeg
func (h *AudioHandler) generateWaveform(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Create temporary file for the waveform
	tmpDir, err := os.MkdirTemp("", "thumbnail-forge-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tmpWaveform := filepath.Join(tmpDir, "waveform.png")

	// Generate waveform using ffmpeg
	// showwavespic colors parameter specifies waveform channel colors
	// We use a bright color for the waveform and fill background separately
	cmd := exec.Command("ffmpeg",
		"-i", info.Path,
		"-filter_complex", fmt.Sprintf(
			"showwavespic=s=%dx%d:colors=0xcdd6f4|0xf5e0dc:split_channels=0",
			opts.Width, opts.Height,
		),
		"-frames:v", "1",
		"-update", "1",
		"-y",
		tmpWaveform,
	)

	// Capture stderr for error messages
	var stderr strings.Builder
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg waveform failed: %w\n%s", err, stderr.String())
	}

	// Load the waveform image
	img, err := LoadImage(tmpWaveform)
	if err != nil {
		return nil, fmt.Errorf("failed to load waveform: %w", err)
	}

	// Check if the waveform is empty (all transparent or all background)
	// This can happen with very short or silent audio
	if isEmptyWaveform(img) {
		return h.generatePlaceholder(info, opts)
	}

	// Resize to requested dimensions
	resized := ResizeImage(img, opts.Width, opts.Height)

	// Composite onto background (ffmpeg produces transparent PNG)
	result := compositeOnBackground(resized, opts.Background)

	return &models.ThumbnailResult{
		Image:    result,
		MimeType: "image/png",
		Width:    opts.Width,
		Height:   opts.Height,
	}, nil
}

// isEmptyWaveform checks if a waveform image is empty (all transparent pixels)
func isEmptyWaveform(img image.Image) bool {
	bounds := img.Bounds()
	nonTransparent := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 4 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 4 {
			if _, _, _, a := img.At(x, y).RGBA(); a > 0 {
				nonTransparent++
			}
		}
	}
	return nonTransparent == 0
}

// generatePlaceholder generates a placeholder image with a music note icon
func (h *AudioHandler) generatePlaceholder(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Create a simple placeholder image
	img := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{opts.Background}, image.Point{}, draw.Src)

	// Draw a simple music note icon (using basic shapes)
	// This is a simplified representation
	centerX := opts.Width / 2
	centerY := opts.Height / 2

	// Draw a circle for the note head
	radius := opts.Width / 8
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if x*x+y*y <= radius*radius {
				img.Set(centerX+x, centerY+y, color.RGBA{R: 200, G: 200, B: 200, A: 255})
			}
		}
	}

	// Draw a vertical line for the stem
	stemWidth := 2
	stemHeight := opts.Height / 3
	for y := centerY - radius; y >= centerY-radius-stemHeight; y-- {
		for x := -stemWidth / 2; x <= stemWidth/2; x++ {
			img.Set(centerX+radius+x, y, color.RGBA{R: 200, G: 200, B: 200, A: 255})
		}
	}

	return &models.ThumbnailResult{
		Image:    img,
		MimeType: "image/png",
		Width:    opts.Width,
		Height:   opts.Height,
	}, nil
}

// init registers the audio format decoders
func init() {
	// dhowden/tag is registered via blank imports above
}

// ensure these are used
var _ = imaging.Resize
var _ = draw.Draw
var _ = jpeg.Encode
var _ = png.Encode
