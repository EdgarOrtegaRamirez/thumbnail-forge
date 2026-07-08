# Thumbnail Forge — AI Agent Guide

## Project Overview

Thumbnail Forge is a Go CLI tool that generates thumbnails for 40+ file types across 10 categories. Two-tier architecture: pure Go core (Tier 1) with optional CGo/external tool extensions (Tier 2). Includes Apple ecosystem support (HEIC/HEIF/AVIF, ICNS, ALAC, ProRes, iWork, DMG, IPA).

- **Module:** `github.com/EdgarOrtegaRamirez/thumbnail-forge`
- **Go version:** 1.25+
- **Lines of code:** ~5,300 (excluding fixtures)
- **Tests:** 61 pass, 2 skip, 0 fail
- **Benchmarks:** 55 across all handlers

## Architecture

### Core Components

1. **Detection Layer** (`internal/detect/detect.go`, 447 lines)
   - Identifies file types using magic bytes (first 512 bytes) and extensions
   - RIFF container disambiguation: checks bytes 8-12 for WEBP/WAVE/AVI
   - ftyp brand detection: ISO BMFF containers (HEIC/HEIF/AVIF/M4A/MP4/MOV) distinguished by brand string
   - ZIP content inspection: distinguishes Office docs (DOCX/XLSX/PPTX/iWork) and IPA from plain archives
   - DMG detection: checks for 'koly' magic trailer at end of file
   - Returns `FileInfo` with file type, MIME type, and metadata

2. **Handler Layer** (`internal/handlers/`, 9 handlers)
   - Each handler implements `CanHandle()` and `Generate()`
   - Returns `ThumbnailResult` with rendered image
   - Graceful degradation: placeholder thumbnails when external tools missing

3. **CLI Layer** (`cmd/`)
   - Cobra-based command structure: `generate`, `list`
   - Routes files to appropriate handlers via switch statement
   - Handles output formatting (file or terminal display)

4. **Terminal Output** (`internal/terminal/terminal.go`, 315 lines)
   - Protocol detection: Kitty graphics, iTerm2 inline images, Sixel, Unicode fallback
   - Base64-encoded PNG output for graphics protocols

### Handler Interface

```go
type Handler interface {
    CanHandle(info *models.FileInfo) bool
    Generate(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error)
}
```

### Handler Routing

Handlers are registered in `cmd/generate.go` via a switch on `info.FileType`:

```go
switch info.FileType {
case models.FileTypeImage:     handler = &handlers.ImageHandler{}
case models.FileTypeCode:      handler = &handlers.CodeHandler{}
case models.FileTypeText:      handler = &handlers.CodeHandler{}
case models.FileTypeMarkdown:  handler = &handlers.CodeHandler{}
case models.FileTypePDF:       handler = &handlers.PDFHandler{}
case models.FileTypeVideo:     handler = &handlers.VideoHandler{}
case models.FileTypeAudio:     handler = &handlers.AudioHandler{}
case models.FileTypeOffice:    handler = &handlers.OfficeHandler{}
case models.FileTypeArchive:   handler = &handlers.ArchiveHandler{}
case models.FileTypeDiskImage: handler = &handlers.DiskImageHandler{}
}
```

### File Type Detection Priority

1. **Magic bytes** — file signature (first 512 bytes)
2. **RIFF disambiguation** — bytes 8-12 distinguish WebP (image) / WAV (audio) / AVI (video)
3. **ftyp brand detection** — ISO BMFF containers distinguished by brand string: heic/heix/mif1 → HEIC, avif/avis → AVIF, M4A → audio, qt → MOV, default → MP4
4. **ZIP content inspection** — extension check distinguishes DOCX/XLSX/PPTX/iWork from plain ZIP, IPA from plain ZIP
5. **DMG koly trailer** — checks for 'koly' magic at end of file for Apple disk images
6. **Extension fallback** — when magic bytes are inconclusive

## Handlers

