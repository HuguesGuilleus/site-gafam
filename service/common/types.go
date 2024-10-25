package common

import "time"

type Index struct {
	Title string
	Lists []*List
	News  []*Item
}

// A playlist, author channel...
type List struct {
	ID          string
	URL         string
	Title       string
	Description []string
	Items       []*Item

	JSON []byte
}

// One video, image or post.
type Item struct {
	Host        string
	ID          string
	URL         string
	Title       string
	Description []string
	Author      string
	Published   time.Time
	Updated     time.Time
	Duration    time.Duration
	Like        int
	View        uint
	IsVideo     bool

	// For video
	Poster       []byte
	PosterAnnex  [][]byte
	PosterWidth  string
	PosterHeight string
	Sources      []Source
}

type Source struct {
	URL    string
	Height int
}
