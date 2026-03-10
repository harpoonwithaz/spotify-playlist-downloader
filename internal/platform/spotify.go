package platform

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"GoDown/internal/models"

	"github.com/gocolly/colly"
)

// Converts string minute:second format to second int
// Returns -1 and err if it was unsuccessful
func toSec(str string) (int, error) {
	lengthSeconds := 0

	delimPos := len(str) - 2
	sec, err := strconv.Atoi(str[delimPos:])

	if err != nil {
		return -1, err
	} else {
		lengthSeconds += sec
	}

	min, err := strconv.Atoi(str[:delimPos-1]) // not including the delim
	if err != nil {
		return -1, err
	} else {
		lengthSeconds += (min * 60)
	}

	return lengthSeconds, nil
}

// I KNOW THIS IS UGLY, BUT THIS IS CUZ SPOTIFY IS SO DUMB
type SpotifyData struct {
	Props struct {
		PageProps struct {
			State struct {
				Data struct {
					Entity struct {
						Name    string `json:"name"`
						Artists []struct {
							Name string `json:"name"`
						} `json:"artists"`
						Duration       int `json:"duration"`
						VisualIdentity struct {
							Image []struct {
								URL    string `json:"url"`
								Height int    `json:"maxHeight"`
								Width  int    `json:"maxWidth"`
							} `json:"image"`
						} `json:"visualIdentity"`
					} `json:"entity"`
				} `json:"data"`
			} `json:"state"`
		} `json:"pageProps"`
	} `json:"props"`
}

func SpotifyTrackMetadata(scrapeURL string) ([]models.Track, error) {
	c := colly.NewCollector(colly.AllowedDomains("open.spotify.com"))
	track := []models.Track{}

	c.OnHTML("script#__NEXT_DATA__", func(e *colly.HTMLElement) {
		var result SpotifyData

		s := e.Text
		err := json.Unmarshal([]byte(s), &result)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		entity := result.Props.PageProps.State.Data.Entity

		var thumbURL string
		for _, img := range entity.VisualIdentity.Image {
			if img.Width == 300 && img.Height == 300 {
				thumbURL = img.URL
				break
			}
		}

		// If your models.Track has an ImageURL field:
		currentTrack := models.Track{
			Title:       entity.Name,
			DurationSec: entity.Duration / 1000,
			ArtURL:      thumbURL,
		}

		for _, artist := range entity.Artists {
			currentTrack.Artists = append(currentTrack.Artists, artist.Name)
		}

		track = append(track, currentTrack)
	})

	c.Visit(scrapeURL)

	return track, nil
}

// Returns the playlist items from a spotify URL as a track struct
// func SpotifyPlaylistMetadata(scrapeURL string) ([]models.Track, error) {
// 	playlistTracks := []models.Track{}

// 	// scrape the embed HTML for track info
// 	c := colly.NewCollector(
// 		colly.AllowedDomains("open.spotify.com"),
// 	)

// 	// Playlist-level art URL (from CSS variable on the main container)
// 	var playlistArtURL string

// 	var visitErr error

// 	c.OnResponse(func(r *colly.Response) {
// 		if r.StatusCode != 200 {
// 			visitErr = fmt.Errorf("unexpected status code: %d", r.StatusCode)
// 			return
// 		}
// 		// Keep both lowercased body for availability checks and original for extraction
// 		bodyLower := strings.ToLower(string(r.Body))
// 		body := string(r.Body)

// 		// Try to extract playlist cover art from inline CSS variable present
// 		// e.g. --image-src:url('https://image-cdn-ak.spotifycdn.com/image/...')
// 		if idx := strings.Index(body, "--image-src:url("); idx != -1 {
// 			start := idx + len("--image-src:url(")
// 			if end := strings.Index(body[start:], ")"); end != -1 {
// 				raw := body[start : start+end]
// 				raw = strings.Trim(raw, "'\" ")
// 				raw = html.UnescapeString(raw)
// 				if playlistArtURL == "" {
// 					playlistArtURL = raw
// 				}
// 			}
// 		}
// 		if strings.Contains(bodyLower, "page not found") || strings.Contains(bodyLower, "this content is not available") || strings.Contains(bodyLower, "playlist is private") {
// 			visitErr = fmt.Errorf("playlist not available or private")
// 			return
// 		}
// 	})

// 	c.OnError(func(r *colly.Response, er error) {
// 		// network / request error
// 		visitErr = fmt.Errorf("request failed: %w", er)
// 	})

// 	// htmlElm := "ol[aria-label=Track list]"
// 	c.OnHTML("li", func(h *colly.HTMLElement) {
// 		var currentTrack models.Track

// 		title := h.ChildText("h3")

