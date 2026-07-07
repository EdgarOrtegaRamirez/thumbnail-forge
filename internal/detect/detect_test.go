package detect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

func TestDetect_ImageFiles(t *testing.T) {
	// Create test directory
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  []byte
		ext      string
		wantType models.FileType
		wantMime string
	}{
		{
			name:     "JPEG file",
			content:  append([]byte{0xFF, 0xD8, 0xFF}, make([]byte, 100)...),
			ext:      ".jpg",
			wantType: models.FileTypeImage,
			wantMime: "image/jpeg",
		},
		{
			name:     "PNG file",
			content:  append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 100)...),
			ext:      ".png",
			wantType: models.FileTypeImage,
			wantMime: "image/png",
		},
		{
			name:     "GIF file",
			content:  append([]byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}, make([]byte, 100)...),
			ext:      ".gif",
			wantType: models.FileTypeImage,
			wantMime: "image/gif",
		},
		{
			name:     "BMP file",
			content:  append([]byte{0x42, 0x4D}, make([]byte, 100)...),
			ext:      ".bmp",
			wantType: models.FileTypeImage,
			wantMime: "image/bmp",
		},
		{
			name:     "WebP file",
			content:  append([]byte{0x52, 0x49, 0x46, 0x46}, []byte("WEBP")...),
			ext:      ".webp",
			wantType: models.FileTypeImage,
			wantMime: "image/webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, "test"+tt.ext)
			if err := os.WriteFile(path, tt.content, 0644); err != nil {
				t.Fatal(err)
			}

			info, err := Detect(path)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if info.FileType != tt.wantType {
				t.Errorf("Detect() FileType = %v, want %v", info.FileType, tt.wantType)
			}
			if info.MimeType != tt.wantMime {
				t.Errorf("Detect() MimeType = %v, want %v", info.MimeType, tt.wantMime)
			}
		})
	}
}

func TestDetect_VideoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  []byte
		ext      string
		wantType models.FileType
		wantMime string
	}{
		{
			name:     "MP4 file",
			content:  append([]byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70}, make([]byte, 100)...),
			ext:      ".mp4",
			wantType: models.FileTypeVideo,
			wantMime: "video/mp4",
		},
		{
			name:     "MKV file",
			content:  append([]byte{0x1A, 0x45, 0xDF, 0xA3}, make([]byte, 100)...),
			ext:      ".mkv",
			wantType: models.FileTypeVideo,
			wantMime: "video/x-matroska",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, "test"+tt.ext)
			if err := os.WriteFile(path, tt.content, 0644); err != nil {
				t.Fatal(err)
			}

			info, err := Detect(path)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if info.FileType != tt.wantType {
				t.Errorf("Detect() FileType = %v, want %v", info.FileType, tt.wantType)
			}
			if info.MimeType != tt.wantMime {
				t.Errorf("Detect() MimeType = %v, want %v", info.MimeType, tt.wantMime)
			}
		})
	}
}

func TestDetect_PDFFiles(t *testing.T) {
	tmpDir := t.TempDir()

	pdfContent := append([]byte{0x25, 0x50, 0x44, 0x46}, make([]byte, 100)...)
	path := filepath.Join(tmpDir, "test.pdf")
	if err := os.WriteFile(path, pdfContent, 0644); err != nil {
		t.Fatal(err)
	}

	info, err := Detect(path)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if info.FileType != models.FileTypePDF {
		t.Errorf("Detect() FileType = %v, want %v", info.FileType, models.FileTypePDF)
	}
	if info.MimeType != "application/pdf" {
		t.Errorf("Detect() MimeType = %v, want %v", info.MimeType, "application/pdf")
	}
}

func TestDetect_AudioFiles(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  []byte
		ext      string
		wantType models.FileType
		wantMime string
	}{
		{
			name:     "MP3 file",
			content:  append([]byte{0x49, 0x44, 0x33}, make([]byte, 100)...),
			ext:      ".mp3",
			wantType: models.FileTypeAudio,
			wantMime: "audio/mpeg",
		},
		{
			name:     "FLAC file",
			content:  append([]byte{0x66, 0x4C, 0x61, 0x63}, make([]byte, 100)...),
			ext:      ".flac",
			wantType: models.FileTypeAudio,
			wantMime: "audio/flac",
		},
		{
			name:     "OGG file",
			content:  append([]byte{0x4F, 0x67, 0x67, 0x53}, make([]byte, 100)...),
			ext:      ".ogg",
			wantType: models.FileTypeAudio,
			wantMime: "audio/ogg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, "test"+tt.ext)
			if err := os.WriteFile(path, tt.content, 0644); err != nil {
				t.Fatal(err)
			}

			info, err := Detect(path)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if info.FileType != tt.wantType {
				t.Errorf("Detect() FileType = %v, want %v", info.FileType, tt.wantType)
			}
			if info.MimeType != tt.wantMime {
				t.Errorf("Detect() MimeType = %v, want %v", info.MimeType, tt.wantMime)
			}
		})
	}
}

