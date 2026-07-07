# Thumbnail Forge

A Go CLI tool that generates thumbnails for virtually any file type — images, videos, PDFs, office documents, audio, code, text, archives, and more.

## Features

- **Multi-format support**: 30+ file types across 8 categories
- **Smart detection**: Magic bytes + extension-based file type identification
- **High quality**: Lanczos resampling, Chroma syntax highlighting, MuPDF rendering
- **Terminal output**: Display thumbnails inline (Kitty, iTerm2, Sixel, Unicode fallback)
- **Customizable**: Configurable dimensions, colors, themes, and output formats
- **Two-tier architecture**: Pure Go core with optional system-tool extensions
- **Benchmarked**: 55 benchmarks across all handlers with full performance data

## Installation

### Prerequisites

- **Go 1.25+** (required)
- **ffmpeg** — video frame extraction, audio waveform generation
- **libmupdf-dev** — PDF rendering via go-fitz (CGo)
- **LibreOffice** — office document conversion (DOCX/XLSX/PPTX)

### Build

```bash
git clone https://github.com/EdgarOrtegaRamirez/thumbnail-forge.git
cd thumbnail-forge

# Install system dependencies (Debian/Ubuntu)
sudo apt-get install -y ffmpeg libmupdf-dev libreoffice-core libreoffice-writer libreoffice-calc libreoffice-impress

# Build
go build -o thumbnail-forge .

# Optional: install system-wide
sudo mv thumbnail-forge /usr/local/bin/
```

### Make

```bash
make build          # Build binary
make test           # Run all tests
make test-race      # Run tests with race detector
make bench          # Run benchmarks
make fixtures       # Generate benchmark fixtures
make install         # Install to /usr/local/bin
```

## Usage

```bash
# Basic usage — generates thumbnail.png in current directory
thumbforge generate <file>

# Specify dimensions
thumbforge generate <file> --width 400 --height 300

# Output to terminal (Kitty/iTerm2/Sixel/Unicode)
thumbforge generate <file> --terminal

# Output format and quality
thumbforge generate <file> --format jpg --quality 90

# Custom background color (for transparent/letterboxed images)
thumbforge generate <file> --background "#ff0000"

# Extract specific video frame (seconds, MM:SS, or HH:MM:SS)
thumbforge generate video.mp4 --timestamp 00:00:05

# Render specific PDF page
thumbforge generate document.pdf --page 3

# Code syntax theme
thumbforge generate main.go --theme monokai

# List all supported file types
thumbforge list
```

## Supported File Types

### Tier 1 — Pure Go (no external dependencies)

| Category | Formats | Handler |
|----------|---------|---------|
| Images | PNG, JPEG, GIF, WebP, BMP, TIFF | `image.go` |
| Code | Go, Python, JavaScript, TypeScript, Rust, Java, C, C++, Ruby, and more | `code.go` |
| Text | TXT, JSON, XML, YAML, CSV, log files | `code.go` |
| Markdown | MD | `code.go` |
| Archives | ZIP, TAR, TAR.GZ | `archive.go` |

### Tier 2 — External tools required

| Category | Formats | Dependency | Handler |
|----------|---------|------------|---------|
| PDF | PDF | MuPDF (libmupdf-dev) | `pdf.go` |
| Video | MP4, MOV, MKV, WebM, AVI | ffmpeg | `video.go` |
| Audio | MP3, WAV, FLAC, OGG | ffmpeg + dhowden/tag | `audio.go` |
| Office | DOCX, XLSX, PPTX, ODT, ODS, ODP | LibreOffice | `office.go` |

