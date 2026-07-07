package detect

import (
	"path/filepath"
	"testing"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

// fixturePath returns the path to a test fixture file
func fixturePath(filename string) string {
	return filepath.Join("..", "..", "tests", "fixtures", filename)
}

func TestDetect_AppleImageFormats(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantType models.FileType
		wantMIME string
	}{
		{"HEIC", "sample.heic", models.FileTypeImage, "image/heic"},
		{"AVIF", "sample.avif", models.FileTypeImage, "image/avif"},
		{"ICNS", "sample.icns", models.FileTypeImage, "image/icns"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fixturePath(tt.filename)
			info, err := Detect(path)
			if err != nil {
				t.Fatalf("Detect() error: %v", err)
			}
			if info.FileType != tt.wantType {
				t.Errorf("FileType = %v, want %v", info.FileType, tt.wantType)
			}
			if info.MimeType != tt.wantMIME {
				t.Errorf("MimeType = %q, want %q", info.MimeType, tt.wantMIME)
			}
		})
	}
}

func TestDetect_AppleAudioFormats(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantType models.FileType
		wantMIME string
	}{
		{"ALAC", "sample.alac.m4a", models.FileTypeAudio, "audio/mp4"},
		{"AIFF", "sample.aiff", models.FileTypeAudio, "audio/aiff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fixturePath(tt.filename)
			info, err := Detect(path)
			if err != nil {
				t.Fatalf("Detect() error: %v", err)
			}
			if info.FileType != tt.wantType {
				t.Errorf("FileType = %v, want %v", info.FileType, tt.wantType)
			}
			if info.MimeType != tt.wantMIME {
				t.Errorf("MimeType = %q, want %q", info.MimeType, tt.wantMIME)
			}
		})
	}
}

func TestDetect_AppleVideoFormats(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantType models.FileType
		wantMIME string
	}{
		{"ProRes MOV", "sample_prores.mov", models.FileTypeVideo, "video/quicktime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fixturePath(tt.filename)
			info, err := Detect(path)
			if err != nil {
				t.Fatalf("Detect() error: %v", err)
			}
			if info.FileType != tt.wantType {
				t.Errorf("FileType = %v, want %v", info.FileType, tt.wantType)
			}
			if info.MimeType != tt.wantMIME {
				t.Errorf("MimeType = %q, want %q", info.MimeType, tt.wantMIME)
			}
		})
	}
}

func TestDetect_AppleIWorkFormats(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantType models.FileType
		wantMIME string
	}{
		{"Pages", "sample.pages", models.FileTypeOffice, "application/vnd.apple.pages"},
		{"Numbers", "sample.numbers", models.FileTypeOffice, "application/vnd.apple.numbers"},
		{"Keynote", "sample.key", models.FileTypeOffice, "application/vnd.apple.keynote"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fixturePath(tt.filename)
			info, err := Detect(path)
			if err != nil {
				t.Fatalf("Detect() error: %v", err)
			}
			if info.FileType != tt.wantType {
				t.Errorf("FileType = %v, want %v", info.FileType, tt.wantType)
			}
			if info.MimeType != tt.wantMIME {
				t.Errorf("MimeType = %q, want %q", info.MimeType, tt.wantMIME)
			}
		})
	}
}

func TestDetect_AppleIPArchive(t *testing.T) {
	path := fixturePath("sample.ipa")
	info, err := Detect(path)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if info.FileType != models.FileTypeArchive {
		t.Errorf("FileType = %v, want %v", info.FileType, models.FileTypeArchive)
	}
}

func TestDetect_AppleDMG(t *testing.T) {
	path := fixturePath("sample.dmg")
	info, err := Detect(path)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if info.FileType != models.FileTypeDiskImage {
		t.Errorf("FileType = %v, want %v", info.FileType, models.FileTypeDiskImage)
	}
	if info.MimeType != "application/x-apple-diskimage" {
		t.Errorf("MimeType = %q, want %q", info.MimeType, "application/x-apple-diskimage")
	}
}

func TestFileType_String_DiskImage(t *testing.T) {
	if got := models.FileTypeDiskImage.String(); got != "Disk Image" {
		t.Errorf("FileTypeDiskImage.String() = %q, want %q", got, "Disk Image")
	}
}
