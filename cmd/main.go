// Oliver T.

package main

import (
	"fmt"
	"log"

	"spotify-playlist-downloader/internal/platform"
)

func main() {
	testURL := "https://open.spotify.com/playlist/0uuSKJdtUWEmgsGHvQbs5O"
	playlistTracks, err := platform.FetchMetadata(testURL)
	if err != nil {
		log.Fatalf("Critical Failure: %v", err)
	}

	for _, track := range playlistTracks {
		fmt.Println(track.FullQuery())
	}
}
