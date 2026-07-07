package handlers

import (
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

func TestVideoHandler_CanHandle(t *testing.T) {
	handler := &VideoHandler{}

	tests := []struct {
		name     string
		fileType models.FileType
		want     bool
	}{
		{"Video file", models.FileTypeVideo, true},
		{"Image file", models.FileTypeImage, false},
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

func TestVideoHandler_Generate_MP4(t *testing.T) {
	// Skip if ffmpeg is not available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping video test")
	}

	tmpDir := t.TempDir()

	// Create a test video using ffmpeg
	videoPath := filepath.Join(tmpDir, "test.mp4")
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", "color=c=red:s=320x240:d=2",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-y",
		videoPath,
	)

	if err := cmd.Run(); err != nil {
		t.Skipf("ffmpeg could not create test video: %v", err)
	}

	// Test the handler
	handler := &VideoHandler{}
	info := &models.FileInfo{
		Path:     videoPath,
		FileType: models.FileTypeVideo,
		MimeType: "video/mp4",
	}
	opts := &models.ThumbnailOptions{
		Width:      160,
		Height:     120,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Timestamp:  "1",
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

func TestVideoHandler_Generate_NoFFmpeg(t *testing.T) {
	// This test verifies the error message when ffmpeg is not available
	// We can't easily test this without mocking exec.LookPath
	// So we'll skip it for now
	t.Skip("Cannot test without mocking exec.LookPath")
}

func TestCheckFFmpeg(t *testing.T) {
	// Skip if ffmpeg is not available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping test")
	}

	err := checkFFmpeg()
	if err != nil {
		t.Errorf("checkFFmpeg() error = %v", err)
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		want      float64
		wantErr   bool
	}{
		{"Simple seconds", "1", 1.0, false},
		{"Decimal seconds", "5.5", 5.5, false},
		{"Zero", "0", 0.0, false},
		{"HH:MM:SS", "00:00:05", 5.0, false},
		{"HH:MM:SS complex", "01:30:00", 5400.0, false},
		{"MM:SS", "01:30", 90.0, false},
		{"Invalid format", "abc", 0, true},
		{"Invalid HH:MM:SS", "00:00:abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimestamp(tt.timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractFrame(t *testing.T) {
	// Skip if ffmpeg is not available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping test")
	}

	tmpDir := t.TempDir()

	// Create a test video
	videoPath := filepath.Join(tmpDir, "test.mp4")
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", "color=c=blue:s=320x240:d=2",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-y",
		videoPath,
	)

	if err := cmd.Run(); err != nil {
		t.Skipf("ffmpeg could not create test video: %v", err)
	}

	// Extract a frame
	outputPath := filepath.Join(tmpDir, "frame.jpg")
	err := extractFrame(videoPath, outputPath, "1")
	if err != nil {
		t.Fatalf("extractFrame() error = %v", err)
	}

	// Verify the frame was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("extractFrame() did not create output file")
	}
}

func TestExtractFrame_InvalidVideo(t *testing.T) {
	// Skip if ffmpeg is not available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping test")
	}

	tmpDir := t.TempDir()

	// Create an invalid video file
	videoPath := filepath.Join(tmpDir, "invalid.mp4")
	if err := os.WriteFile(videoPath, []byte("Not a video"), 0644); err != nil {
		t.Fatal(err)
	}

	// Try to extract a frame
	outputPath := filepath.Join(tmpDir, "frame.jpg")
	err := extractFrame(videoPath, outputPath, "1")
	if err == nil {
		t.Error("extractFrame() expected error for invalid video, got nil")
	}
}