// 		// Ignore span tag when track is explicit
// 		artistRaw := ""
// 		h.DOM.Find("h4").Contents().Each(func(i int, s *goquery.Selection) {
// 			if goquery.NodeName(s) == "#text" {
// 				artistRaw += s.Text()
// 			}
// 		})

// 		// Multiple artists
// 		artists := strings.Split(artistRaw, "\u00a0")
// 		for i, a := range artists {
// 			// removes trailing comma for every artist except last
// 			if i != (len(artists) - 1) {
// 				artists[i] = a[:len(a)-1]
// 			} else {
// 				artists[i] = strings.TrimSpace(a)
// 			}
// 		}

// 		// convert to seconds from minutes:seconds format
// 		durationText := h.ChildText("div[data-testid=duration-cell]")
// 		duration, err := toSec(durationText)
// 		if err != nil {
// 			fmt.Printf("error converting duration: %v\nDuration text %v\nDuration set to -1", err, durationText)
// 		}

// 		if title != "" && len(artists) != 0 {
// 			// fmt.Printf("Title: %s\nMain Artist: *%v*\nDuration: %v\n\n", title, artists, duration)
// 			currentTrack.Title = title
// 			currentTrack.Artists = artists
// 			currentTrack.DurationSec = duration

// 			// If we found a playlist-level cover art, attach it to the track
// 			if playlistArtURL != "" {
// 				currentTrack.ArtURL = playlistArtURL
// 			}

// 			playlistTracks = append(playlistTracks, currentTrack)
// 		}
// 	})

// 	// Extract playlist cover image from the main container's inline style:
// 	// --image-src:url('https://image-cdn-ak.spotifycdn.com/image/...')
// 	c.OnHTML("div[data-testid=main-page]", func(e *colly.HTMLElement) {
// 		style := e.Attr("style")
// 		// Find url(...) and strip quotes
// 		if idx := strings.Index(style, "url("); idx != -1 {
// 			start := idx + len("url(")
// 			if end := strings.Index(style[start:], ")"); end != -1 {
// 				raw := style[start : start+end]
// 				raw = strings.Trim(raw, "'\" ")
// 				raw = html.UnescapeString(raw)
// 				if playlistArtURL == "" {
// 					playlistArtURL = raw
// 				}
// 			}
// 		}
// 	})

// 	c.OnRequest(func(r *colly.Request) {
// 		fmt.Println("Visiting", r.URL.String())
// 	})

// 	c.Visit(scrapeURL)

// 	if visitErr != nil {
// 		return []models.Track{}, visitErr
// 	}

// 	return playlistTracks, nil
// }

// Internal structs to map the JSON structure provided in the prompt
type spotifyEmbedData struct {
	Props struct {
		PageProps struct {
			State struct {
				Data struct {
					Entity struct {
						Name     string `json:"name"`
						CoverArt struct {
							Sources []struct {
								URL string `json:"url"`
							} `json:"sources"`
						} `json:"coverArt"`
						TrackList []struct {
							Title    string `json:"title"`
							Subtitle string `json:"subtitle"`
							Duration int    `json:"duration"`
						} `json:"trackList"`
					} `json:"entity"`
				} `json:"data"`
			} `json:"state"`
		} `json:"pageProps"`
	} `json:"props"`
}

func SpotifyPlaylistMetadata(scrapeURL string) ([]models.Track, error) {
	playlistTracks := []models.Track{}
	var visitErr error

	c := colly.NewCollector(
		colly.AllowedDomains("embed.spotify.com", "open.spotify.com"),
	)

	c.OnHTML("script#__NEXT_DATA__", func(e *colly.HTMLElement) {
		var result spotifyEmbedData
		err := json.Unmarshal([]byte(e.Text), &result)
		if err != nil {
			visitErr = fmt.Errorf("JSON Unmarshal error: %w", err)
			return
		}

		entity := result.Props.PageProps.State.Data.Entity

		// Extract Playlist Art (using the first source as default)
		var playlistArt string
		if len(entity.CoverArt.Sources) > 0 {
			playlistArt = entity.CoverArt.Sources[0].URL
		}

		// Parse the Track List
		for _, t := range entity.TrackList {
			// Subtitle typically contains artist names like "Artist 1, Artist 2"
			artists := strings.Split(t.Subtitle, ", ")

			playlistTracks = append(playlistTracks, models.Track{
				Title:       t.Title,
				Artists:     artists,
				DurationSec: t.Duration / 1000,
				ArtURL:      playlistArt,
			})
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		visitErr = fmt.Errorf("colly request error: %w", err)
	})

	c.Visit(scrapeURL)

	if visitErr != nil {
		return nil, visitErr
	}

	if len(playlistTracks) == 0 {
		return nil, fmt.Errorf("no tracks found: check if the selector script#__NEXT_DATA__ is present")
	}

	return playlistTracks, nil
}
