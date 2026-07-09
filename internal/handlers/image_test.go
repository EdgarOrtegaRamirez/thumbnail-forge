package handlers

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

func TestImageHandler_CanHandle(t *testing.T) {
	handler := &ImageHandler{}

	tests := []struct {
		name     string
		fileType models.FileType
		want     bool
	}{
		{"Image file", models.FileTypeImage, true},
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

func TestImageHandler_Generate_PNG(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test PNG image (100x100 red square)
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	// Save test image
	path := filepath.Join(tmpDir, "test.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()

	// Test the handler
	handler := &ImageHandler{}
	info := &models.FileInfo{
		Path:     path,
		FileType: models.FileTypeImage,
		MimeType: "image/png",
	}
	opts := &models.ThumbnailOptions{
		Width:      50,
		Height:     50,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Format:     "png",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify result
	if result.Width != 50 {
		t.Errorf("Result width = %d, want 50", result.Width)
	}
	if result.Height != 50 {
		t.Errorf("Result height = %d, want 50", result.Height)
	}
	if result.MimeType != "image/png" {
		t.Errorf("Result MIME type = %v, want image/png", result.MimeType)
	}

	// Verify the image is not nil
	if result.Image == nil {
		t.Fatal("Result image is nil")
	}

	// Verify dimensions
	bounds := result.Image.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("Image dimensions = %dx%d, want 50x50", bounds.Dx(), bounds.Dy())
	}
}

func TestImageHandler_Generate_JPEG(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test JPEG image (200x100 blue rectangle)
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{R: 0, G: 0, B: 255, A: 255})
		}
	}

	// Save test image
	path := filepath.Join(tmpDir, "test.jpg")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()

	// Test the handler
	handler := &ImageHandler{}
	info := &models.FileInfo{
		Path:     path,
		FileType: models.FileTypeImage,
		MimeType: "image/jpeg",
	}
	opts := &models.ThumbnailOptions{
		Width:      100,
		Height:     100,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Format:     "jpg",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify result
	if result.Width != 100 {
		t.Errorf("Result width = %d, want 100", result.Width)
	}
	if result.Height != 100 {
		t.Errorf("Result height = %d, want 100", result.Height)
	}

	// Verify aspect ratio preservation
	bounds := result.Image.Bounds()
	origAspect := float64(200) / float64(100)
	newAspect := float64(bounds.Dx()) / float64(bounds.Dy())
	if origAspect != newAspect {
		t.Errorf("Aspect ratio changed: orig=%v, new=%v", origAspect, newAspect)
	}
}

func TestImageHandler_Generate_Transparency(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test PNG with transparency
	img := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			// Semi-transparent red
			img.Set(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 128})
		}
	}

	// Save test image
	path := filepath.Join(tmpDir, "test_transparent.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()

	// Test the handler with custom background
	handler := &ImageHandler{}
	info := &models.FileInfo{
		Path:     path,
		FileType: models.FileTypeImage,
		MimeType: "image/png",
	}
	opts := &models.ThumbnailOptions{
		Width:      50,
		Height:     50,
		Background: color.RGBA{R: 0, G: 255, B: 0, A: 255}, // Green background
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

func TestImageHandler_Generate_AspectRatio(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test image (400x200 - 2:1 aspect ratio)
	img := image.NewRGBA(image.Rect(0, 0, 400, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	// Save test image
	path := filepath.Join(tmpDir, "test_wide.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()

	// Test with 100x100 output - should maintain 2:1 aspect ratio
	handler := &ImageHandler{}
	info := &models.FileInfo{
		Path:     path,
		FileType: models.FileTypeImage,
		MimeType: "image/png",
	}
	opts := &models.ThumbnailOptions{
		Width:      100,
		Height:     100,
		Background: color.RGBA{R: 30, G: 30, B: 46, A: 255},
		Format:     "png",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify aspect ratio is preserved (2:1)
	bounds := result.Image.Bounds()
	aspect := float64(bounds.Dx()) / float64(bounds.Dy())
	expectedAspect := 2.0
	if aspect != expectedAspect {
		t.Errorf("Aspect ratio = %v, want %v", aspect, expectedAspect)
	}
}

func TestLoadImage(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}

	// Save as PNG
	path := filepath.Join(tmpDir, "test_load.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()

	// Load the image
	loaded, err := LoadImage(path)
	if err != nil {
		t.Fatalf("LoadImage() error = %v", err)
	}

	// Verify dimensions
	bounds := loaded.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("Loaded image dimensions = %dx%d, want 50x50", bounds.Dx(), bounds.Dy())
	}
}

func TestLoadImage_Nonexistent(t *testing.T) {
	_, err := LoadImage("/nonexistent/image.png")
	if err == nil {
		t.Error("LoadImage() expected error for nonexistent file, got nil")
	}
}

func TestResizeImage(t *testing.T) {
	// Create a test image (200x100)
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	tests := []struct {
		name      string
		maxWidth  int
		maxHeight int
		wantW     int
		wantH     int
	}{
		{"Fit within square", 100, 100, 100, 50},
		{"Fit width only", 50, 200, 50, 25},
		{"Fit height only", 400, 50, 100, 50},
		{"Don't upscale", 400, 400, 200, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resized := ResizeImage(img, tt.maxWidth, tt.maxHeight)
			bounds := resized.Bounds()
			if bounds.Dx() != tt.wantW || bounds.Dy() != tt.wantH {
				t.Errorf("ResizeImage() dimensions = %dx%d, want %dx%d",
					bounds.Dx(), bounds.Dy(), tt.wantW, tt.wantH)
			}
		})
	}
}

func TestCompositeOnBackground(t *testing.T) {
	// Create a semi-transparent image
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 128})
		}
	}

	bgColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	result := compositeOnBackground(img, bgColor)

	// Verify the result is an RGBA image (not NRGBA)
	if _, ok := result.(*image.RGBA); !ok {
		t.Errorf("compositeOnBackground() returned %T, want *image.RGBA", result)
	}

	// Verify dimensions
	bounds := result.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("Result dimensions = %dx%d, want 10x10", bounds.Dx(), bounds.Dy())
	}
}

func TestGetOutputMimeType(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{"png", "image/png"},
		{"jpg", "image/jpeg"},
		{"jpeg", "image/jpeg"},
		{"webp", "image/webp"},
		{"gif", "image/png"}, // default
		{"", "image/png"},    // default
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			if got := getOutputMimeType(tt.format); got != tt.want {
				t.Errorf("getOutputMimeType(%q) = %v, want %v", tt.format, got, tt.want)
			}
		})
	}
}
