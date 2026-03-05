// Oliver T.

package main

import (
	"context"
	"fmt"
	"log"

	"spotify-playlist-downloader/internal/downloader"
	"spotify-playlist-downloader/internal/platform"
)

func main() {
	testURL := "https://open.spotify.com/playlist/0uuSKJdtUWEmgsGHvQbs5O"
	outputDir := "downloads/"

	playlistTracks, err := platform.FetchMetadata(testURL)
	if err != nil {
		log.Fatalf("Critical Failure: %v", err)
	}

	var options downloader.DownloadOptions
	options.Tolerances = []int{5, 10, 60}
	options.Output = outputDir + "%(title)s.%(ext)s"
	options.Queries = []string{"Topic", "Official Audio", ""}

	skipped := 0
	tracksAmount := len(playlistTracks)

	// TRACKS IT CURRENTLY FAILS ON:
	// Sorry - Madonna      Reason: Duration difference
	// 93' til infinity     Reason: Duration difference
	for i, track := range playlistTracks {
		fmt.Printf("(%v/%v) Downloading: %s\n", i+1, tracksAmount, track.Title)
		err = downloader.DownloadTrack(context.Background(), track, options, true)
		if err != nil {
			fmt.Printf("error downloading %s: %v\n", track.Title, err)
			skipped++
		}
	}

	fmt.Printf("Finished downloading. Downloaded %v/%v tracks.\n", tracksAmount-skipped, tracksAmount)
}
