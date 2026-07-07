package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "thumbnail-forge",
	Short: "Generate thumbnails for almost any file type",
	Long: `Thumbnail Forge is a CLI tool that generates thumbnails for almost any file type.

Supported file types:
  • Images: JPEG, PNG, GIF, BMP, TIFF, WebP, HEIF, AVIF, SVG
  • Video: MP4, MOV, AVI, MKV, WebM, FLV (requires ffmpeg)
  • PDF: All PDF documents (requires MuPDF)
  • Office: DOCX, XLSX, PPTX, ODT, etc. (requires LibreOffice)
  • Audio: MP3, WAV, FLAC, OGG, M4A (album art extraction)
  • Code: Go, Python, JavaScript, TypeScript, Java, C, Rust, etc.
  • Text: TXT, LOG, JSON, XML, YAML, TOML, CSV
  • Markdown: MD, MDOWN, MKD
  • Archives: ZIP, 7z, RAR, TAR (contents listing)

Examples:
  thumbnail-forge generate photo.jpg
  thumbnail-forge generate video.mp4 --timestamp 5s
  thumbnail-forge generate document.pdf --page 2
  thumbnail-forge generate code.go --width 512 --height 384
  thumbnail-forge generate photo.png --terminal
  thumbnail-forge list`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("thumbnail-forge %s (commit: %s, built: %s)\n", version, commit, date)
	},
}
