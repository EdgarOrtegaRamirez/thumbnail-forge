package handlers

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

// ArchiveHandler handles archive file thumbnails
type ArchiveHandler struct{}

// CanHandle returns true if this handler can process the given file type
func (h *ArchiveHandler) CanHandle(info *models.FileInfo) bool {
	return info.FileType == models.FileTypeArchive
}

// Generate creates a thumbnail from an archive file
func (h *ArchiveHandler) Generate(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Get archive contents
	contents, err := getArchiveContents(info.Path)
	if err != nil {
		// Fall back to placeholder
		return h.generatePlaceholder(info, opts)
	}

	// Generate a thumbnail showing the archive contents
	return h.generateFromContents(contents, info, opts)
}

// getArchiveContents gets the list of files in an archive
func getArchiveContents(path string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".zip":
		return getZipContents(path)
	case ".tar":
		return getTarContents(path)
	case ".gz":
		return getTarGzContents(path)
	case ".tgz":
		return getTarGzContents(path)
	default:
		return nil, fmt.Errorf("unsupported archive format: %s", ext)
	}
}

// getZipContents gets the contents of a ZIP file
func getZipContents(path string) ([]string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	var contents []string
	for _, f := range r.File {
		contents = append(contents, f.Name)
	}

	return contents, nil
}

// getTarContents gets the contents of a TAR file
func getTarContents(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	tr := tar.NewReader(f)
	var contents []string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		contents = append(contents, header.Name)
	}

	return contents, nil
}

// getTarGzContents gets the contents of a TAR.GZ file
func getTarGzContents(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	var contents []string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		contents = append(contents, header.Name)
	}

	return contents, nil
}

// generateFromContents generates a thumbnail from archive contents
func (h *ArchiveHandler) generateFromContents(contents []string, info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Create a simple placeholder image
	img := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{opts.Background}, image.Point{}, draw.Src)

	// Draw a simple archive icon (using basic shapes)
	centerX := opts.Width / 2
	centerY := opts.Height / 2

	// Draw a rectangle for the archive
	archiveWidth := opts.Width / 2
	archiveHeight := opts.Height / 3
	archiveX := centerX - archiveWidth/2
	archiveY := centerY - archiveHeight/2

	// Fill archive area
	for y := archiveY; y < archiveY+archiveHeight; y++ {
		for x := archiveX; x < archiveX+archiveWidth; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 200, B: 200, A: 255})
		}
	}

	// Draw a few horizontal lines to represent files
	lineColor := color.RGBA{R: 150, G: 150, B: 150, A: 255}
	lineHeight := 2
	lineSpacing := 8

	maxLines := len(contents)
	if maxLines > 5 {
		maxLines = 5
	}

	for i := 0; i < maxLines; i++ {
		lineY := archiveY + 10 + i*lineSpacing
		if lineY+lineHeight >= archiveY+archiveHeight-5 {
			break
		}

		lineWidth := archiveWidth - 20

		for y := lineY; y < lineY+lineHeight; y++ {
			for x := archiveX + 10; x < archiveX+10+lineWidth; x++ {
				img.Set(x, y, lineColor)
			}
		}
	}

	return &models.ThumbnailResult{
		Image:    img,
		MimeType: "image/png",
		Width:    opts.Width,
		Height:   opts.Height,
	}, nil
}

// generatePlaceholder generates a placeholder image
func (h *ArchiveHandler) generatePlaceholder(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Create a simple placeholder image
	img := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{opts.Background}, image.Point{}, draw.Src)

	// Draw a simple archive icon (using basic shapes)
	centerX := opts.Width / 2
	centerY := opts.Height / 2

	// Draw a rectangle for the archive
	archiveWidth := opts.Width / 2
	archiveHeight := opts.Height / 3
	archiveX := centerX - archiveWidth/2
	archiveY := centerY - archiveHeight/2

	// Fill archive area
	for y := archiveY; y < archiveY+archiveHeight; y++ {
		for x := archiveX; x < archiveX+archiveWidth; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 200, B: 200, A: 255})
		}
	}

	return &models.ThumbnailResult{
		Image:    img,
		MimeType: "image/png",
		Width:    opts.Width,
		Height:   opts.Height,
	}, nil
}

// init registers the archive format decoders
func init() {
	// Archive handling is done via Go's standard library
}

// ensure these are used
var _ = imaging.Resize
var _ = draw.Draw