Handlers gracefully degrade: if an external tool is missing, a placeholder thumbnail is generated instead of failing.

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--width` | 256 | Thumbnail width in pixels |
| `--height` | 256 | Thumbnail height in pixels |
| `--output` | (temp file) | Output file path |
| `--format` | png | Output format: png, jpg |
| `--quality` | 85 | JPEG quality (1-100) |
| `--background` | #1e1e2e | Background color (hex) |
| `--timestamp` | 1s | Video frame timestamp (s, MM:SS, HH:MM:SS) |
| `--page` | 1 | PDF page number |
| `--theme` | dracula | Chroma syntax theme |
| `--terminal` | false | Display thumbnail in terminal |
| `--freedesktop` | false | Generate Freedesktop.org-compliant thumbnail |

### Code Themes

Available via [Chroma](https://github.com/alecthomas/chroma):
dracula (default), monokai, github, solarized-dark, solarized-light, swapoff, dracula, and [200+ more](https://swapoff.org/chroma/playground/).

## Architecture

```
File → Detect (magic bytes + ext) → Handler → Resize (Lanczos) → Composite (bg) → Output (file/terminal)
```

### Two-Tier Design

- **Tier 1 (Pure Go)**: Image, code, text, archive handlers work without any system dependencies. Uses `disintegration/imaging`, `alecthomas/chroma`, `fogleman/gg`, and stdlib `archive/*`.
- **Tier 2 (CGo/External)**: PDF uses `gen2brain/go-fitz` (MuPDF CGo bindings). Video/audio shell out to `ffmpeg`. Office shells out to `LibreOffice` → PDF → MuPDF.

### Handler Interface

```go
type Handler interface {
    CanHandle(info *models.FileInfo) bool
    Generate(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error)
}
```

### File Type Detection

Priority order:
1. **Magic bytes** — file signature (first 512 bytes)
2. **RIFF container disambiguation** — checks bytes 8-12 for WEBP/WAVE/AVI
3. **ZIP content inspection** — distinguishes Office docs (DOCX/XLSX/PPTX) from plain archives
4. **File extension** — fallback when magic bytes are inconclusive

### Project Structure

```
thumbnail-forge/
├── main.go                         # Entry point
├── cmd/                            # CLI commands (Cobra)
│   ├── root.go                     # Root command + version
│   ├── generate.go                 # Generate subcommand + handler routing
│   └── list.go                     # List supported types
├── internal/
│   ├── models/models.go            # FileType enum, ThumbnailOptions, Handler interface
│   ├── detect/
│   │   ├── detect.go               # Magic byte + extension detection
│   │   └── detect_test.go          # Detection tests (11 tests)
│   ├── handlers/
│   │   ├── image.go                # PNG/JPEG/GIF/WebP/BMP/TIFF handler
│   │   ├── code.go                 # Code/text/markdown handler (Chroma)
│   │   ├── pdf.go                  # PDF handler (go-fitz/MuPDF)
│   │   ├── video.go                # Video handler (ffmpeg)
│   │   ├── audio.go                # Audio handler (tag + ffmpeg waveform)
│   │   ├── office.go               # Office handler (LibreOffice → PDF)
│   │   ├── archive.go              # ZIP/TAR/TAR.GZ handler
│   │   ├── benchmark_test.go       # 55 benchmarks across all handlers
│   │   └── *_test.go               # Unit tests per handler
│   └── terminal/
│       └── terminal.go            # Kitty/iTerm2/Sixel/Unicode output
├── tests/
│   ├── fixtures/                   # Basic test files (sample.*)
│   └── bench/                      # Benchmark fixtures (58 files, 8 categories)
├── .github/workflows/ci.yml       # CI: test, lint, release
├── Makefile                        # build, test, bench, fixtures, install
├── BENCHMARKS.md                   # Full benchmark report
├── AGENTS.md                       # AI agent guide
└── go.mod                          # Module: github.com/EdgarOrtegaRamirez/thumbnail-forge
```

## Benchmarks

55 benchmarks across all handlers. See [BENCHMARKS.md](BENCHMARKS.md) for the full report.

### Performance Summary

| Tier | Time Range | Categories |
|------|-----------|-----------|
| Ultra-fast | 0.003-0.6 ms | Archives, Terminal output |
| Fast | 1.9-7.0 ms | Images (small), Text, Markdown |
| Moderate | 13.7-53.6 ms | Code, Images (medium/large) |
| Slow | 118-287 ms | PDF, Video, Audio |
| Very slow | 711-1267 ms | Office, Code (large 2000 lines) |

Run benchmarks:
```bash
make bench          # all benchmarks
make fixtures       # generate test fixtures first
```

## Development

### Adding New Handlers

1. Create a handler file in `internal/handlers/`
2. Implement the `Handler` interface (`CanHandle` + `Generate`)
3. Register in `cmd/generate.go` switch statement
4. Add tests in `internal/handlers/`
5. Add benchmark entries in `benchmark_test.go`
6. Update README supported formats table
7. Run `go test ./internal/... ./cmd/...`

### Testing

```bash
# Run all tests (excludes bench fixtures)
go test ./internal/... ./cmd/...

# Verbose
go test ./internal/... -v

# Specific handler
go test ./internal/handlers/... -run TestImageHandler

# Benchmarks
go test ./internal/handlers/... -bench=. -benchmem

# Race detector
go test -race ./internal/...
```

### Test Status

- **61 tests pass**, 2 skip, 0 fail
- **55 benchmarks** across all handlers
- **58 benchmark fixtures** (2.2 MB total)

## Dependencies

### Go Modules

| Module | Purpose |
|--------|---------|
| [spf13/cobra](https://github.com/spf13/cobra) | CLI framework |
| [disintegration/imaging](https://github.com/disintegration/imaging) | Image loading + Lanczos resize |
| [golang.org/x/image/webp](https://golang.org/x/image) | WebP decoding |
| [golang.org/x/image/tiff](https://golang.org/x/image) | TIFF encoding/decoding |
| [alecthomas/chroma/v2](https://github.com/alecthomas/chroma) | Syntax highlighting |
| [fogleman/gg](https://github.com/fogleman/gg) | 2D graphics / text rendering |
| [gen2brain/go-fitz](https://github.com/gen2brain/go-fitz) | MuPDF CGo bindings (PDF) |
| [dhowden/tag](https://github.com/dhowden/tag) | Audio metadata (album art) |
| [yuin/goldmark](https://github.com/yuin/goldmark) | Markdown parsing |

### System Packages

| Package | Required for | Debian/Ubuntu install |
|---------|-------------|----------------------|
| ffmpeg | Video, Audio | `apt-get install ffmpeg` |
| libmupdf-dev | PDF | `apt-get install libmupdf-dev` |
| LibreOffice | Office docs | `apt-get install libreoffice-core libreoffice-writer libreoffice-calc libreoffice-impress` |

## License

MIT License — see [LICENSE](LICENSE).

## Contributing

Contributions welcome! Please submit a Pull Request. See [AGENTS.md](AGENTS.md) for the full development guide.
