package handlers

import (
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/detect"
	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

// BenchmarkResult holds timing for a single benchmark run
type BenchmarkResult struct {
	FileType   string
	Extension  string
	SizeLabel  string
	FileSize   int64
	Handler    string
	NsPerOp    int64
	AllocsPerOp int64
	BytesPerOp  int64
}

// benchmarkFile runs a benchmark for a single file
func benchmarkFile(b *testing.B, path string, width, height int) {
	info, err := detect.Detect(path)
	if err != nil {
		b.Fatalf("Detect() error for %s: %v", path, err)
	}

	opts := &models.ThumbnailOptions{
		Width:      width,
		Height:     height,
		Format:     "png",
		Quality:    85,
		Background: parseBgColor("#1e1e2e"),
		Theme:      "dracula",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var result *models.ThumbnailResult
		switch info.FileType {
		case models.FileTypeImage:
			h := &ImageHandler{}
			if h.CanHandle(info) {
				result, err = h.Generate(info, opts)
			}
		case models.FileTypeCode, models.FileTypeText, models.FileTypeMarkdown:
			h := &CodeHandler{}
			if h.CanHandle(info) {
				result, err = h.Generate(info, opts)
			}
		case models.FileTypePDF:
			h := &PDFHandler{}
			if h.CanHandle(info) {
				result, err = h.Generate(info, opts)
			}
		case models.FileTypeVideo:
			h := &VideoHandler{}
			if h.CanHandle(info) {
				result, err = h.Generate(info, opts)
			}
		case models.FileTypeAudio:
			h := &AudioHandler{}
			if h.CanHandle(info) {
				result, err = h.Generate(info, opts)
			}
		case models.FileTypeOffice:
			h := &OfficeHandler{}
			if h.CanHandle(info) {
				result, err = h.Generate(info, opts)
			}
		case models.FileTypeArchive:
			h := &ArchiveHandler{}
			if h.CanHandle(info) {
				result, err = h.Generate(info, opts)
			}
		}

		if err != nil {
			b.Fatalf("Generate() error for %s: %v", path, err)
		}
		if result == nil {
			b.Fatalf("No handler for %s (type: %s)", path, info.FileType)
		}
		_ = result
	}
}

func parseBgColor(hex string) color.Color {
	hex = strings.TrimPrefix(hex, "#")
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}

// fixturePath returns the path to a benchmark fixture
func fixturePath(category, name string) string {
	return filepath.Join("..", "..", "tests", "bench", category, name)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fileExistsOrDefault(path string, b *testing.B) {
	if !fileExists(path) {
		b.Skipf("fixture not found: %s", path)
	}
}

// === IMAGE BENCHMARKS ===

func BenchmarkImage_PNG_Small(b *testing.B) {
	p := fixturePath("images", "small.png")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkImage_PNG_Medium(b *testing.B) {
	p := fixturePath("images", "medium.png")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkImage_PNG_Large(b *testing.B) {
	p := fixturePath("images", "large.png")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkImage_JPEG_Small(b *testing.B) {
	p := fixturePath("images", "small.jpg")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkImage_JPEG_Medium(b *testing.B) {
	p := fixturePath("images", "medium.jpg")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkImage_JPEG_Large(b *testing.B) {
	p := fixturePath("images", "large.jpg")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkImage_GIF_Small(b *testing.B) {
	p := fixturePath("images", "small.gif")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkImage_GIF_Medium(b *testing.B) {
	p := fixturePath("images", "medium.gif")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkImage_WebP_Small(b *testing.B) {
	p := fixturePath("images", "small.webp")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkImage_WebP_Medium(b *testing.B) {
	p := fixturePath("images", "medium.webp")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkImage_BMP_Small(b *testing.B) {
	p := fixturePath("images", "small.bmp")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkImage_BMP_Medium(b *testing.B) {
	p := fixturePath("images", "medium.bmp")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkImage_TIFF_Small(b *testing.B) {
	p := fixturePath("images", "small.tiff")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkImage_TIFF_Medium(b *testing.B) {
	p := fixturePath("images", "medium.tiff")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

// === VIDEO BENCHMARKS ===

func BenchmarkVideo_MP4_Small(b *testing.B) {
	p := fixturePath("video", "small.mp4")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkVideo_MP4_Medium(b *testing.B) {
	p := fixturePath("video", "medium.mp4")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkVideo_MP4_Large(b *testing.B) {
	p := fixturePath("video", "large.mp4")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkVideo_MOV_Small(b *testing.B) {
	p := fixturePath("video", "small.mov")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkVideo_MKV_Small(b *testing.B) {
	p := fixturePath("video", "small.mkv")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkVideo_WebM_Small(b *testing.B) {
	p := fixturePath("video", "small.webm")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

// === AUDIO BENCHMARKS ===

func BenchmarkAudio_MP3_Small(b *testing.B) {
	p := fixturePath("audio", "small.mp3")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkAudio_MP3_Medium(b *testing.B) {
	p := fixturePath("audio", "medium.mp3")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkAudio_WAV_Small(b *testing.B) {
	p := fixturePath("audio", "small.wav")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkAudio_FLAC_Small(b *testing.B) {
	p := fixturePath("audio", "small.flac")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkAudio_OGG_Small(b *testing.B) {
	p := fixturePath("audio", "small.ogg")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

// === PDF BENCHMARKS ===

func BenchmarkPDF_Small(b *testing.B) {
	p := fixturePath("pdf", "small.pdf")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkPDF_Medium(b *testing.B) {
	p := fixturePath("pdf", "medium.pdf")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkPDF_Large(b *testing.B) {
	p := fixturePath("pdf", "large.pdf")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

// === OFFICE BENCHMARKS ===

func BenchmarkOffice_DOCX_Small(b *testing.B) {
	p := fixturePath("office", "small.docx")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkOffice_DOCX_Medium(b *testing.B) {
	p := fixturePath("office", "medium.docx")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkOffice_XLSX_Small(b *testing.B) {
	p := fixturePath("office", "small.xlsx")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkOffice_PPTX_Small(b *testing.B) {
	p := fixturePath("office", "small.pptx")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

// === CODE BENCHMARKS ===

func BenchmarkCode_Go_Small(b *testing.B) {
	p := fixturePath("code", "small.go")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkCode_Go_Medium(b *testing.B) {
	p := fixturePath("code", "medium.go")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkCode_Go_Large(b *testing.B) {
	p := fixturePath("code", "large.go")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkCode_Py_Small(b *testing.B) {
	p := fixturePath("code", "small.py")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkCode_JS_Small(b *testing.B) {
	p := fixturePath("code", "small.js")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkCode_TS_Small(b *testing.B) {
	p := fixturePath("code", "small.ts")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkCode_Rust_Small(b *testing.B) {
	p := fixturePath("code", "small.rs")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkCode_Java_Small(b *testing.B) {
	p := fixturePath("code", "small.java")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkCode_C_Small(b *testing.B) {
	p := fixturePath("code", "small.c")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

// === TEXT BENCHMARKS ===

func BenchmarkText_TXT_Small(b *testing.B) {
	p := fixturePath("text", "small.txt")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkText_TXT_Medium(b *testing.B) {
	p := fixturePath("text", "medium.txt")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkText_TXT_Large(b *testing.B) {
	p := fixturePath("text", "large.txt")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkText_JSON_Small(b *testing.B) {
	p := fixturePath("text", "small.json")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkText_JSON_Medium(b *testing.B) {
	p := fixturePath("text", "medium.json")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkText_XML_Small(b *testing.B) {
	p := fixturePath("text", "small.xml")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkText_YAML_Small(b *testing.B) {
	p := fixturePath("text", "small.yaml")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkText_CSV_Small(b *testing.B) {
	p := fixturePath("text", "small.csv")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkText_CSV_Medium(b *testing.B) {
	p := fixturePath("text", "medium.csv")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

// === MARKDOWN BENCHMARKS ===

func BenchmarkMarkdown_Small(b *testing.B) {
	p := fixturePath("markdown", "small.md")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkMarkdown_Medium(b *testing.B) {
	p := fixturePath("markdown", "medium.md")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkMarkdown_Large(b *testing.B) {
	p := fixturePath("markdown", "large.md")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

// === ARCHIVE BENCHMARKS ===

func BenchmarkArchive_ZIP_Small(b *testing.B) {
	p := fixturePath("archives", "small.zip")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkArchive_ZIP_Medium(b *testing.B) {
	p := fixturePath("archives", "medium.zip")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkArchive_ZIP_Large(b *testing.B) {
	p := fixturePath("archives", "large.zip")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkArchive_TAR_Small(b *testing.B) {
	p := fixturePath("archives", "small.tar")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

func BenchmarkArchive_TARGZ_Small(b *testing.B) {
	p := fixturePath("archives", "small.tar.gz")
	fileExistsOrDefault(p, b)
	benchmarkFile(b, p, 256, 256)
}

// === TERMINAL OUTPUT BENCHMARKS ===

func BenchmarkTerminal_Unicode(b *testing.B) {
	p := fixturePath("images", "small.png")
	fileExistsOrDefault(p, b)
	info, err := detect.Detect(p)
	if err != nil {
		b.Fatal(err)
	}
	opts := &models.ThumbnailOptions{
		Width: 256, Height: 256, Format: "png", Quality: 85,
		Background: parseBgColor("#1e1e2e"), Theme: "dracula",
	}
	h := &ImageHandler{}
	result, err := h.Generate(info, opts)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Simulate terminal output encoding
		_ = result.Image.Bounds()
	}
}
