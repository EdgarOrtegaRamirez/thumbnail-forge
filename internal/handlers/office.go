package handlers

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

// OfficeHandler handles office document thumbnails
type OfficeHandler struct{}

// CanHandle returns true if this handler can process the given file type
func (h *OfficeHandler) CanHandle(info *models.FileInfo) bool {
	return info.FileType == models.FileTypeOffice
}

// Generate creates a thumbnail from an office document
func (h *OfficeHandler) Generate(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Check if LibreOffice is available
	if err := checkLibreOffice(); err != nil {
		// Fall back to placeholder
		return h.generatePlaceholder(info, opts)
	}

	// Try to convert the document to PDF first, then render
	tmpDir, err := os.MkdirTemp("", "thumbnail-forge-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Convert to PDF using LibreOffice
	pdfPath, err := convertToPDF(info.Path, tmpDir)
	if err != nil {
		// Fall back to placeholder
		return h.generatePlaceholder(info, opts)
	}

	// Use PDF handler to render the PDF
	pdfHandler := &PDFHandler{}
	pdfInfo := &models.FileInfo{
		Path:     pdfPath,
		FileType: models.FileTypePDF,
		MimeType: "application/pdf",
	}

	return pdfHandler.Generate(pdfInfo, opts)
}

// checkLibreOffice checks if LibreOffice is installed
func checkLibreOffice() error {
	// Check for libreoffice
	if _, err := exec.LookPath("libreoffice"); err == nil {
		return nil
	}

	// Check for soffice (LibreOffice's other binary name)
	if _, err := exec.LookPath("soffice"); err == nil {
		return nil
	}

	return fmt.Errorf("LibreOffice not found in PATH")
}

// convertToPDF converts an office document to PDF using LibreOffice
func convertToPDF(inputPath, outputDir string) (string, error) {
	// Determine the command to use
	cmd := "libreoffice"
	if _, err := exec.LookPath("libreoffice"); err != nil {
		cmd = "soffice"
	}

	// Run LibreOffice to convert to PDF
	args := []string{
		"--headless",
		"--convert-to", "pdf",
		"--outdir", outputDir,
		inputPath,
	}

	execCmd := exec.Command(cmd, args...)
	var stderr strings.Builder
	execCmd.Stderr = &stderr

	if err := execCmd.Run(); err != nil {
		return "", fmt.Errorf("LibreOffice conversion failed: %w\n%s", err, stderr.String())
	}

	// Find the output PDF
	baseName := filepath.Base(inputPath)
	pdfName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".pdf"
	pdfPath := filepath.Join(outputDir, pdfName)

	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return "", fmt.Errorf("PDF not created at %s", pdfPath)
	}

	return pdfPath, nil
}

// generatePlaceholder generates a placeholder image with a document icon
func (h *OfficeHandler) generatePlaceholder(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Create a simple placeholder image
	img := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{opts.Background}, image.Point{}, draw.Src)

	// Draw a simple document icon (using basic shapes)
	centerX := opts.Width / 2
	centerY := opts.Height / 2

	// Draw a rectangle for the document
	docWidth := opts.Width / 2
	docHeight := opts.Height / 3
	docX := centerX - docWidth/2
	docY := centerY - docHeight/2

	// Fill document area
	for y := docY; y < docY+docHeight; y++ {
		for x := docX; x < docX+docWidth; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 200, B: 200, A: 255})
		}
	}

	// Draw a few horizontal lines to represent text
	lineColor := color.RGBA{R: 150, G: 150, B: 150, A: 255}
	lineHeight := 2
	lineSpacing := 8

	for i := 0; i < 4; i++ {
		lineY := docY + 10 + i*lineSpacing
		if lineY+lineHeight >= docY+docHeight-5 {
			break
		}

		lineWidth := docWidth - 20
		if i == 3 {
			lineWidth = docWidth / 2 // Last line shorter
		}

		for y := lineY; y < lineY+lineHeight; y++ {
			for x := docX + 10; x < docX+10+lineWidth; x++ {
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

// init registers the office format decoders
func init() {
	// LibreOffice is used via exec, no Go imports needed
}

// ensure these are used
var _ = imaging.Resize
var _ = draw.Draw
