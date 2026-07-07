package terminal

import (
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

func TestDetectProtocol(t *testing.T) {
	// Save original environment
	origKitty := os.Getenv("KITTY_WINDOW_ID")
	origTerm := os.Getenv("TERM_PROGRAM")
	origSixel := os.Getenv("SIXEL_SUPPORT")

	// Restore environment after test
	defer func() {
		os.Setenv("KITTY_WINDOW_ID", origKitty)
		os.Setenv("TERM_PROGRAM", origTerm)
		os.Setenv("SIXEL_SUPPORT", origSixel)
	}()

	tests := []struct {
		name     string
		env      map[string]string
		expected Protocol
	}{
		{
			name:     "Kitty terminal",
			env:      map[string]string{"KITTY_WINDOW_ID": "12345"},
			expected: ProtocolKitty,
		},
		{
			name:     "iTerm2 terminal",
			env:      map[string]string{"TERM_PROGRAM": "iTerm.app"},
			expected: ProtocolITerm2,
		},
		{
			name:     "WezTerm terminal",
			env:      map[string]string{"TERM_PROGRAM": "WezTerm"},
			expected: ProtocolITerm2,
		},
		{
			name:     "Sixel support",
			env:      map[string]string{"SIXEL_SUPPORT": "1"},
			expected: ProtocolSixel,
		},
		{
			name:     "No special terminal",
			env:      map[string]string{},
			expected: ProtocolUnicode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables
			os.Unsetenv("KITTY_WINDOW_ID")
			os.Unsetenv("TERM_PROGRAM")
			os.Unsetenv("SIXEL_SUPPORT")

			// Set test environment variables
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			// Test detection
			protocol := DetectProtocol()
			if protocol != tt.expected {
				t.Errorf("DetectProtocol() = %v, want %v", protocol, tt.expected)
			}
		})
	}
}

func TestNewOutput(t *testing.T) {
	output := NewOutput(ProtocolAuto)
	if output.protocol != ProtocolAuto {
		t.Errorf("NewOutput() protocol = %v, want %v", output.protocol, ProtocolAuto)
	}

	output = NewOutput(ProtocolKitty)
	if output.protocol != ProtocolKitty {
		t.Errorf("NewOutput() protocol = %v, want %v", output.protocol, ProtocolKitty)
	}
}

func TestDisplay_NilImage(t *testing.T) {
	output := NewOutput(ProtocolUnicode)
	result := &models.ThumbnailResult{
		Width:  100,
		Height: 100,
	}

	err := output.Display(nil, result)
	if err == nil {
		t.Error("Display() expected error for nil image, got nil")
	}
}

func TestDisplay_Unicode(t *testing.T) {
	// Create a small test image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 25), G: uint8(y * 25), B: 128, A: 255})
		}
	}

	output := NewOutput(ProtocolUnicode)
	result := &models.ThumbnailResult{
		Width:  10,
		Height: 10,
	}

	// This should not panic
	err := output.Display(img, result)
	if err != nil {
		t.Errorf("Display() error = %v", err)
	}
}

func TestEncodeBase64(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"Empty", []byte{}, ""},
		{"A", []byte("A"), "QQ=="},
		{"AB", []byte("AB"), "QUI="},
		{"ABC", []byte("ABC"), "QUJD"},
		{"ABCD", []byte("ABCD"), "QUJDRA=="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encodeBase64(tt.input)
			if result != tt.expected {
				t.Errorf("encodeBase64() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSplitString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		size      int
		expected  []string
	}{
		{"Empty", "", 5, []string{}},
		{"Short", "abc", 5, []string{"abc"}},
		{"Exact", "abcde", 5, []string{"abcde"}},
		{"Split", "abcdefgh", 3, []string{"abc", "def", "gh"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitString(tt.input, tt.size)
			if len(result) != len(tt.expected) {
				t.Errorf("splitString() returned %d chunks, want %d", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitString()[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestGeneratePlaceholder(t *testing.T) {
	img := GeneratePlaceholder(100, 50)
	if img == nil {
		t.Fatal("GeneratePlaceholder() returned nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 50 {
		t.Errorf("GeneratePlaceholder() dimensions = %dx%d, want 100x50", bounds.Dx(), bounds.Dy())
	}
}

func TestEncodePNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	var buf []byte
	err := encodePNG(img, &buf)
	if err != nil {
		t.Fatalf("encodePNG() error = %v", err)
	}

	if len(buf) == 0 {
		t.Error("encodePNG() returned empty buffer")
	}

	// Check PNG header
	if len(buf) < 8 {
		t.Error("encodePNG() buffer too small for PNG header")
	}

	// PNG magic bytes
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i := 0; i < 8; i++ {
		if buf[i] != pngHeader[i] {
			t.Errorf("encodePNG() buffer[%d] = %x, want %x", i, buf[i], pngHeader[i])
			break
		}
	}
}
