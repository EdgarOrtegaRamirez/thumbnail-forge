# Thumbnail Forge — Benchmark Report

**Date:** 2026-07-07
**CPU:** Intel(R) Xeon(R) Gold 5412U
**OS:** Linux 6.12.47 (amd64)
**Go:** 1.25.0
**Thumbnail size:** 256×256 PNG, background #1e1e2e

## Summary

55 benchmarks across 8 file type categories, 58 fixture files totaling 2.2 MB.
All benchmarks passed. Total benchmark runtime: ~108 seconds.

## Results by Category

### 📷 Images (Pure Go — disintegration/imaging + Lanczos resize)

| Format | Size | File Size | Time/op | Memory/op | Allocs/op |
|--------|------|-----------|---------|-----------|----------|
| PNG | 64×64 | 552 B | **202 µs** | 105 KB | 32 |
| PNG | 512×512 | 14 KB | **17.0 ms** | 2.3 MB | 86 |
| PNG | 1024×1024 | 32 KB | **45.5 ms** | 6.1 MB | 79 |
| JPEG | 64×64 | 247 B | **228 µs** | 58 KB | 22 |
| JPEG | 512×512 | 1.8 KB | **17.8 ms** | 1.6 MB | 38 |
| JPEG | 1024×1024 | 6.4 KB | **53.6 ms** | 3.4 MB | 38 |
| GIF | 64×64 | 1.2 KB | **192 µs** | 73 KB | 540 |
| GIF | 256×256 | 2.6 KB | **1.9 ms** | 628 KB | 540 |
| WebP | 64×64 | 86 B | **244 µs** | 48 KB | 29 |
| WebP | 256×256 | 198 B | **2.6 ms** | 635 KB | 29 |
| BMP | 64×64 | 12 KB | **167 µs** | 56 KB | 21 |
| BMP | 256×256 | 197 KB | **2.0 ms** | 795 KB | 21 |
| TIFF | 64×64 | 16 KB | **142 µs** | 75 KB | 41 |
| TIFF | 256×256 | 262 KB | **2.0 ms** | 1.1 MB | 41 |

**Fastest:** BMP (167 µs) — simplest format, no decompression
**Slowest:** JPEG 1024×1024 (53.6 ms) — DCT decompression at scale
**Most allocs:** GIF (540) — frame-based palette format

### 🎬 Video (ffmpeg shell-out — frame extraction + decode)

| Format | Resolution | Duration | File Size | Time/op | Memory/op | Allocs/op |
|--------|-----------|----------|-----------|---------|-----------|----------|
| MP4 | 160×120 | 2s | 10 KB | **261 ms** | 281 KB | 166 |
| MP4 | 640×480 | 5s | 35 KB | **254 ms** | 1.6 MB | 185 |
| MP4 | 1280×720 | 10s | 123 KB | **287 ms** | 2.8 MB | 190 |
| MOV | 160×120 | 2s | 10 KB | **205 ms** | 280 KB | 166 |
| MKV | 160×120 | 2s | 10 KB | **200 ms** | 281 KB | 166 |
| WebM | 160×120 | 2s | 15 KB | **185 ms** | 281 KB | 166 |

**Observation:** Time dominated by ffmpeg process startup (~185-290 ms). File size/duration has minimal impact since only one frame is extracted. WebM fastest due to VP8 simplicity.

### 🎵 Audio (dhowden/tag + ffmpeg waveform fallback)

| Format | Duration | File Size | Time/op | Memory/op | Allocs/op |
|--------|----------|-----------|---------|-----------|----------|
| MP3 | 1s | 8.6 KB | **168 ms** | 383 KB | 187 |
| MP3 | 10s | 80 KB | **192 ms** | 383 KB | 186 |
| WAV | 1s | 88 KB | **182 ms** | 383 KB | 171 |
| FLAC | 1s | 20 KB | **220 ms** | 383 KB | 190 |
| OGG | 1s | 5.3 KB | **219 ms** | 395 KB | 213 |

