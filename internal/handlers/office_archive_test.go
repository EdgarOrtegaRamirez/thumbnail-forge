package handlers

import (
	"archive/zip"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

func TestOfficeHandler_CanHandle(t *testing.T) {
	handler := &OfficeHandler{}

	tests := []struct {
		name     string
		fileType models.FileType
		want     bool
	}{
		{"Office file", models.FileTypeOffice, true},
		{"Image file", models.FileTypeImage, false},
		{"Video file", models.FileTypeVideo, false},
		{"PDF file", models.FileTypePDF, false},
		{"Audio file", models.FileTypeAudio, false},
		{"Code file", models.FileTypeCode, false},
		{"Text file", models.FileTypeText, false},
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

func TestOfficeHandler_Generate_Placeholder(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy office file (not a real office file)
	officePath := filepath.Join(tmpDir, "test.docx")
	if err := os.WriteFile(officePath, []byte("Not an office file"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test the handler (should fall back to placeholder)
	handler := &OfficeHandler{}
	info := &models.FileInfo{
		Path:     officePath,
		FileType: models.FileTypeOffice,
		MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
	opts := &models.ThumbnailOptions{
		Width:      160,
		Height:     120,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Format:     "png",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify result
	if result.Width != 160 {
		t.Errorf("Result width = %d, want 160", result.Width)
	}
	if result.Height != 120 {
		t.Errorf("Result height = %d, want 120", result.Height)
	}
	if result.MimeType != "image/png" {
		t.Errorf("Result MIME type = %v, want image/png", result.MimeType)
	}

	// Verify the image is not nil
	if result.Image == nil {
		t.Fatal("Result image is nil")
	}
}

func TestCheckLibreOffice(t *testing.T) {
	// This test just checks if the function runs without error
	// It doesn't fail if LibreOffice is not installed
	err := checkLibreOffice()
	// We don't fail the test if LibreOffice is not installed
	// We just want to make sure the function doesn't panic
	_ = err
}

func TestArchiveHandler_CanHandle(t *testing.T) {
	handler := &ArchiveHandler{}

	tests := []struct {
		name     string
		fileType models.FileType
		want     bool
	}{
		{"Archive file", models.FileTypeArchive, true},
		{"Image file", models.FileTypeImage, false},
		{"Video file", models.FileTypeVideo, false},
		{"PDF file", models.FileTypePDF, false},
		{"Audio file", models.FileTypeAudio, false},
		{"Code file", models.FileTypeCode, false},
		{"Text file", models.FileTypeText, false},
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

func TestArchiveHandler_Generate_Zip(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test ZIP file
	zipPath := filepath.Join(tmpDir, "test.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zipFile.Close() }()

	w := zip.NewWriter(zipFile)

	// Add a file to the ZIP
	f, err := w.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("Hello, World!"))
	if err != nil {
		t.Fatal(err)
	}

	// Close the ZIP writer
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	_ = zipFile.Close()

	// Test the handler
	handler := &ArchiveHandler{}
	info := &models.FileInfo{
		Path:     zipPath,
		FileType: models.FileTypeArchive,
		MimeType: "application/zip",
	}
	opts := &models.ThumbnailOptions{
		Width:      160,
		Height:     120,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Format:     "png",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify result
	if result.Width != 160 {
		t.Errorf("Result width = %d, want 160", result.Width)
	}
	if result.Height != 120 {
		t.Errorf("Result height = %d, want 120", result.Height)
	}
	if result.MimeType != "image/png" {
		t.Errorf("Result MIME type = %v, want image/png", result.MimeType)
	}

	// Verify the image is not nil
	if result.Image == nil {
		t.Fatal("Result image is nil")
	}
}

func TestArchiveHandler_Generate_InvalidArchive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid archive file
	archivePath := filepath.Join(tmpDir, "invalid.zip")
	if err := os.WriteFile(archivePath, []byte("Not a zip file"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test the handler (should fall back to placeholder)
	handler := &ArchiveHandler{}
	info := &models.FileInfo{
		Path:     archivePath,
		FileType: models.FileTypeArchive,
		MimeType: "application/zip",
	}
	opts := &models.ThumbnailOptions{
		Width:      160,
		Height:     120,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Format:     "png",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify the image is not nil (should be placeholder)
	if result.Image == nil {
		t.Fatal("Result image is nil")
	}
}

func TestGetZipContents(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test ZIP file with multiple files
	zipPath := filepath.Join(tmpDir, "test.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zipFile.Close() }()

	w := zip.NewWriter(zipFile)

	// Add multiple files
	files := []string{"file1.txt", "file2.txt", "dir/file3.txt"}
	for _, name := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.Write([]byte("content"))
		if err != nil {
			t.Fatal(err)
		}
	}

	// Close the ZIP writer
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	_ = zipFile.Close()

	// Get the contents
	contents, err := getZipContents(zipPath)
	if err != nil {
		t.Fatalf("getZipContents() error = %v", err)
	}

	// Verify the contents
	if len(contents) != len(files) {
		t.Errorf("getZipContents() returned %d files, want %d", len(contents), len(files))
	}

	// Check that all files are present
	fileMap := make(map[string]bool)
	for _, f := range contents {
		fileMap[f] = true
	}

	for _, f := range files {
		if !fileMap[f] {
			t.Errorf("getZipContents() missing file %s", f)
		}
	}
}

func TestGetArchiveContents_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with unsupported extension
	filePath := filepath.Join(tmpDir, "test.xyz")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Try to get contents (should fail)
	_, err := getArchiveContents(filePath)
	if err == nil {
		t.Error("getArchiveContents() expected error for unsupported format, got nil")
	}
}
