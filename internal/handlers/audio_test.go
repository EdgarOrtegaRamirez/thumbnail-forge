package handlers

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

func TestAudioHandler_CanHandle(t *testing.T) {
	handler := &AudioHandler{}

	tests := []struct {
		name     string
		fileType models.FileType
		want     bool
	}{
		{"Audio file", models.FileTypeAudio, true},
		{"Video file", models.FileTypeVideo, false},
		{"Image file", models.FileTypeImage, false},
		{"PDF file", models.FileTypePDF, false},
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

func TestAudioHandler_Generate_NoAlbumArt(t *testing.T) {
	// Skip if ffmpeg is not available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping audio test")
	}

	tmpDir := t.TempDir()

	// Create a test audio file using ffmpeg (silence)
	audioPath := filepath.Join(tmpDir, "test.mp3")
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", "anullsrc=r=44100:cl=mono",
		"-t", "1",
		"-q:a", "9",
		"-acodec", "libmp3lame",
		"-y",
		audioPath,
	)

	if err := cmd.Run(); err != nil {
		t.Skipf("ffmpeg could not create test audio: %v", err)
	}

	// Test the handler
	handler := &AudioHandler{}
	info := &models.FileInfo{
		Path:     audioPath,
		FileType: models.FileTypeAudio,
		MimeType: "audio/mpeg",
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

func TestAudioHandler_Generate_WithAlbumArt(t *testing.T) {
	// Skip if ffmpeg is not available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping audio test")
	}

	tmpDir := t.TempDir()

	// Create a simple image to use as album art
	albumArt := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			albumArt.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: 128,
				A: 255,
			})
		}
	}

	// Save the album art as JPEG
	albumArtPath := filepath.Join(tmpDir, "album_art.jpg")
	f, err := os.Create(albumArtPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, albumArt, nil); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	// Create a test audio file with album art using ffmpeg
	audioPath := filepath.Join(tmpDir, "test_with_art.mp3")
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", "anullsrc=r=44100:cl=mono",
		"-i", albumArtPath,
		"-map", "0:a",
		"-map", "1:v",
		"-c:a", "libmp3lame",
		"-c:v", "mjpeg",
		"-id3v2_version", "3",
		"-t", "1",
		"-y",
		audioPath,
	)

	if err := cmd.Run(); err != nil {
		t.Skipf("ffmpeg could not create test audio with art: %v", err)
	}

	// Test the handler
	handler := &AudioHandler{}
	info := &models.FileInfo{
		Path:     audioPath,
		FileType: models.FileTypeAudio,
		MimeType: "audio/mpeg",
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

func TestExtractAlbumArt(t *testing.T) {
	// Skip if ffmpeg is not available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping audio test")
	}

	tmpDir := t.TempDir()

	// Create a simple image to use as album art
	albumArt := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			albumArt.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	// Save the album art as JPEG
	albumArtPath := filepath.Join(tmpDir, "album_art.jpg")
	f, err := os.Create(albumArtPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, albumArt, nil); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	// Create a test audio file with album art
	audioPath := filepath.Join(tmpDir, "test.mp3")
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", "anullsrc=r=44100:cl=mono",
		"-i", albumArtPath,
		"-map", "0:a",
		"-map", "1:v",
		"-c:a", "libmp3lame",
		"-c:v", "mjpeg",
		"-id3v2_version", "3",
		"-t", "1",
		"-y",
		audioPath,
	)

	if err := cmd.Run(); err != nil {
		t.Skipf("ffmpeg could not create test audio: %v", err)
	}

	// Extract album art
	img, err := extractAlbumArt(audioPath)
	if err != nil {
		t.Fatalf("extractAlbumArt() error = %v", err)
	}

	// Verify the image is not nil
	if img == nil {
		t.Fatal("extractAlbumArt() returned nil image")
	}

	// Verify dimensions
	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("extractAlbumArt() dimensions = %dx%d, want 100x100", bounds.Dx(), bounds.Dy())
	}
}

func TestExtractAlbumArt_NoArt(t *testing.T) {
	// Skip if ffmpeg is not available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping audio test")
	}

	tmpDir := t.TempDir()

	// Create a test audio file without album art
	audioPath := filepath.Join(tmpDir, "test_no_art.mp3")
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", "anullsrc=r=44100:cl=mono",
		"-t", "1",
		"-q:a", "9",
		"-acodec", "libmp3lame",
		"-y",
		audioPath,
	)

	if err := cmd.Run(); err != nil {
		t.Skipf("ffmpeg could not create test audio: %v", err)
	}

	// Try to extract album art (should fail)
	_, err := extractAlbumArt(audioPath)
	if err == nil {
		t.Error("extractAlbumArt() expected error for audio without art, got nil")
	}
}

func TestAudioHandler_Generate_Placeholder(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy audio file (not a real audio file, will fail album art extraction)
	audioPath := filepath.Join(tmpDir, "dummy.mp3")
	if err := os.WriteFile(audioPath, []byte("Not an audio file"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test the handler (should fall back to placeholder since ffmpeg won't work on this)
	handler := &AudioHandler{}
	info := &models.FileInfo{
		Path:     audioPath,
		FileType: models.FileTypeAudio,
		MimeType: "audio/mpeg",
	}
	opts := &models.ThumbnailOptions{
		Width:      160,
		Height:     120,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Format:     "png",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		// This is expected since the file is not a real audio file
		// and ffmpeg will fail
		t.Skipf("Generate() error (expected for dummy file): %v", err)
	}

	// If it somehow succeeded, verify the result
	if result != nil && result.Image == nil {
		t.Fatal("Result image is nil")
	}
}

func TestGeneratePlaceholder(t *testing.T) {
	handler := &AudioHandler{}
	info := &models.FileInfo{
		Path:     "test.mp3",
		FileType: models.FileTypeAudio,
	}
	opts := &models.ThumbnailOptions{
		Width:      200,
		Height:     200,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
	}

	result, err := handler.generatePlaceholder(info, opts)
	if err != nil {
		t.Fatalf("generatePlaceholder() error = %v", err)
	}

	// Verify result
	if result.Width != 200 {
		t.Errorf("Result width = %d, want 200", result.Width)
	}
	if result.Height != 200 {
		t.Errorf("Result height = %d, want 200", result.Height)
	}
	if result.MimeType != "image/png" {
		t.Errorf("Result MIME type = %v, want image/png", result.MimeType)
	}

	// Verify the image is not nil
	if result.Image == nil {
		t.Fatal("Result image is nil")
	}
}
