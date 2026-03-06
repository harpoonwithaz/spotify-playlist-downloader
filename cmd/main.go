// Oliver T.

package main

import (
	"context"
	"fmt"
	"log"

	"spotify-playlist-downloader/internal/downloader"
	"spotify-playlist-downloader/internal/models"
	"spotify-playlist-downloader/internal/platform"
)

func main() {
	testURL := "https://open.spotify.com/playlist/0uuSKJdtUWEmgsGHvQbs5O"
	outputDir := "downloads/"

	var options downloader.DownloadOptions
	options.Tolerances = []int{5, 10, 60}
	options.Output = outputDir + "%(title)s.%(ext)s"
	options.Queries = []string{"Topic", "Official Audio", ""}

	playlistTracks, err := platform.FetchMetadata(testURL)
	if err != nil {
		log.Fatalf("Critical Failure: %v", err)
	}

	skipped := 0
	tracksAmount := len(playlistTracks)

	// create the channels
	jobs := make(chan models.Track, tracksAmount)
	results := make(chan error, tracksAmount)

	// start 3 workers
	for w := 1; w <= 3; w++ {
		go func(workerID int) {
			for track := range jobs {
				fmt.Printf("Worker %d starting: %s\n", workerID, track.Title)
				err := downloader.DownloadTrack(context.Background(), track, options, false)
				if err != nil {
					skipped++
				}
				results <- err
			}
		}(w)
	}

	// send jobs to workers
	for _, track := range playlistTracks {
		jobs <- track
	}
	close(jobs)

	for i := 0; i < len(playlistTracks); i++ {
		err := <-results
		if err != nil {
			fmt.Println("A download failed:", err)
		}
	}

	fmt.Printf("Finished downloading. Downloaded %v/%v tracks.\n", tracksAmount-skipped, tracksAmount)
}
