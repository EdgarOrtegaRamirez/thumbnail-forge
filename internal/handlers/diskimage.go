package handlers

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

// DiskImageHandler handles disk image file thumbnails (DMG, ISO, IMG)
type DiskImageHandler struct{}

// CanHandle returns true if this handler can process the given file type
func (h *DiskImageHandler) CanHandle(info *models.FileInfo) bool {
	return info.FileType == models.FileTypeDiskImage
}

// Generate creates a placeholder thumbnail for disk images
func (h *DiskImageHandler) Generate(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	img := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{opts.Background}, image.Point{}, draw.Src)

	// Draw a disk/drive icon
	cx := opts.Width / 2
	cy := opts.Height / 2
	r := opts.Width / 3
	if r > opts.Height/3 {
		r = opts.Height / 3
	}

	// Draw outer circle (disk platter)
	diskColor := color.RGBA{R: 180, G: 180, B: 190, A: 255}
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				img.Set(cx+x, cy+y, diskColor)
			}
		}
	}

	// Draw inner circle (spindle hole)
	holeR := r / 4
	holeColor := opts.Background
	for y := -holeR; y <= holeR; y++ {
		for x := -holeR; x <= holeR; x++ {
			if x*x+y*y <= holeR*holeR {
				img.Set(cx+x, cy+y, holeColor)
			}
		}
	}

	// Draw a label with the extension
	labelColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	labelY := cy + r + r/4
	if labelY >= opts.Height {
		labelY = opts.Height - 4
	}
	for x := cx - r/2; x < cx+r/2; x++ {
		if x >= 0 && x < opts.Width {
			img.Set(x, labelY, labelColor)
		}
	}

	return &models.ThumbnailResult{
		Image:    img,
		MimeType: "image/png",
		Width:    opts.Width,
		Height:   opts.Height,
	}, nil
}

// Ensure the handler satisfies the interface
var _ models.Handler = (*DiskImageHandler)(nil)

func init() {
	// Register disk image handler
	_ = fmt.Sprintf
}
