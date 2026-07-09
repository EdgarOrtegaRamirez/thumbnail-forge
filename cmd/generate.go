package cmd

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/detect"
	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/handlers"
	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/terminal"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate <file>",
	Short: "Generate a thumbnail from a file",
	Long: `Generate a thumbnail from almost any file type.

The file type is automatically detected using magic bytes and file extension.
Different file types use different processing pipelines:

  Images → Decode, resize, composite
  Video  → Extract frame via ffmpeg, resize
  PDF    → Render page via MuPDF, resize
  Office → Convert to PDF via LibreOffice, render page
  Audio  → Extract album art or generate waveform
  Code   → Syntax highlight, render to image
  Text   → Render text content to image
  Archive → List contents, render folder icon`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerate,
}

var (
	flagWidth      int
	flagHeight     int
	flagOutput     string
	flagFormat     string
	flagQuality    int
	flagBackground string
	flagTimestamp  string
	flagPage       int
	flagTheme      string
	flagTerminal   bool
)

func init() {
	generateCmd.Flags().IntVar(&flagWidth, "width", 256, "Thumbnail width in pixels")
	generateCmd.Flags().IntVar(&flagHeight, "height", 256, "Thumbnail height in pixels")
	generateCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Output file path (default: stdout as PNG)")
	generateCmd.Flags().StringVarP(&flagFormat, "format", "f", "png", "Output format: png, jpg, webp")
	generateCmd.Flags().IntVarP(&flagQuality, "quality", "q", 85, "JPEG/WebP quality 1-100")
	generateCmd.Flags().StringVar(&flagBackground, "background", "#1e1e2e", "Background color for transparent files (hex)")
	generateCmd.Flags().StringVar(&flagTimestamp, "timestamp", "1s", "Video frame timestamp (e.g., 1s, 00:00:05)")
	generateCmd.Flags().IntVar(&flagPage, "page", 1, "PDF page number (1-indexed)")
	generateCmd.Flags().StringVar(&flagTheme, "theme", "dracula", "Code theme: dracula, monokai, github")
	generateCmd.Flags().BoolVar(&flagTerminal, "terminal", false, "Output to terminal instead of file")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Detect file type
	info, err := detect.Detect(filePath)
	if err != nil {
		return fmt.Errorf("failed to detect file type: %w", err)
	}

	// Parse background color
	bgColor, err := parseColor(flagBackground)
	if err != nil {
		return fmt.Errorf("invalid background color: %w", err)
	}

	// Build options
	opts := &models.ThumbnailOptions{
		Width:      flagWidth,
		Height:     flagHeight,
		OutputPath: flagOutput,
		Format:     flagFormat,
		Quality:    flagQuality,
		Background: bgColor,
		Timestamp:  flagTimestamp,
		Page:       flagPage,
		Theme:      flagTheme,
		Terminal:   flagTerminal,
	}

	// For now, just show what we detected
	fmt.Printf("File: %s\n", info.Path)
	fmt.Printf("Type: %s\n", info.FileType)
	fmt.Printf("MIME: %s\n", info.MimeType)
	fmt.Printf("Size: %d bytes\n", info.Size)
	fmt.Printf("Extension: %s\n", info.Extension)
	if info.IsAnimated {
		fmt.Printf("Animated: yes\n")
	}
	fmt.Printf("\nOptions:\n")
	fmt.Printf("  Width: %d\n", opts.Width)
	fmt.Printf("  Height: %d\n", opts.Height)
	fmt.Printf("  Format: %s\n", opts.Format)
	fmt.Printf("  Quality: %d\n", opts.Quality)
	fmt.Printf("  Background: %s\n", flagBackground)
	fmt.Printf("  Theme: %s\n", opts.Theme)

	// Route to appropriate handler
	var result *models.ThumbnailResult

	switch info.FileType {
	case models.FileTypeImage:
		handler := &handlers.ImageHandler{}
		if handler.CanHandle(info) {
			result, err = handler.Generate(info, opts)
			if err != nil {
				return fmt.Errorf("failed to generate thumbnail: %w", err)
			}
		}
	case models.FileTypeCode, models.FileTypeText, models.FileTypeMarkdown:
		handler := &handlers.CodeHandler{}
		if handler.CanHandle(info) {
			result, err = handler.Generate(info, opts)
			if err != nil {
				return fmt.Errorf("failed to generate thumbnail: %w", err)
			}
		}
	case models.FileTypePDF:
		handler := &handlers.PDFHandler{}
		if handler.CanHandle(info) {
			result, err = handler.Generate(info, opts)
			if err != nil {
				return fmt.Errorf("failed to generate thumbnail: %w", err)
			}
		}
	case models.FileTypeVideo:
		handler := &handlers.VideoHandler{}
		if handler.CanHandle(info) {
			result, err = handler.Generate(info, opts)
			if err != nil {
				return fmt.Errorf("failed to generate thumbnail: %w", err)
			}
		}
	case models.FileTypeAudio:
		handler := &handlers.AudioHandler{}
		if handler.CanHandle(info) {
			result, err = handler.Generate(info, opts)
			if err != nil {
				return fmt.Errorf("failed to generate thumbnail: %w", err)
			}
		}
	case models.FileTypeOffice:
		handler := &handlers.OfficeHandler{}
		if handler.CanHandle(info) {
			result, err = handler.Generate(info, opts)
			if err != nil {
				return fmt.Errorf("failed to generate thumbnail: %w", err)
			}
		}
	case models.FileTypeArchive:
		handler := &handlers.ArchiveHandler{}
		if handler.CanHandle(info) {
			result, err = handler.Generate(info, opts)
			if err != nil {
				return fmt.Errorf("failed to generate thumbnail: %w", err)
			}
		}
	case models.FileTypeDiskImage:
		handler := &handlers.DiskImageHandler{}
		if handler.CanHandle(info) {
			result, err = handler.Generate(info, opts)
			if err != nil {
				return fmt.Errorf("failed to generate thumbnail: %w", err)
			}
		}
	}

	// Fallback to placeholder for unhandled types
	if result == nil {
		result = generatePlaceholderThumbnail(info, opts)
		fmt.Printf("\nNote: Placeholder thumbnail generated for %s files\n", info.FileType)
		fmt.Printf("Real thumbnail generation coming in future phases\n")
	}

	// Output to terminal if requested
	if opts.Terminal {
		return outputToTerminal(result.Image, result)
	}

	// Save output
	if opts.OutputPath != "" {
		return saveImage(result.Image, opts.OutputPath, opts.Format, opts.Quality)
	}

	// Save to temp file and show path
	ext := "." + opts.Format
	if opts.Format == "jpg" {
		ext = ".jpeg"
	}
	tmpFile, err := os.CreateTemp("", "thumbnail-*"+ext)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = tmpFile.Close() }()

	if err := saveImage(result.Image, tmpFile.Name(), opts.Format, opts.Quality); err != nil {
		return err
	}

	fmt.Printf("\nThumbnail saved to: %s\n", tmpFile.Name())
	return nil
}