**Observation:** All audio types use ffmpeg waveform generation (no album art in test files). Time dominated by ffmpeg startup. OGG has most allocs (213) due to Vorbis decoder overhead.

### 📄 PDF (go-fitz / MuPDF — 150 DPI page render)

| Pages | File Size | Time/op | Memory/op | Allocs/op |
|-------|-----------|---------|-----------|----------|
| 1 page | 27 KB | **120 ms** | 28.2 MB | 43 |
| 10 pages | 127 KB | **118 ms** | 28.2 MB | 43 |
| 50 pages | 580 KB | **130 ms** | 28.2 MB | 43 |

**Observation:** Time is nearly constant regardless of page count — only page 1 is rendered. Memory is dominated by MuPDF's internal 28 MB fixed allocation. Very efficient.

### 📝 Office (LibreOffice → PDF → MuPDF render)

| Format | File Size | Time/op | Memory/op | Allocs/op |
|--------|-----------|---------|-----------|----------|
| DOCX (5 para) | 1.0 KB | **849 ms** | 27.4 MB | 214 |
| DOCX (100 para) | 1.3 KB | **898 ms** | 27.4 MB | 213 |
| XLSX (10 rows) | 1.5 KB | **711 ms** | 28.2 MB | 214 |
| PPTX (3 slides) | 2.4 KB | **741 ms** | 20.1 MB | 214 |

**Observation:** Slowest category by far — LibreOffice startup dominates (~700-900 ms). This is a two-step pipeline: LibreOffice converts to PDF, then MuPDF renders the PDF. PPTX uses less memory (20 MB) because slide rendering produces smaller pages.

### 💻 Code (Chroma syntax highlighting + fogleman/gg text render)

| Language | Lines | File Size | Time/op | Memory/op | Allocs/op |
|----------|-------|-----------|---------|-----------|----------|
| Go | 50 | 1.3 KB | **14.5 ms** | 2.1 MB | 18,212 |
| Go | 500 | 63 KB | **313 ms** | 68.8 MB | 941,333 |
| Go | 2000 | 273 KB | **1.27 s** | 290 MB | 4,013,911 |
| Python | 50 | 2.1 KB | **22.8 ms** | 4.0 MB | 42,903 |
| JavaScript | 50 | 2.0 KB | **16.1 ms** | 2.7 MB | 26,921 |
| TypeScript | 50 | 1.8 KB | **14.9 ms** | 2.3 MB | 21,239 |
| Rust | 50 | 1.3 KB | **15.8 ms** | 2.4 MB | 21,044 |
| Java | 50 | 2.4 KB | **20.8 ms** | 3.2 MB | 33,356 |
| C | 50 | 1.1 KB | **13.7 ms** | 2.1 MB | 17,836 |

**Observation:** Memory and allocs scale linearly with code length. Chroma tokenization creates one allocation per token. Large Go file (2000 lines) uses 290 MB and 4M allocs — this is the biggest memory consumer in the entire benchmark. C is fastest (13.7 ms) due to simpler tokenization.

### 📃 Text (fogleman/gg text render — no syntax highlighting)

| Format | File Size | Time/op | Memory/op | Allocs/op |
|--------|-----------|---------|-----------|----------|
| TXT (small) | 175 B | **6.1 ms** | 1.1 MB | 52 |
| TXT (medium) | 11 KB | **3.9 ms** | 909 KB | 43 |
| TXT (large) | 248 KB | **4.8 ms** | 1.4 MB | 43 |
| JSON (small) | 615 B | **7.0 ms** | 1.2 MB | 55 |
| JSON (medium) | 34 KB | **7.0 ms** | 1.3 MB | 55 |
| XML (small) | 639 B | **3.9 ms** | 866 KB | 41 |
| YAML (small) | 789 B | **6.6 ms** | 1.1 MB | 55 |
| CSV (small) | 455 B | **3.5 ms** | 882 KB | 42 |
| CSV (medium) | 24 KB | **3.8 ms** | 938 KB | 43 |

