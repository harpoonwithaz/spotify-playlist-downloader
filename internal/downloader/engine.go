package downloader

import (
	"context"
	"fmt"
	"os/exec"
	"spotify-playlist-downloader/internal/models"
	"strings"
)

type DownloadOptions struct {
	Tolerances []int
	Queries    []string
	Output     string
}

// Download individual track using yt-dlp
func DownloadTrack(ctx context.Context, track models.Track, options DownloadOptions, debugging bool) error {
	maxRetries := len(options.Tolerances)

	for i, tolerance := range options.Tolerances {
		searchQuery := track.FullQuery(options.Queries[i])

		matchFilter := fmt.Sprintf("duration >= %d & duration <= %d",
			track.DurationSec-tolerance,
			track.DurationSec+tolerance,
		)

		args := []string{
			fmt.Sprintf("ytsearch1:%s", searchQuery),
			"--match-filter", matchFilter,
			"--extract-audio",
			"--audio-format", "mp3",
			"--audio-quality", "0",
			"--output", options.Output,
			"--no-playlist",
		}

		cmd := exec.CommandContext(ctx, "yt-dlp", args...)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Handle Hard Errors (Network down, binary missing, etc.)
		if err != nil && !strings.Contains(outputStr, "does not pass filter") {
			if debugging {
				fmt.Printf("[ENGINE] DEBUG: Hard Error on attempt %d: %v\n", i+1, outputStr)
			}
			// If it's the last attempt, return the error. Otherwise, try next fallback.
			if i+1 == maxRetries {
				return fmt.Errorf("yt-dlp final attempt failed: %w", err)
			}
			continue
		}

		// Handle Filter Skips
		if strings.Contains(outputStr, "does not pass filter") {
			fmt.Printf("[ENGINE] Attempt %d: No match within ±%ds. Widening...\n", i+1, tolerance)
			continue
		}

		// Handle Success
		if strings.Contains(outputStr, "[download] Destination:") || strings.Contains(outputStr, "has already been downloaded") {
			fmt.Printf("[ENGINE] Success downloading: %s (Tolerance: %ds)\n", track.Title, tolerance)
			return nil
		}
	}

	return fmt.Errorf("failed to download %s after %d attempts", track.Title, maxRetries)
}
