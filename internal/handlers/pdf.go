package handlers

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/disintegration/imaging"
	"github.com/gen2brain/go-fitz"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

// PDFHandler handles PDF file thumbnails
type PDFHandler struct{}

// CanHandle returns true if this handler can process the given file type
func (h *PDFHandler) CanHandle(info *models.FileInfo) bool {
	return info.FileType == models.FileTypePDF
}

// Generate creates a thumbnail from a PDF file
func (h *PDFHandler) Generate(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Open PDF document
	doc, err := fitz.New(info.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	// Get page count
	pageCount := doc.NumPage()
	if pageCount == 0 {
		return nil, fmt.Errorf("PDF has no pages")
	}

	// Validate page number (1-indexed)
	pageNum := opts.Page - 1 // Convert to 0-indexed
	if pageNum < 0 {
		pageNum = 0
	}
	if pageNum >= pageCount {
		pageNum = pageCount - 1
	}

	// Render page to image
	// go-fitz renders at 72 DPI by default, we'll render at higher DPI for quality
	dpi := 150.0
	img, err := doc.ImageDPI(pageNum, dpi)
	if err != nil {
		// Fallback to default DPI
		img, err = doc.Image(pageNum)
		if err != nil {
			return nil, fmt.Errorf("failed to render PDF page: %w", err)
		}
	}

	// Convert to NRGBA for consistent processing
	nrgba := image.NewNRGBA(img.Bounds())
	draw.Draw(nrgba, nrgba.Bounds(), img, img.Bounds().Min, draw.Src)

	// Resize to fit dimensions
	resized := ResizeImage(nrgba, opts.Width, opts.Height)

	// Composite onto background
	result := compositeOnBackground(resized, opts.Background)

	return &models.ThumbnailResult{
		Image:    result,
		MimeType: "image/png",
		Width:    opts.Width,
		Height:   opts.Height,
	}, nil
}

// GetPageCount returns the number of pages in a PDF file
func GetPageCount(path string) (int, error) {
	doc, err := fitz.New(path)
	if err != nil {
		return 0, err
	}
	defer doc.Close()
	return doc.NumPage(), nil
}

// init registers the PDF format decoders
func init() {
	// go-fitz is registered via blank imports above
}

// ensure these are used
var _ = imaging.Resize
