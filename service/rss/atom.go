package rss

import (
	"encoding/xml"
	"frontend-gafam/service/common"
	"sniffle/tool"
	"time"
)

func atom(t *tool.Tool, url string, data []byte) *common.List {
	dto := struct {
		Title string `xml:"title"`
		Link  []struct {
			Rel  string `xml:"rel,attr"`
			Href string `xml:"href,attr"`
		}
		Entry []struct {
			ID      string    `xml:"id"`
			Updated time.Time `xml:"updated"`
			Title   string    `xml:"title"`
			Content string    `xml:"content"`
			Link    struct {
				Href string `xml:"href,attr"`
			} `xml:"link"`
			Author struct {
				Name string `xml:"name"`
			} `xml:"author"`
			Thumbnail struct {
				URL string `xml:"url,attr"`
			} `xml:"thumbnail"`
		} `xml:"entry"`
	}{}

	if err := xml.Unmarshal(data, &dto); err != nil {
		t.Warn("xml.decode", "url", url, "err", err.Error())
		return nil
	}

	u := url
	for _, link := range dto.Link {
		if link.Rel == "alternate" {
			u = link.Href
		}
	}

	list := &common.List{
		Host:  "atom",
		ID:    genID(u),
		URL:   u,
		Title: dto.Title,
		Items: make([]*common.Item, len(dto.Entry)),
	}
	for i, entry := range dto.Entry {
		poster, posterWidth, posterHeight := common.FetchPoster(t, entry.Thumbnail.URL)
		if len(poster) == 0 {
			poster, posterWidth, posterHeight = common.TextPoster(entry.Link.Href)
		}

		list.Items[i] = &common.Item{
			Host:         "atom",
			ID:           genID(entry.Link.Href),
			URL:          entry.Link.Href,
			Title:        entry.Title,
			Description:  common.Description(entry.Content),
			Author:       entry.Author.Name,
			Published:    entry.Updated,
			Poster:       poster,
			PosterWidth:  posterWidth,
			PosterHeight: posterHeight,
		}
	}

	return list
}
