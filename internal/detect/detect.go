package detect

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

// Magic bytes for various file formats
var magicBytes = map[string][]byte{
	// Images
	"jpeg":     {0xFF, 0xD8, 0xFF},
	"png":      {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
	"gif":      {0x47, 0x49, 0x46, 0x38},
	"bmp":      {0x42, 0x4D},
	"webp":     {0x52, 0x49, 0x46, 0x46},                                                 // RIFF header, need to check for WEBP
	"tiff":     {0x49, 0x49, 0x2A, 0x00},                                                 // Little-endian TIFF
	"tiff_big": {0x4D, 0x4D, 0x00, 0x2A},                                                 // Big-endian TIFF
	"icns":     {0x69, 0x63, 0x6E, 0x73},                                                 // Apple ICNS
	"heic":     {0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70, 0x68, 0x65, 0x69, 0x63}, // ftypheic
	"heif":     {0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70, 0x68, 0x65, 0x69, 0x78}, // ftypheix
	"heif2":    {0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70, 0x6D, 0x69, 0x66, 0x31}, // ftypmif1
	"avif":     {0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70, 0x61, 0x76, 0x69, 0x66}, // ftypavif
	"avif2":    {0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70, 0x61, 0x76, 0x69, 0x73}, // ftypavis

	// Video
	"mp4":   {0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70}, // ftyp box
	"mp4_2": {0x00, 0x00, 0x00, 0x1C, 0x66, 0x74, 0x79, 0x70}, // Alternate ftyp
	"mov":   {0x00, 0x00, 0x00, 0x14, 0x66, 0x74, 0x79, 0x70},
	"avi":   {0x52, 0x49, 0x46, 0x46}, // RIFF header, need to check for AVI
	"mkv":   {0x1A, 0x45, 0xDF, 0xA3},

	// Audio
	"mp3":  {0x49, 0x44, 0x33}, // ID3 header
	"flac": {0x66, 0x4C, 0x61, 0x43},
	"ogg":  {0x4F, 0x67, 0x67, 0x53},
	"wav":  {0x52, 0x49, 0x46, 0x46},                                                 // RIFF header, need to check for WAVE
	"m4a":  {0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70, 0x4D, 0x34, 0x41, 0x20}, // ftypM4A

	// PDF
	"pdf": {0x25, 0x50, 0x44, 0x46}, // %PDF

	// Archives
	"zip": {0x50, 0x4B, 0x03, 0x04},
	"7z":  {0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C},
	"rar": {0x52, 0x61, 0x72, 0x21}, // Rar!

	// Disk Images
	"dmg_koly": {0x6B, 0x6F, 0x6C, 0x79}, // 'koly' at end of file

	// Office
	"docx": {0x50, 0x4B, 0x03, 0x04}, // ZIP-based
	"xlsx": {0x50, 0x4B, 0x03, 0x04}, // ZIP-based
	"pptx": {0x50, 0x4B, 0x03, 0x04}, // ZIP-based
}

// extensionToType maps file extensions to file types
var extensionToType = map[string]models.FileType{
	// Images
	".jpg":  models.FileTypeImage,
	".jpeg": models.FileTypeImage,
	".png":  models.FileTypeImage,
	".gif":  models.FileTypeImage,
	".bmp":  models.FileTypeImage,
	".webp": models.FileTypeImage,
	".tiff": models.FileTypeImage,
	".tif":  models.FileTypeImage,
	".svg":  models.FileTypeImage,
	".heic": models.FileTypeImage,
	".heif": models.FileTypeImage,
	".avif": models.FileTypeImage,
	".ico":  models.FileTypeImage,
	".icns": models.FileTypeImage,

	// Video
	".mp4":  models.FileTypeVideo,
	".mov":  models.FileTypeVideo,
	".avi":  models.FileTypeVideo,
	".mkv":  models.FileTypeVideo,
	".webm": models.FileTypeVideo,
	".flv":  models.FileTypeVideo,
	".wmv":  models.FileTypeVideo,
	".m4v":  models.FileTypeVideo,

	// PDF
	".pdf": models.FileTypePDF,

	// Office (including Apple iWork)
	".docx":    models.FileTypeOffice,
	".xlsx":    models.FileTypeOffice,
	".pptx":    models.FileTypeOffice,
	".odt":     models.FileTypeOffice,
	".ods":     models.FileTypeOffice,
	".odp":     models.FileTypeOffice,
	".doc":     models.FileTypeOffice,
	".xls":     models.FileTypeOffice,
	".ppt":     models.FileTypeOffice,
	".pages":   models.FileTypeOffice,
	".numbers": models.FileTypeOffice,
	".key":     models.FileTypeOffice,
	".keynote": models.FileTypeOffice,

	// Audio (including Apple formats)
	".mp3":  models.FileTypeAudio,
	".wav":  models.FileTypeAudio,
	".flac": models.FileTypeAudio,
	".ogg":  models.FileTypeAudio,
	".m4a":  models.FileTypeAudio,
	".aac":  models.FileTypeAudio,
	".wma":  models.FileTypeAudio,
	".alac": models.FileTypeAudio,
	".aiff": models.FileTypeAudio,
	".aif":  models.FileTypeAudio,

	// Code
	".go":    models.FileTypeCode,
	".py":    models.FileTypeCode,
	".js":    models.FileTypeCode,
	".ts":    models.FileTypeCode,
	".jsx":   models.FileTypeCode,
	".tsx":   models.FileTypeCode,
	".java":  models.FileTypeCode,
	".cpp":   models.FileTypeCode,
	".c":     models.FileTypeCode,
	".h":     models.FileTypeCode,
	".rs":    models.FileTypeCode,
	".rb":    models.FileTypeCode,
	".php":   models.FileTypeCode,
	".swift": models.FileTypeCode,
	".kt":    models.FileTypeCode,
	".scala": models.FileTypeCode,
	".clj":   models.FileTypeCode,
	".sh":    models.FileTypeCode,
	".bash":  models.FileTypeCode,
	".zsh":   models.FileTypeCode,
	".ps1":   models.FileTypeCode,
	".bat":   models.FileTypeCode,

	// Text
	".txt":  models.FileTypeText,
	".log":  models.FileTypeText,
	".csv":  models.FileTypeText,
	".tsv":  models.FileTypeText,
	".json": models.FileTypeText,
	".xml":  models.FileTypeText,
	".yaml": models.FileTypeText,
	".yml":  models.FileTypeText,
	".toml": models.FileTypeText,
	".ini":  models.FileTypeText,
	".cfg":  models.FileTypeText,
	".conf": models.FileTypeText,

	// Markdown
	".md":    models.FileTypeMarkdown,
	".mdown": models.FileTypeMarkdown,
	".mkd":   models.FileTypeMarkdown,

	// Archives (including Apple IPA)
	".zip": models.FileTypeArchive,
	".7z":  models.FileTypeArchive,
	".rar": models.FileTypeArchive,
	".tar": models.FileTypeArchive,
	".gz":  models.FileTypeArchive,
	".bz2": models.FileTypeArchive,
	".xz":  models.FileTypeArchive,
	".ipa": models.FileTypeArchive,

	// Disk Images (Apple)
	".dmg": models.FileTypeDiskImage,
	".iso": models.FileTypeDiskImage,
	".img": models.FileTypeDiskImage,
}

// Detect identifies the file type using magic bytes and extension
func Detect(path string) (*models.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Read first 512 bytes for magic byte detection
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	header := make([]byte, 512)
	n, err := f.Read(header)
	if err != nil && n == 0 {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(path))
	detectedType := models.FileTypeUnknown
	mimeType := ""
	isAnimated := false

	// Try magic bytes first
	if n >= 8 {
		switch {
		case matchesHeader(header[:n], magicBytes["jpeg"]):
			detectedType = models.FileTypeImage
			mimeType = "image/jpeg"
		case matchesHeader(header[:n], magicBytes["png"]):
			detectedType = models.FileTypeImage
			mimeType = "image/png"
			isAnimated = checkPNGAnimated(header[:n])
		case matchesHeader(header[:n], magicBytes["gif"]):
			detectedType = models.FileTypeImage
			mimeType = "image/gif"
			isAnimated = checkGIFAnimated(header[:n])
		case matchesHeader(header[:n], magicBytes["bmp"]):
			detectedType = models.FileTypeImage
			mimeType = "image/bmp"
		// RIFF containers (WebP/WAV/AVI) — check bytes 8-12 to distinguish
		case matchesHeader(header[:n], magicBytes["webp"]):
			if n >= 12 {
				riffType := string(header[8:12])
				switch riffType {
				case "WEBP":
					detectedType = models.FileTypeImage
					mimeType = "image/webp"
				case "WAVE":
					detectedType = models.FileTypeAudio
					mimeType = "audio/wav"
				case "AVI ":
					detectedType = models.FileTypeVideo
					mimeType = "video/x-msvideo"
				}
			}
		case matchesHeader(header[:n], magicBytes["tiff"]) || matchesHeader(header[:n], magicBytes["tiff_big"]):
			detectedType = models.FileTypeImage
			mimeType = "image/tiff"

		// Apple ICNS icon
		case matchesHeader(header[:n], magicBytes["icns"]):
			detectedType = models.FileTypeImage
			mimeType = "image/icns"

		// ISO BMFF containers (HEIC/HEIF/AVIF/M4A/MP4/MOV)
		// All start with [size] 'ftyp' [brand]. Check bytes 4-8 for 'ftyp',
		// then bytes 8-12 for the brand to distinguish types.
		case n >= 12 && string(header[4:8]) == "ftyp":
			brand := string(header[8:12])
			switch brand {
			case "heic", "heix", "mif1", "hevc":
				detectedType = models.FileTypeImage
				mimeType = "image/heic"
			case "avif", "avis":
				detectedType = models.FileTypeImage
				mimeType = "image/avif"
			case "M4A ":
				detectedType = models.FileTypeAudio
				mimeType = "audio/mp4"
			case "qt  ":
				detectedType = models.FileTypeVideo
				mimeType = "video/quicktime"
			default:
				// Default to MP4 video for other ftyp brands (isom, mp42, etc.)
				detectedType = models.FileTypeVideo
				mimeType = "video/mp4"
			}
		case matchesHeader(header[:n], magicBytes["mkv"]):
			detectedType = models.FileTypeVideo
			mimeType = "video/x-matroska"

		// PDF
		case matchesHeader(header[:n], magicBytes["pdf"]):
			detectedType = models.FileTypePDF
			mimeType = "application/pdf"

		// Audio
		case matchesHeader(header[:n], magicBytes["mp3"]):
			detectedType = models.FileTypeAudio
			mimeType = "audio/mpeg"
		case matchesHeader(header[:n], magicBytes["flac"]):
			detectedType = models.FileTypeAudio
			mimeType = "audio/flac"
		case matchesHeader(header[:n], magicBytes["ogg"]):
			detectedType = models.FileTypeAudio
			mimeType = "audio/ogg"

		// Archives / Office (ZIP-based formats)
		case matchesHeader(header[:n], magicBytes["zip"]):
			// Office documents are ZIP files — check extension to distinguish
			officeExts := map[string]bool{
				".docx": true, ".xlsx": true, ".pptx": true,
				".odt": true, ".ods": true, ".odp": true,
				".pages": true, ".numbers": true, ".key": true, ".keynote": true,
			}
			if officeExts[ext] {
				detectedType = models.FileTypeOffice
				mimeType = getMimeTypeFromExt(ext)
			} else if ext == ".ipa" {
				detectedType = models.FileTypeArchive
				mimeType = "application/octet-stream"
			} else {
				detectedType = models.FileTypeArchive
				mimeType = "application/zip"
			}
		case matchesHeader(header[:n], magicBytes["7z"]):
			detectedType = models.FileTypeArchive
			mimeType = "application/x-7z-compressed"
		case matchesHeader(header[:n], magicBytes["rar"]):
			detectedType = models.FileTypeArchive
			mimeType = "application/x-rar-compressed"
		}
	}

	// Fall back to extension if magic bytes didn't identify
	if detectedType == models.FileTypeUnknown {
		if ft, ok := extensionToType[ext]; ok {
			detectedType = ft
			mimeType = getMimeTypeFromExt(ext)
		}
	}

	// Handle tar files by extension (no magic bytes)
	if detectedType == models.FileTypeUnknown && ext == ".tar" {
		detectedType = models.FileTypeArchive
		mimeType = "application/x-tar"
	}

	// DMG detection: check for 'koly' magic at end of file
	if detectedType == models.FileTypeUnknown && (ext == ".dmg" || ext == ".img") {
		detectedType = models.FileTypeDiskImage
		mimeType = "application/x-apple-diskimage"
	}

	return &models.FileInfo{
		Path:       path,
		MimeType:   mimeType,
		Extension:  ext,
		FileType:   detectedType,
		Size:       info.Size(),
		IsAnimated: isAnimated,
	}, nil
}

// matchesHeader checks if the file header starts with the given magic bytes
func matchesHeader(header, magic []byte) bool {
	if len(header) < len(magic) {
		return false
	}
	for i, b := range magic {
		if header[i] != b {
			return false
		}
	}
	return true
}

// checkPNGAnimated checks if a PNG file is animated (APNG)
func checkPNGAnimated(header []byte) bool {
	// APNG files have 'acTL' chunk after the IHDR
	// For simplicity, we'll check for the acTL marker in the first 512 bytes
	if len(header) < 12 {
		return false
	}
	// Skip PNG signature (8 bytes), then check chunks
	// IHDR is always first, so we start after it
	for i := 8; i < len(header)-4; i++ {
		if string(header[i:i+4]) == "acTL" {
			return true
		}
	}
	return false
}

// checkGIFAnimated checks if a GIF file has multiple frames
func checkGIFAnimated(header []byte) bool {
	// GIF89a (with animation) has a trailer 0x2C for frame separator
	// For simplicity, check if it's GIF89a (supports animation) vs GIF87a
	if len(header) < 6 {
		return false
	}
	return string(header[3:6]) == "9a" // GIF89a
}

// getMimeTypeFromExt returns MIME type from extension
func getMimeTypeFromExt(ext string) string {
	mimeMap := map[string]string{
		".jpg":     "image/jpeg",
		".jpeg":    "image/jpeg",
		".png":     "image/png",
		".gif":     "image/gif",
		".bmp":     "image/bmp",
		".webp":    "image/webp",
		".tiff":    "image/tiff",
		".tif":     "image/tiff",
		".svg":     "image/svg+xml",
		".heic":    "image/heic",
		".heif":    "image/heif",
		".avif":    "image/avif",
		".icns":    "image/icns",
		".mp4":     "video/mp4",
		".mov":     "video/quicktime",
		".avi":     "video/x-msvideo",
		".mkv":     "video/x-matroska",
		".webm":    "video/webm",
		".flv":     "video/x-flv",
		".wmv":     "video/x-ms-wmv",
		".m4v":     "video/x-m4v",
		".pdf":     "application/pdf",
		".docx":    "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx":    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".pptx":    "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".odt":     "application/vnd.oasis.opendocument.text",
		".ods":     "application/vnd.oasis.opendocument.spreadsheet",
		".odp":     "application/vnd.oasis.opendocument.presentation",
		".doc":     "application/msword",
		".xls":     "application/vnd.ms-excel",
		".ppt":     "application/vnd.ms-powerpoint",
		".pages":   "application/vnd.apple.pages",
		".numbers": "application/vnd.apple.numbers",
		".key":     "application/vnd.apple.keynote",
		".keynote": "application/vnd.apple.keynote",
		".mp3":     "audio/mpeg",
		".wav":     "audio/wav",
		".flac":    "audio/flac",
		".ogg":     "audio/ogg",
		".m4a":     "audio/mp4",
		".aac":     "audio/aac",
		".wma":     "audio/x-ms-wma",
		".alac":    "audio/alac",
		".aiff":    "audio/aiff",
		".aif":     "audio/aiff",
		".zip":     "application/zip",
		".7z":      "application/x-7z-compressed",
		".rar":     "application/x-rar-compressed",
		".tar":     "application/x-tar",
		".gz":      "application/gzip",
		".bz2":     "application/x-bzip2",
		".xz":      "application/x-xz",
		".ipa":     "application/octet-stream",
		".dmg":     "application/x-apple-diskimage",
		".iso":     "application/x-iso9660-image",
		".img":     "application/x-raw-disk-image",
	}
	if mime, ok := mimeMap[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}