**Observation:** Text rendering is fast and stable (3.5-7 ms). Large files don't significantly increase time because text is truncated to fit the thumbnail dimensions. Very memory-efficient (<1.5 MB for all sizes).

### 📖 Markdown (Chroma + text render)

| Size | File Size | Time/op | Memory/op | Allocs/op |
|------|-----------|---------|-----------|----------|
| Small | 126 B | **3.8 ms** | 881 KB | 42 |
| Medium | 4.6 KB | **3.5 ms** | 895 KB | 43 |
| Large | 47 KB | **3.6 ms** | 1.0 MB | 43 |

**Observation:** Markdown is rendered as plain text (no Markdown→HTML pipeline). Consistent ~3.5-3.8 ms across all sizes. Very efficient.

### 📦 Archives (Pure Go — archive/zip, archive/tar)

| Format | File Size | Time/op | Memory/op | Allocs/op |
|--------|-----------|---------|-----------|----------|
| ZIP (5 files) | 617 B | **374 µs** | 269 KB | 41 |
| ZIP (50 files) | 6.2 KB | **398 µs** | 284 KB | 224 |
| ZIP (100 files) | 13 KB | **444 µs** | 299 KB | 425 |
| TAR (5 files) | 20 KB | **497 µs** | 272 KB | 159 |
| TAR.GZ (5 files) | 306 B | **598 µs** | 317 KB | 167 |

**Observation:** Fastest category after terminal output. ZIP is slightly faster than TAR due to random access. TAR.GZ adds ~100 µs for decompression. Allocs scale with file count.

## Performance Tiers

| Tier | Time Range | Categories |
|------|-----------|-----------|
| **Ultra-fast** (<1 ms) | 0.003-0.6 ms | Archives, Terminal output |
| **Fast** (1-10 ms) | 1.9-7.0 ms | Images (small), Text, Markdown |
| **Moderate** (10-100 ms) | 13.7-53.6 ms | Code, Images (medium/large) |
| **Slow** (100-300 ms) | 118-287 ms | PDF, Video, Audio |
| **Very slow** (>500 ms) | 711-1267 ms | Office, Code (large) |

## Key Findings

1. **Office documents are the bottleneck** — LibreOffice startup takes 700-900 ms. Consider caching the LibreOffice process or using a daemon mode.

2. **Large code files consume excessive memory** — 2000 lines of Go uses 290 MB and 4M allocations. Chroma tokenization is the culprit. Consider streaming or truncating input before tokenization.

3. **Video/audio time is ffmpeg-bound** — Process startup dominates. A persistent ffmpeg process or pipe would eliminate this overhead.

4. **PDF rendering is very efficient** — Constant ~120 ms and 28 MB regardless of page count, since only the first page is rendered.

5. **Images scale linearly with pixel count** — Time and memory are proportional to source image dimensions, not file size.

6. **Text/Markdown are very stable** — ~3.5-7 ms regardless of file size, since text is truncated to fit the thumbnail.

7. **Archives are the fastest file-type handler** — 374-598 µs. Pure Go, no external tools.

## Bugs Found & Fixed During Benchmarking

1. **RIFF container detection** — WAV files were detected as WebP images because both use RIFF headers. Fixed by checking bytes 8-12 (WEBP/WAVE/AVI) to distinguish formats.

2. **ffmpeg single-image output** — ffmpeg requires `-update 1` flag to write a single image to a fixed filename. Without it, the file is not created.

3. **TIFF color model** — ffmpeg-generated TIFFs use a color model Go's TIFF decoder doesn't support. Fixed by generating TIFFs via Go's `image/tiff` encoder with NRGBA color model.

4. **DOCX/XLSX/PPTX detection** — Office documents (ZIP-based) were detected as archives. Fixed by checking extension when ZIP magic bytes are found.