func TestDetect_ArchiveFiles(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  []byte
		ext      string
		wantType models.FileType
		wantMime string
	}{
		{
			name:     "ZIP file",
			content:  append([]byte{0x50, 0x4B, 0x03, 0x04}, make([]byte, 100)...),
			ext:      ".zip",
			wantType: models.FileTypeArchive,
			wantMime: "application/zip",
		},
		{
			name:     "7z file",
			content:  append([]byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}, make([]byte, 100)...),
			ext:      ".7z",
			wantType: models.FileTypeArchive,
			wantMime: "application/x-7z-compressed",
		},
		{
			name:     "RAR file",
			content:  append([]byte{0x52, 0x61, 0x72, 0x21}, make([]byte, 100)...),
			ext:      ".rar",
			wantType: models.FileTypeArchive,
			wantMime: "application/x-rar-compressed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, "test"+tt.ext)
			if err := os.WriteFile(path, tt.content, 0644); err != nil {
				t.Fatal(err)
			}

			info, err := Detect(path)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if info.FileType != tt.wantType {
				t.Errorf("Detect() FileType = %v, want %v", info.FileType, tt.wantType)
			}
			if info.MimeType != tt.wantMime {
				t.Errorf("Detect() MimeType = %v, want %v", info.MimeType, tt.wantMime)
			}
		})
	}
}

func TestDetect_CodeFiles(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		ext      string
		wantType models.FileType
	}{
		{"Go file", "package main", ".go", models.FileTypeCode},
		{"Python file", "#!/usr/bin/env python3", ".py", models.FileTypeCode},
		{"JavaScript file", "// JavaScript", ".js", models.FileTypeCode},
		{"TypeScript file", "// TypeScript", ".ts", models.FileTypeCode},
		{"Java file", "public class Main {}", ".java", models.FileTypeCode},
		{"C file", "#include <stdio.h>", ".c", models.FileTypeCode},
		{"Rust file", "fn main() {}", ".rs", models.FileTypeCode},
		{"Ruby file", "# Ruby", ".rb", models.FileTypeCode},
		{"PHP file", "<?php", ".php", models.FileTypeCode},
		{"Shell script", "#!/bin/bash", ".sh", models.FileTypeCode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, "test"+tt.ext)
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			info, err := Detect(path)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if info.FileType != tt.wantType {
				t.Errorf("Detect() FileType = %v, want %v", info.FileType, tt.wantType)
			}
		})
	}
}

func TestDetect_TextFiles(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		ext      string
		wantType models.FileType
	}{
		{"Text file", "Hello, World!", ".txt", models.FileTypeText},
		{"Log file", "2024-01-01 INFO Started", ".log", models.FileTypeText},
		{"JSON file", `{"key": "value"}`, ".json", models.FileTypeText},
		{"XML file", "<root><item/></root>", ".xml", models.FileTypeText},
		{"YAML file", "key: value", ".yaml", models.FileTypeText},
		{"TOML file", "[section]\nkey = \"value\"", ".toml", models.FileTypeText},
		{"CSV file", "name,age\nJohn,30", ".csv", models.FileTypeText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, "test"+tt.ext)
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			info, err := Detect(path)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if info.FileType != tt.wantType {
				t.Errorf("Detect() FileType = %v, want %v", info.FileType, tt.wantType)
			}
		})
	}
}

func TestDetect_MarkdownFiles(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		ext      string
		wantType models.FileType
	}{
		{"Markdown file", "# Title\n\nBody text", ".md", models.FileTypeMarkdown},
		{"Markdown (mdown)", "# Title\n\nBody text", ".mdown", models.FileTypeMarkdown},
		{"Markdown (mkd)", "# Title\n\nBody text", ".mkd", models.FileTypeMarkdown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, "test"+tt.ext)
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			info, err := Detect(path)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if info.FileType != tt.wantType {
				t.Errorf("Detect() FileType = %v, want %v", info.FileType, tt.wantType)
			}
		})
	}
}

func TestDetect_UnknownFileType(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with unknown extension
	path := filepath.Join(tmpDir, "test.xyz")
	if err := os.WriteFile(path, []byte("unknown content"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := Detect(path)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if info.FileType != models.FileTypeUnknown {
		t.Errorf("Detect() FileType = %v, want %v", info.FileType, models.FileTypeUnknown)
	}
}

func TestDetect_NonexistentFile(t *testing.T) {
	_, err := Detect("/nonexistent/file.txt")
	if err == nil {
		t.Error("Detect() expected error for nonexistent file, got nil")
	}
}

func TestFileType_String(t *testing.T) {
	tests := []struct {
		ft   models.FileType
		want string
	}{
		{models.FileTypeImage, "Image"},
		{models.FileTypeVideo, "Video"},
		{models.FileTypePDF, "PDF"},
		{models.FileTypeOffice, "Office Document"},
		{models.FileTypeAudio, "Audio"},
		{models.FileTypeCode, "Source Code"},
		{models.FileTypeText, "Text File"},
		{models.FileTypeMarkdown, "Markdown"},
		{models.FileTypeArchive, "Archive"},
		{models.FileTypeUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.ft.String(); got != tt.want {
				t.Errorf("FileType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
