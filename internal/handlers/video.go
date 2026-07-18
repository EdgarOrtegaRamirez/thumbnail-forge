package handlers

import (
	"fmt"
	"image/draw"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

// VideoHandler handles video file thumbnails
type VideoHandler struct{}

// CanHandle returns true if this handler can process the given file type
func (h *VideoHandler) CanHandle(info *models.FileInfo) bool {
	return info.FileType == models.FileTypeVideo
}

// Generate creates a thumbnail from a video file
func (h *VideoHandler) Generate(info *models.FileInfo, opts *models.ThumbnailOptions) (*models.ThumbnailResult, error) {
	// Check if ffmpeg is available
	if err := checkFFmpeg(); err != nil {
		return nil, fmt.Errorf("ffmpeg is required for video thumbnails: %w", err)
	}

	// Create temporary file for the extracted frame
	tmpDir, err := os.MkdirTemp("", "thumbnail-forge-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tmpFrame := filepath.Join(tmpDir, "frame.jpg")

	// Parse timestamp
	timestamp := opts.Timestamp
	if timestamp == "" {
		timestamp = "1"
	}

	// Extract frame using ffmpeg
	if err := extractFrame(info.Path, tmpFrame, timestamp); err != nil {
		return nil, fmt.Errorf("failed to extract video frame: %w", err)
	}

	// Load the extracted frame
	img, err := LoadImage(tmpFrame)
	if err != nil {
		return nil, fmt.Errorf("failed to load extracted frame: %w", err)
	}

	// Resize the image
	resized := ResizeImage(img, opts.Width, opts.Height)

	// Composite onto background
	result := compositeOnBackground(resized, opts.Background)

	return &models.ThumbnailResult{
		Image:    result,
		MimeType: "image/png",
		Width:    opts.Width,
		Height:   opts.Height,
	}, nil
}

// checkFFmpeg checks if ffmpeg is installed and accessible
func checkFFmpeg() error {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}
	return nil
}

// extractFrame extracts a single frame from a video file
func extractFrame(videoPath, outputPath, timestamp string) error {
	// Parse timestamp to seconds
	seconds, err := parseTimestamp(timestamp)
	if err != nil {
		return fmt.Errorf("invalid timestamp %q: %w", timestamp, err)
	}

	// Build ffmpeg command
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-ss", strconv.FormatFloat(seconds, 'f', 3, 64),
		"-vframes", "1",
		"-q:v", "2", // High quality JPEG
		"-y", // Overwrite output
		outputPath,
	)

	// Capture stderr for error messages
	var stderr strings.Builder
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w\n%s", err, stderr.String())
	}

	// Verify output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("ffmpeg did not create output file")
	}

	return nil
}

// parseTimestamp parses a timestamp string into seconds
func parseTimestamp(timestamp string) (float64, error) {
	// Handle "Ns" format (e.g., "1s", "5.5s")
	timestamp = strings.TrimSuffix(timestamp, "s")

	// Handle simple numeric values (e.g., "1", "5.5")
	if seconds, err := strconv.ParseFloat(timestamp, 64); err == nil {
		return seconds, nil
	}

	// Handle HH:MM:SS format (e.g., "00:00:05", "01:30:00")
	parts := strings.Split(timestamp, ":")
	if len(parts) == 3 {
		hours, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		minutes, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, err
		}
		seconds, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return 0, err
		}
		return hours*3600 + minutes*60 + seconds, nil
	}

	// Handle MM:SS format (e.g., "01:30")
	if len(parts) == 2 {
		minutes, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		seconds, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, err
		}
		return minutes*60 + seconds, nil
	}

	return 0, fmt.Errorf("invalid timestamp format: %s", timestamp)
}

// init registers the video format decoders
func init() {
	// ffmpeg is used via exec, no Go imports needed
}

// ensure these are used
var _ = imaging.Resize
var _ = draw.Draw
var _ = jpeg.Encode
