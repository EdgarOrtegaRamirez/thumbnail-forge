package models

import (
	"image"
	"image/color"
)

// FileType represents the category of file being processed
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeImage
	FileTypeVideo
	FileTypePDF
	FileTypeOffice
	FileTypeAudio
	FileTypeCode
	FileTypeText
	FileTypeMarkdown
	FileTypeArchive
	FileTypeDiskImage
)

// String returns the human-readable name of the file type
func (ft FileType) String() string {
	switch ft {
	case FileTypeImage:
		return "Image"
	case FileTypeVideo:
		return "Video"
	case FileTypePDF:
		return "PDF"
	case FileTypeOffice:
		return "Office Document"
	case FileTypeAudio:
		return "Audio"
	case FileTypeCode:
		return "Source Code"
	case FileTypeText:
		return "Text File"
	case FileTypeMarkdown:
		return "Markdown"
	case FileTypeArchive:
		return "Archive"
	case FileTypeDiskImage:
		return "Disk Image"
	default:
		return "Unknown"
	}
}

// FileInfo contains information about the detected file
type FileInfo struct {
	Path       string
	MimeType   string
	Extension  string
	FileType   FileType
	Size       int64
	IsAnimated bool // For GIFs
}

// ThumbnailOptions contains user-provided options for thumbnail generation
type ThumbnailOptions struct {
	Width      int
	Height     int
	OutputPath string
	Format     string // png, jpg, webp
	Quality    int    // 1-100 for lossy formats
	Background color.Color
	Timestamp  string // Video frame timestamp (e.g., "1s", "00:00:05")
	Page       int    // PDF page number
	Theme      string // Code theme: dracula, monokai, github
	Terminal   bool   // Output to terminal
	Freedesktop bool  // Cache to ~/.cache/thumbnails/
	List       bool   // List supported file types
}

// ThumbnailResult contains the generated thumbnail
type ThumbnailResult struct {
	Image    image.Image
	MimeType string
	Width    int
	Height   int
}

// DefaultOptions returns the default thumbnail options
func DefaultOptions() *ThumbnailOptions {
	return &ThumbnailOptions{
		Width:      256,
		Height:     256,
		Format:     "png",
		Quality:    85,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255}, // #1e1e2e
		Page:       1,
		Theme:      "dracula",
	}
}

// Handler is the interface that all file type handlers must implement
type Handler interface {
	// CanHandle returns true if this handler can process the given file type
	CanHandle(info *FileInfo) bool

	// Generate creates a thumbnail from the file
	Generate(info *FileInfo, opts *ThumbnailOptions) (*ThumbnailResult, error)
}
