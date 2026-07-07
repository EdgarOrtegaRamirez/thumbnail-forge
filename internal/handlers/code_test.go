package handlers

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

func TestCodeHandler_CanHandle(t *testing.T) {
	handler := &CodeHandler{}

	tests := []struct {
		name     string
		fileType models.FileType
		want     bool
	}{
		{"Code file", models.FileTypeCode, true},
		{"Text file", models.FileTypeText, true},
		{"Markdown file", models.FileTypeMarkdown, true},
		{"Image file", models.FileTypeImage, false},
		{"Video file", models.FileTypeVideo, false},
		{"PDF file", models.FileTypePDF, false},
		{"Audio file", models.FileTypeAudio, false},
		{"Archive file", models.FileTypeArchive, false},
		{"Unknown file", models.FileTypeUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &models.FileInfo{FileType: tt.fileType}
			if got := handler.CanHandle(info); got != tt.want {
				t.Errorf("CanHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCodeHandler_Generate_GoCode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test Go file
	content := `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
`
	path := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test the handler
	handler := &CodeHandler{}
	info := &models.FileInfo{
		Path:      path,
		FileType:  models.FileTypeCode,
		MimeType:  "text/plain",
		Extension: ".go",
	}
	opts := &models.ThumbnailOptions{
		Width:      400,
		Height:     300,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Theme:      "dracula",
		Format:     "png",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify result
	if result.Width != 400 {
		t.Errorf("Result width = %d, want 400", result.Width)
	}
	if result.Height != 300 {
		t.Errorf("Result height = %d, want 300", result.Height)
	}
	if result.MimeType != "image/png" {
		t.Errorf("Result MIME type = %v, want image/png", result.MimeType)
	}

	// Verify the image is not nil
	if result.Image == nil {
		t.Fatal("Result image is nil")
	}
}

func TestCodeHandler_Generate_PythonCode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test Python file
	content := `#!/usr/bin/env python3
def hello():
    print("Hello, World!")

if __name__ == "__main__":
    hello()
`
	path := filepath.Join(tmpDir, "test.py")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test the handler
	handler := &CodeHandler{}
	info := &models.FileInfo{
		Path:      path,
		FileType:  models.FileTypeCode,
		MimeType:  "text/plain",
		Extension: ".py",
	}
	opts := &models.ThumbnailOptions{
		Width:      400,
		Height:     300,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Theme:      "dracula",
		Format:     "png",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify the image is not nil
	if result.Image == nil {
		t.Fatal("Result image is nil")
	}
}

func TestCodeHandler_Generate_TextFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test text file
	content := `This is a sample text file.
It has multiple lines.
Each line should be rendered.
`
	path := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test the handler
	handler := &CodeHandler{}
	info := &models.FileInfo{
		Path:      path,
		FileType:  models.FileTypeText,
		MimeType:  "text/plain",
		Extension: ".txt",
	}
	opts := &models.ThumbnailOptions{
		Width:      400,
		Height:     300,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Theme:      "dracula",
		Format:     "png",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify the image is not nil
	if result.Image == nil {
		t.Fatal("Result image is nil")
	}
}

func TestCodeHandler_Generate_MarkdownFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test Markdown file
	content := `# Title

This is **bold** text.

- Item 1
- Item 2
- Item 3
`
	path := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test the handler
	handler := &CodeHandler{}
	info := &models.FileInfo{
		Path:      path,
		FileType:  models.FileTypeMarkdown,
		MimeType:  "text/markdown",
		Extension: ".md",
	}
	opts := &models.ThumbnailOptions{
		Width:      400,
		Height:     300,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Theme:      "dracula",
		Format:     "png",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify the image is not nil
	if result.Image == nil {
		t.Fatal("Result image is nil")
	}
}

func TestCodeHandler_Generate_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an empty file
	path := filepath.Join(tmpDir, "empty.go")
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	// Test the handler
	handler := &CodeHandler{}
	info := &models.FileInfo{
		Path:      path,
		FileType:  models.FileTypeCode,
		MimeType:  "text/plain",
		Extension: ".go",
	}
	opts := &models.ThumbnailOptions{
		Width:      400,
		Height:     300,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Theme:      "dracula",
		Format:     "png",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify the image is not nil
	if result.Image == nil {
		t.Fatal("Result image is nil")
	}
}

func TestCodeHandler_Generate_LongCode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with many lines
	content := ""
	for i := 0; i < 50; i++ {
		content += "line " + string(rune('A'+i%26)) + "\n"
	}

	path := filepath.Join(tmpDir, "long.go")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test the handler
	handler := &CodeHandler{}
	info := &models.FileInfo{
		Path:      path,
		FileType:  models.FileTypeCode,
		MimeType:  "text/plain",
		Extension: ".go",
	}
	opts := &models.ThumbnailOptions{
		Width:      400,
		Height:     300,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Theme:      "dracula",
		Format:     "png",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify the image is not nil
	if result.Image == nil {
		t.Fatal("Result image is nil")
	}
}
