package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

// CodeHandler handles code and text file thumbnails
type CodeHandler struct{}

// CanHandle returns true if this handler can process the given file type
func (h *CodeHandler) CanHandle(info *models.FileInfo) bool {
	return info.FileType == models.FileTypeCode ||
		info.FileType == models.FileTypeText ||
		info.FileType == models.FileTypeMarkdown
}

// Generate creates a thumbnail from a code or text file
func (h *CodeHandler) Generate(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Read file content
	content, err := os.ReadFile(info.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// For markdown, try to render with syntax highlighting
	if info.FileType == models.FileTypeMarkdown {
		// Markdown is plain text with formatting, render as text
		return h.renderText(string(content), info, opts)
	}

	// For code files, try syntax highlighting
	if info.FileType == models.FileTypeCode {
		return h.renderCode(string(content), info, opts)
	}

	// For text files, render as plain text
	return h.renderText(string(content), info, opts)
}

// renderCode renders code with syntax highlighting
func (h *CodeHandler) renderCode(content string, info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Detect language from extension
	lexer := lexers.Match(info.Path)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Get the style
	style := styles.Get(opts.Theme)
	if style == nil {
		style = styles.Fallback
	}

	// Format as HTML
	formatter := html.New()
	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		// Fall back to plain text rendering
		return h.renderText(content, info, opts)
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		// Fall back to plain text rendering
		return h.renderText(content, info, opts)
	}

	// For now, render as plain text with line numbers
	// HTML rendering would require a headless browser
	return h.renderTextWithLineNumbers(content, info, opts)
}

// renderText renders plain text to an image
func (h *CodeHandler) renderText(content string, info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	return h.renderTextWithLineNumbers(content, info, opts)
}

// renderTextWithLineNumbers renders text with line numbers
func (h *CodeHandler) renderTextWithLineNumbers(content string, info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Split content into lines
	lines := strings.Split(content, "\n")

	// Limit to first 20 lines for thumbnail
	maxLines := 20
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	// Calculate dimensions
	fontWidth := 8  // Approximate character width
	fontHeight := 16 // Line height
	padding := 20
	lineNumberWidth := 40 // Width for line numbers

	// Calculate required width (find longest line)
	maxLineLength := 0
	for _, line := range lines {
		if len(line) > maxLineLength {
			maxLineLength = len(line)
		}
	}

	// Calculate image dimensions
	imgWidth := padding*2 + lineNumberWidth + maxLineLength*fontWidth
	imgHeight := padding*2 + len(lines)*fontHeight

	// Constrain to max dimensions
	if imgWidth > opts.Width {
		imgWidth = opts.Width
	}
	if imgHeight > opts.Height {
		imgHeight = opts.Height
	}

	// Ensure minimum dimensions
	if imgWidth < padding*2+lineNumberWidth {
		imgWidth = padding*2 + lineNumberWidth + 50
	}
	if imgHeight < padding*2+fontHeight {
		imgHeight = padding*2 + fontHeight + 10
	}

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{opts.Background}, image.Point{}, draw.Src)

	// Draw text
	face := basicfont.Face7x13
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{R: 200, G: 200, B: 200, A: 255}), // Light gray text
		Face: face,
		Dot:  fixed.P(padding, padding+fontHeight),
	}

	// Draw line numbers and text
	for i, line := range lines {
		y := padding + (i * fontHeight)
		if y+fontHeight > imgHeight {
			break
		}

		// Draw line number
		lineNum := fmt.Sprintf("%3d ", i+1)
		drawer.Dot = fixed.P(padding, y+fontHeight)
		drawer.DrawString(lineNum)

		// Draw line content (truncate if too long)
		maxChars := (imgWidth - padding*2 - lineNumberWidth) / fontWidth
		if maxChars < 0 {
			maxChars = 0
		}
		displayLine := line
		if len(displayLine) > maxChars {
			if maxChars > 3 {
				displayLine = displayLine[:maxChars-3] + "..."
			} else {
				displayLine = displayLine[:maxChars]
			}
		}

		drawer.Dot = fixed.P(padding+lineNumberWidth, y+fontHeight)
		drawer.DrawString(displayLine)
	}

	// Draw title bar with filename
	titleHeight := 30
	titleImg := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight+titleHeight))
	draw.Draw(titleImg, titleImg.Bounds(), &image.Uniform{color.RGBA{R: 40, G: 40, B: 55, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(titleImg, image.Rect(0, titleHeight, imgWidth, imgHeight+titleHeight), img, image.Point{}, draw.Src)

	// Draw filename in title bar
	titleDrawer := &font.Drawer{
		Dst:  titleImg,
		Src:  image.NewUniform(color.RGBA{R: 150, G: 150, B: 150, A: 255}),
		Face: face,
		Dot:  fixed.P(padding, titleHeight-10),
	}
	titleDrawer.DrawString(info.Extension[1:]) // Remove the dot

	// Resize to fit dimensions if needed
	result := imaging.Resize(titleImg, opts.Width, opts.Height, imaging.Lanczos)

	return &models.ThumbnailResult{
		Image:    result,
		MimeType: "image/png",
		Width:    opts.Width,
		Height:   opts.Height,
	}, nil
}

// init registers the code format decoders
func init() {
	// Chroma is registered via blank imports above
}
