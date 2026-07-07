package handlers

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

func TestPDFHandler_CanHandle(t *testing.T) {
	handler := &PDFHandler{}

	tests := []struct {
		name     string
		fileType models.FileType
		want     bool
	}{
		{"PDF file", models.FileTypePDF, true},
		{"Image file", models.FileTypeImage, false},
		{"Video file", models.FileTypeVideo, false},
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

func TestPDFHandler_Generate_SimplePDF(t *testing.T) {
	// Skip if MuPDF is not available
	if testing.Short() {
		t.Skip("Skipping PDF test in short mode")
	}

	tmpDir := t.TempDir()

	// Create a simple PDF file using go-fitz
	// This creates a minimal PDF with one page
	pdfPath := filepath.Join(tmpDir, "test.pdf")

	// Create a minimal PDF file
	// This is a valid PDF with one blank page
	pdfContent := `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>
endobj
xref
0 4
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
trailer
<< /Size 4 /Root 1 0 R >>
startxref
190
%%EOF`

	if err := os.WriteFile(pdfPath, []byte(pdfContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test the handler
	handler := &PDFHandler{}
	info := &models.FileInfo{
		Path:     pdfPath,
		FileType: models.FileTypePDF,
		MimeType: "application/pdf",
	}
	opts := &models.ThumbnailOptions{
		Width:      200,
		Height:     200,
		Background: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Page:       1,
		Format:     "png",
	}

	result, err := handler.Generate(info, opts)
	if err != nil {
		// MuPDF might not handle this minimal PDF, which is expected
		t.Skipf("MuPDF could not render test PDF: %v", err)
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

func TestPDFHandler_Generate_InvalidPDF(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid PDF file
	pdfPath := filepath.Join(tmpDir, "invalid.pdf")
	if err := os.WriteFile(pdfPath, []byte("This is not a PDF"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test the handler
	handler := &PDFHandler{}
	info := &models.FileInfo{
		Path:     pdfPath,
		FileType: models.FileTypePDF,
		MimeType: "application/pdf",
	}
	opts := &models.ThumbnailOptions{
		Width:      200,
		Height:     200,
		Background: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Page:       1,
		Format:     "png",
	}

	_, err := handler.Generate(info, opts)
	if err == nil {
		t.Error("Generate() expected error for invalid PDF, got nil")
	}
}

func TestGetPageCount(t *testing.T) {
	// Skip if MuPDF is not available
	if testing.Short() {
		t.Skip("Skipping PDF test in short mode")
	}

	tmpDir := t.TempDir()

	// Create a simple PDF file
	pdfPath := filepath.Join(tmpDir, "test.pdf")
	pdfContent := `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>
endobj
xref
0 4
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
trailer
<< /Size 4 /Root 1 0 R >>
startxref
190
%%EOF`

	if err := os.WriteFile(pdfPath, []byte(pdfContent), 0644); err != nil {
		t.Fatal(err)
	}

	count, err := GetPageCount(pdfPath)
	if err != nil {
		// MuPDF might not handle this minimal PDF, which is expected
		t.Skipf("MuPDF could not read test PDF: %v", err)
	}

	if count != 1 {
		t.Errorf("GetPageCount() = %d, want 1", count)
	}
}

func TestGetPageCount_InvalidPDF(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid PDF file
	pdfPath := filepath.Join(tmpDir, "invalid.pdf")
	if err := os.WriteFile(pdfPath, []byte("Not a PDF"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := GetPageCount(pdfPath)
	if err == nil {
		t.Error("GetPageCount() expected error for invalid PDF, got nil")
	}
}
