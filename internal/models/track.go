package models

import "fmt"

type Track struct {
	Title       string
	Artists     []string
	DurationSec int
}

// FullQuery returns the formatted search string for YouTube (Artist - Title {query parameter}).
func (t Track) FullQuery(queryParam string) string {
	// Put artists into comma separated format
	artistString := ""
	for i, v := range t.Artists {
		if i != len(t.Artists)-1 {
			artistString += fmt.Sprintf("%v, ", v)
		} else {
			artistString += v
		}
	}

	query := fmt.Sprintf("%s - %s %s", artistString, t.Title, queryParam)
	return query
}
