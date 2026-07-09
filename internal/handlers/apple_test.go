package handlers

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

// heifConvertAvailable checks whether the heif-convert external tool is installed.
func heifConvertAvailable() bool {
	_, err := exec.LookPath("heif-convert")
	return err == nil
}

func TestImageHandler_AppleFormats(t *testing.T) {
	handler := &ImageHandler{}
	opts := &models.ThumbnailOptions{
		Width:      128,
		Height:     128,
		Format:     "png",
		Background: models.DefaultOptions().Background,
	}

	tests := []struct {
		name      string
		filename  string
		needsTool bool // true if this format requires the heif-convert external tool
	}{
		{"HEIC", "sample.heic", true},
		{"AVIF", "sample.avif", true},
		{"ICNS", "sample.icns", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := appleFixturePath(tt.filename)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("fixture not found: %s", path)
			}

			// HEIC/AVIF decoding requires the external heif-convert tool;
			// skip gracefully when it is not installed (e.g. CI without libheif-examples).
			if tt.needsTool && !heifConvertAvailable() {
				t.Skip("heif-convert not installed; skipping HEIC/AVIF test")
			}

			info := &models.FileInfo{
				Path:      path,
				Extension: extFromName(tt.filename),
				FileType:  models.FileTypeImage,
			}

			if !handler.CanHandle(info) {
				t.Fatal("CanHandle() returned false for image type")
			}

			result, err := handler.Generate(info, opts)
			if err != nil {
				// If the error is due to a missing external tool, skip instead of fail.
				if strings.Contains(err.Error(), "executable file not found") {
					t.Skipf("external tool not available: %v", err)
				}
				t.Fatalf("Generate() error: %v", err)
			}
			if result == nil {
				t.Fatal("Generate() returned nil result")
			}
			if result.Image == nil {
				t.Fatal("Generate() returned nil image")
			}
			bounds := result.Image.Bounds()
			if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
				t.Fatalf("invalid image dimensions: %dx%d", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestDiskImageHandler_CanHandle(t *testing.T) {
	handler := &DiskImageHandler{}

	tests := []struct {
		name    string
		ftype   models.FileType
		expect  bool
	}{
		{"DiskImage", models.FileTypeDiskImage, true},
		{"Image", models.FileTypeImage, false},
		{"Video", models.FileTypeVideo, false},
		{"Archive", models.FileTypeArchive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &models.FileInfo{FileType: tt.ftype}
			if got := handler.CanHandle(info); got != tt.expect {
				t.Errorf("CanHandle() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestDiskImageHandler_Generate(t *testing.T) {
	handler := &DiskImageHandler{}
	opts := &models.ThumbnailOptions{
		Width:      64,
		Height:     64,
		Format:     "png",
		Background: models.DefaultOptions().Background,
	}

	info := &models.FileInfo{
		Path:      "test.dmg",
		Extension: ".dmg",
		FileType:  models.FileTypeDiskImage,
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if result == nil {
		t.Fatal("Generate() returned nil result")
	}
	if result.Image == nil {
		t.Fatal("Generate() returned nil image")
	}
	bounds := result.Image.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("dimensions = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
	}
}

// appleFixturePath returns the path to a test fixture
func appleFixturePath(filename string) string {
	return "../../tests/fixtures/" + filename
}

// extFromName extracts the extension from a filename
func extFromName(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[i:]
		}
	}
	return ""
}
