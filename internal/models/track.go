package models

import "fmt"

type Track struct {
	Title       string
	Artists     []string
	DurationSec int
}

// FullQuery returns the formatted search string for YouTube (Artist - Title Official Audio).
func (t Track) FullQuery() string {
	// Put artists into comma separated format
	artistString := ""
	for i, v := range t.Artists {
		if i != len(t.Artists)-1 {
			artistString += fmt.Sprintf("%v, ", v)
		} else {
			artistString += v
		}
	}

	cleanTitle := t.Title
	query := fmt.Sprintf("%s - %s Topic", artistString, cleanTitle)
	return query
}
