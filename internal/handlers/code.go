package handlers

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
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

// readTruncatedFile reads up to maxLines or maxBytes of a file
func readTruncatedFile(path string, maxLines int, maxBytes int64) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	var buf strings.Builder
	reader := bufio.NewReader(file)
	linesRead := 0
	bytesRead := int64(0)

	for linesRead < maxLines && bytesRead < maxBytes {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			if bytesRead+int64(len(line)) > maxBytes {
				allowed := maxBytes - bytesRead
				buf.WriteString(line[:allowed])
				break
			}
			buf.WriteString(line)
			bytesRead += int64(len(line))
			linesRead++
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
	}

	return buf.String(), nil
}

// Generate creates a thumbnail from a code or text file
func (h *CodeHandler) Generate(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Read file content up to 100 lines or 50KB to optimize memory and speed on large files
	content, err := readTruncatedFile(info.Path, 100, 50*1024)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// For markdown, try to render with syntax highlighting
	if info.FileType == models.FileTypeMarkdown {
		// Markdown is plain text with formatting, render as text
		return h.renderText(content, info, opts)
	}

	// For code files, try syntax highlighting
	if info.FileType == models.FileTypeCode {
		return h.renderCode(content, info, opts)
	}

	// For text files, render as plain text
	return h.renderText(content, info, opts)
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

	// Resolve style-defined default foreground and background colors
	var defaultFg color.Color = color.RGBA{R: 220, G: 220, B: 220, A: 255}
	textEntry := style.Get(chroma.Text)
	if textEntry.Colour.IsSet() {
		defaultFg = color.RGBA{R: textEntry.Colour.Red(), G: textEntry.Colour.Green(), B: textEntry.Colour.Blue(), A: 255}
	}

	var defaultBg color.Color = opts.Background
	bgEntry := style.Get(chroma.Background)
	if bgEntry.Background.IsSet() {
		defaultBg = color.RGBA{R: bgEntry.Background.Red(), G: bgEntry.Background.Green(), B: bgEntry.Background.Blue(), A: 255}
	}

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		// Fall back to plain text rendering
		return h.renderText(content, info, opts)
	}

	tokens := iterator.Tokens()

	return h.renderHighlightedCode(tokens, defaultFg, defaultBg, info, opts)
}

// renderText renders plain text to an image
func (h *CodeHandler) renderText(content string, info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	tokens := []chroma.Token{
		{Type: chroma.Text, Value: content},
	}
	defaultFg := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	return h.renderHighlightedCode(tokens, defaultFg, opts.Background, info, opts)
}

// StyledText represents a styled span of text
type StyledText struct {
	Text  string
	Color color.Color
}

// CodeLine represents a line of styled text
type CodeLine []StyledText

// Length returns the total character length of the line
func (cl CodeLine) Length() int {
	length := 0
	for _, st := range cl {
		length += len(st.Text)
	}
	return length
}

// renderHighlightedCode renders tokens directly onto a canvas with their styles and syntax colors
func (h *CodeHandler) renderHighlightedCode(tokens []chroma.Token, defaultFg color.Color, defaultBg color.Color, info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Let's group tokens into lines, handling internal newlines in tokens correctly.
	var lines []CodeLine
	var currentLine CodeLine

	style := styles.Get(opts.Theme)
	if style == nil {
		style = styles.Fallback
	}

	for _, t := range tokens {
		parts := strings.Split(t.Value, "\n")

		// Map token type to style color
		entry := style.Get(t.Type)
		col := defaultFg
		if entry.Colour.IsSet() {
			col = color.RGBA{R: entry.Colour.Red(), G: entry.Colour.Green(), B: entry.Colour.Blue(), A: 255}
		}

		for i, part := range parts {
			if i > 0 {
				lines = append(lines, currentLine)
				currentLine = CodeLine{}
			}
			if len(part) > 0 {
				currentLine = append(currentLine, StyledText{Text: part, Color: col})
			}
		}
	}
	lines = append(lines, currentLine)

	// Limit to first 20 lines for thumbnail
	maxLines := 20
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	// Calculate dimensions
	fontWidth := 8   // Approximate character width
	fontHeight := 16 // Line height
	padding := 20
	lineNumberWidth := 40 // Width for line numbers

	// Calculate required width (find longest line)
	maxLineLength := 0
	for _, line := range lines {
		lineLen := line.Length()
		if lineLen > maxLineLength {
			maxLineLength = lineLen
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
	draw.Draw(img, img.Bounds(), &image.Uniform{defaultBg}, image.Point{}, draw.Src)

	// Draw text
	face := basicfont.Face7x13
	drawer := &font.Drawer{
		Dst:  img,
		Face: face,
	}

	// Draw line numbers and text
	for i, line := range lines {
		y := padding + (i * fontHeight)
		if y+fontHeight > imgHeight {
			break
		}

		// Draw line number
		lineNum := fmt.Sprintf("%3d ", i+1)
		drawer.Src = image.NewUniform(color.RGBA{R: 120, G: 120, B: 120, A: 255})
		drawer.Dot = fixed.P(padding, y+fontHeight)
		drawer.DrawString(lineNum)

		// Draw line content (truncate if too long)
		maxChars := (imgWidth - padding*2 - lineNumberWidth) / fontWidth
		if maxChars < 0 {
			maxChars = 0
		}

		currentX := padding + lineNumberWidth
		charsDrawn := 0

		for _, st := range line {
			if charsDrawn >= maxChars {
				break
			}

			displayStr := st.Text
			// Expand tabs to 4 spaces
			displayStr = strings.ReplaceAll(displayStr, "\t", "    ")

			if charsDrawn+len(displayStr) > maxChars {
				needed := maxChars - charsDrawn
				if needed > 3 {
					displayStr = displayStr[:needed-3] + "..."
				} else {
					displayStr = displayStr[:needed]
				}
			}

			drawer.Src = image.NewUniform(st.Color)
			drawer.Dot = fixed.P(currentX, y+fontHeight)
			drawer.DrawString(displayStr)

			currentX += len(displayStr) * fontWidth
			charsDrawn += len(displayStr)
		}
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

	extName := ""
	if len(info.Extension) > 1 {
		extName = info.Extension[1:]
	} else {
		extName = filepath.Base(info.Path)
	}
	titleDrawer.DrawString(extName)

	// Resize to fit dimensions if needed
	result := imaging.Resize(titleImg, opts.Width, opts.Height, imaging.Lanczos)

	return &models.ThumbnailResult{
		Image:    result,
		MimeType: "image/png",
		Width:    opts.Width,
		Height:   opts.Height,
	}, nil
}

func init() {
	// Chroma lexers/styles are registered via blank imports or standard library
}