| Handler | File | Lines | Tier | Dependencies | Formats |
|---------|------|-------|------|-------------|---------|
| Image | `image.go` | 259 | 1+2 (Pure Go + External) | imaging, x/image, heif-convert | PNG, JPEG, GIF, WebP, BMP, TIFF, ICNS (pure Go), HEIC/HEIF/AVIF (heif-convert) |
| Code | `code.go` | 213 | 1 (Pure Go) | Chroma, fogleman/gg, Goldmark | Go, Python, JS, TS, Rust, Java, C, Ruby, etc. |
| Archive | `archive.go` | 232 | 1 (Pure Go) | stdlib archive/* | ZIP, TAR, TAR.GZ, 7z, RAR, IPA |
| DiskImage | `diskimage.go` | 82 | 1 (Pure Go) | stdlib | DMG, ISO, IMG (placeholder) |
| PDF | `pdf.go` | 92 | 2 (CGo) | go-fitz (MuPDF) | PDF |
| Video | `video.go` | 166 | 2 (External) | ffmpeg | MP4, MOV, MKV, WebM, AVI, ProRes |
| Audio | `audio.go` | 181 | 2 (External) | dhowden/tag, ffmpeg | MP3, WAV, FLAC, OGG, M4A, ALAC, AAC, AIFF |
| Office | `office.go` | 173 | 2 (External) | LibreOffice + MuPDF | DOCX, XLSX, PPTX, ODT, ODS, ODP, PAGES, NUMBERS, KEY |

### Exported Functions

- `handlers.LoadImage(path string)` — loads any supported image format (PNG, JPEG, GIF, WebP, BMP, TIFF, HEIC, HEIF, AVIF, ICNS)
- `handlers.ResizeImage(img image.Image, w, h int)` — Lanczos resize
- `handlers.loadHEIF(path string)` — HEIC/HEIF decoding via heif-convert (internal)
- `handlers.loadICNS(path string)` — Apple ICNS parsing via pure Go (internal)
- `handlers.loadPNG(path string)` — PNG loading helper (internal)
- These are exported for cross-package use by PDF and Video handlers

## Development

### Build

```bash
go build -o thumbnail-forge .
# or
make build
```

### Run Tests

```bash
# All tests (excludes bench fixtures that contain .c files)
go test ./internal/... ./cmd/...

# Verbose
go test ./internal/... -v

# Specific handler
go test ./internal/handlers/... -run TestImageHandler

# Race detector
go test -race ./internal/...
```

### Run Benchmarks

```bash
# Generate fixtures first
python3 generate_bench_fixtures.py

# Run benchmarks
go test ./internal/handlers/... -bench=. -benchmem -benchtime=1s -timeout=600s

# or
make bench
```

### Test Fixtures

- **`tests/fixtures/`** — basic sample files for unit tests (sample.go, sample.png, sample.mp4, etc.)
- **`tests/bench/`** — 58 benchmark fixtures across 8 categories, 3 sizes each (2.2 MB total)
  - Generated by `generate_bench_fixtures.py`
  - Contains `.go` and `.c` files — excluded from `go test ./...` via directory isolation

### Adding New Handlers

1. Create handler file in `internal/handlers/`
2. Implement the `Handler` interface (`CanHandle` + `Generate`)
3. Register in `cmd/generate.go` switch statement
4. Add unit tests in `internal/handlers/<name>_test.go`
5. Add benchmark entries in `internal/handlers/benchmark_test.go`
6. Add fixture generation in `generate_bench_fixtures.py`
7. Update README supported formats table
8. Update detect.go with magic bytes / extension mapping
9. Run `go test ./internal/... ./cmd/...`

### Code Style

- Follow Go conventions and idioms
- Export functions that are needed cross-package (e.g., `LoadImage`, `ResizeImage`)
- Handle errors gracefully — return placeholder thumbnails when external tools are missing
- Add comments for complex logic
- Prefer composition over inheritance

## Performance

See [BENCHMARKS.md](BENCHMARKS.md) for the full report.

### Performance Tiers

| Tier | Time Range | Categories |
|------|-----------|-----------|
| Ultra-fast | 0.003-0.6 ms | Archives, Terminal output |
| Fast | 1.9-7.0 ms | Images (small), Text, Markdown |
| Moderate | 13.7-53.6 ms | Code, Images (medium/large) |
| Slow | 118-287 ms | PDF, Video, Audio |
| Very slow | 711-1267 ms | Office, Code (large) |

### Known Bottlenecks

1. **Office documents** — LibreOffice startup dominates (700-900 ms). Consider daemon mode.
2. **Large code files** — Chroma tokenization creates one alloc per token. 2000 lines = 290 MB, 4M allocs.
3. **Video/audio** — ffmpeg process startup dominates (~185-290 ms). Consider persistent process.
4. **PDF** — MuPDF allocates 28 MB fixed. Efficient: constant time regardless of page count.

## Known Bugs Fixed

1. **RIFF container detection** — WAV files were misdetected as WebP. Fixed by checking bytes 8-12.
2. **ffmpeg single-image output** — Missing `-update 1` flag prevented waveform PNG creation.
3. **TIFF color model** — ffmpeg-generated TIFFs use unsupported color model. Fixed by generating via Go's `image/tiff` encoder with NRGBA.
4. **Office-as-archive detection** — DOCX/XLSX/PPTX detected as ZIP archives. Fixed by checking extension when ZIP magic bytes match.
5. **HEIC/MP4 ftyp overlap** — HEIC and MP4 both use ISO BMFF ftyp boxes with similar sizes. Fixed by brand-based detection (heic/heix/mif1 → HEIC, mp41/mp42/isom → MP4, qt → MOV) instead of box size matching.
6. **Audio waveform compositing** — ffmpeg `showwavespic` produces transparent PNGs. Fixed by compositing waveform onto background color after loading.
7. **ProRes timestamp parsing** — `parseTimestamp` didn't handle "1s" format with 's' suffix. Fixed by stripping trailing 's'.

## Security Notes

- Validate file paths before processing
- Sanitize user input (especially for shell commands passed to ffmpeg/LibreOffice)
- Set appropriate file permissions on output files
- Handle large files gracefully (streaming where possible)
- Be cautious with external tool execution

## Future Enhancements

- [x] ~~Add HEIC/AVIF image formats~~ (done — via heif-convert)
- [x] ~~Add Apple ICNS icon format~~ (done — pure Go parser)
- [x] ~~Add Apple ALAC/ProRes/iWork/DMG/IPA~~ (done)
- [ ] Add PSD (Photoshop) image format
- [ ] Implement batch processing mode
- [ ] Add video thumbnail caching
- [ ] Support remote files (HTTP/FTP)
- [ ] Add progress indicators
- [ ] Implement plugin system
- [ ] Add configuration file support
- [ ] Persistent ffmpeg/LibreOffice process for performance
- [ ] Stream/truncate code input before Chroma tokenization
- [ ] Native Sixel encoder (currently falls back to Unicode)
- [ ] DMG content rendering (mount and show actual contents instead of placeholder)
- [ ] HEIC/AVIF pure Go decoder (remove heif-convert dependency)
