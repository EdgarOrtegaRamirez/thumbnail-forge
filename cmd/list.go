package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List supported file types",
	Long:  `List all supported file types and their categories.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Thumbnail Forge - Supported File Types")
		fmt.Println("======================================")
		fmt.Println()

		fmt.Println("📷 Images")
		fmt.Println("  Pure Go: JPEG, PNG, GIF, BMP, TIFF, WebP")
		fmt.Println("  Apple:   HEIC, HEIF, AVIF (via heif-convert), ICNS (pure Go parser)")
		fmt.Println()

		fmt.Println("🎬 Video (requires ffmpeg)")
		fmt.Println("  MP4, MOV, AVI, MKV, WebM, FLV, WMV, M4V")
		fmt.Println("  Apple: ProRes (via ffmpeg, detected as MOV)")
		fmt.Println()

		fmt.Println("📄 PDF (requires MuPDF)")
		fmt.Println("  All PDF documents")
		fmt.Println()

		fmt.Println("📝 Office Documents (requires LibreOffice)")
		fmt.Println("  Microsoft: DOCX, XLSX, PPTX, DOC, XLS, PPT")
		fmt.Println("  OpenDoc:   ODT, ODS, ODP")
		fmt.Println("  Apple iWork: PAGES, NUMBERS, KEY")
		fmt.Println()

		fmt.Println("🎵 Audio (requires ffmpeg)")
		fmt.Println("  MP3, WAV, FLAC, OGG, M4A, AAC, WMA")
		fmt.Println("  Apple: ALAC (.alac), AIFF (.aiff, .aif)")
		fmt.Println("  (Album art extraction or waveform visualization)")
		fmt.Println()

		fmt.Println("💻 Source Code (Pure Go)")
		fmt.Println("  Go, Python, JavaScript, TypeScript, JSX, TSX")
		fmt.Println("  Java, C, C++, Header files, Rust, Ruby, PHP")
		fmt.Println("  Swift, Kotlin, Scala, Clojure, Shell scripts")
		fmt.Println()

		fmt.Println("📃 Text Files (Pure Go)")
		fmt.Println("  TXT, LOG, JSON, XML, YAML, YML, TOML")
		fmt.Println("  INI, CFG, CONF, CSV, TSV")
		fmt.Println()

		fmt.Println("📖 Markdown (Pure Go)")
		fmt.Println("  MD, MDOWN, MKD")
		fmt.Println()

		fmt.Println("📦 Archives (Pure Go for ZIP)")
		fmt.Println("  ZIP, 7z, RAR, TAR, GZ, BZ2, XZ")
		fmt.Println("  Apple: IPA (iOS app packages, ZIP-based)")
		fmt.Println("  (Contents listing with folder icon)")
		fmt.Println()

		fmt.Println("💿 Disk Images")
		fmt.Println("  Apple: DMG, IMG")
		fmt.Println("  Generic: ISO")
		fmt.Println("  (Placeholder with disk icon)")
		fmt.Println()

		fmt.Println("❓ Unknown Types")
		fmt.Println("  Generic file icon with extension badge")
		fmt.Println()

		fmt.Println("Dependencies:")
		fmt.Println("  • ffmpeg — Required for video/audio processing")
		fmt.Println("  • MuPDF (libmupdf-dev) — Optional, for PDF rendering")
		fmt.Println("  • LibreOffice — Optional, for office document conversion")
		fmt.Println("  • heif-convert (libheif-examples) — Optional, for HEIC/HEIF/AVIF")
		fmt.Println("  • libvips-dev — Optional, for extended image format support")
	},
}