func generatePlaceholderThumbnail(info *models.FileInfo, opts *models.ThumbnailOptions) *models.ThumbnailResult {
	// Create a placeholder image with the file type info
	img := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))

	// Fill background
	for y := 0; y < opts.Height; y++ {
		for x := 0; x < opts.Width; x++ {
			img.Set(x, y, opts.Background)
		}
	}

	return &models.ThumbnailResult{
		Image:    img,
		MimeType: "image/png",
		Width:    opts.Width,
		Height:   opts.Height,
	}
}

func saveImage(img image.Image, path, format string, quality int) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = f.Close() }()

	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		return jpeg.Encode(f, img, &jpeg.Options{Quality: quality})
	case "png":
		encoder := &png.Encoder{CompressionLevel: png.BestSpeed}
		return encoder.Encode(f, img)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func parseColor(hex string) (color.Color, error) {
	hex = strings.TrimPrefix(hex, "#")

	switch len(hex) {
	case 6:
		r, err := strconv.ParseUint(hex[0:2], 16, 8)
		if err != nil {
			return nil, err
		}
		g, err := strconv.ParseUint(hex[2:4], 16, 8)
		if err != nil {
			return nil, err
		}
		b, err := strconv.ParseUint(hex[4:6], 16, 8)
		if err != nil {
			return nil, err
		}
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
	case 8:
		r, err := strconv.ParseUint(hex[0:2], 16, 8)
		if err != nil {
			return nil, err
		}
		g, err := strconv.ParseUint(hex[2:4], 16, 8)
		if err != nil {
			return nil, err
		}
		b, err := strconv.ParseUint(hex[4:6], 16, 8)
		if err != nil {
			return nil, err
		}
		a, err := strconv.ParseUint(hex[6:8], 16, 8)
		if err != nil {
			return nil, err
		}
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}, nil
	default:
		return nil, fmt.Errorf("invalid hex color length: %s", hex)
	}
}

func outputToTerminal(img image.Image, result *models.ThumbnailResult) error {
	output := terminal.NewOutput(terminal.ProtocolAuto)
	return output.Display(img, result)
}

// ensure this file is used
var _ = filepath.Base
