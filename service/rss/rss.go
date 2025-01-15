package rss

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"frontend-gafam/service/common"
	"regexp"
	"sniffle/tool"
	"sniffle/tool/fetch"
	"strconv"
	"strings"
	"time"
)

func Fetch(t *tool.Tool, url string) *common.List {
	r := fetch.URL(url)

	dto := struct {
		Channel struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			Item        []struct {
				Title       string  `xml:"title"`
				Link        string  `xml:"link"`
				Description string  `xml:"description"`
				PubDate     rssTime `xml:"pubDate"`
				Author      string  `xml:"author"`
				Duration    string  `xml:"duration"`
				Enclosure   struct {
					URL    string `xml:"url,attr"`
					Type   string `xml:"type,attr"`
					Length int    `xml:"length,attr"`
				} `xml:"enclosure"`
				Image struct {
					Href string `xml:"href,attr"`
				} `xml:"image"`
			} `xml:"item"`
		} `xml:"channel"`
	}{}

	data := tool.FetchAll(t, r)
	if err := xml.Unmarshal(data, &dto); err != nil {
		t.Warn("xml.decode", "url", url, "err", err.Error())
		return nil
	}

	list := &common.List{
		Host:        "rss",
		ID:          genID(url),
		URL:         dto.Channel.Link,
		Title:       dto.Channel.Title,
		Description: common.Description(dto.Channel.Description),
		Items:       make([]*common.Item, len(dto.Channel.Item)),
	}
	for i, dto := range dto.Channel.Item {
		poster, posterWidth, posterHeight := common.FetchPoster(t, dto.Image.Href)
		if len(poster) == 0 {
			poster, posterWidth, posterHeight = common.TextPoster(dto.Link)
		}

		item := &common.Item{
			Host:        "rss",
			ID:          genID(dto.Link),
			URL:         dto.Link,
			Title:       dto.Title,
			Description: common.Description(dto.Description),
			Author:      dto.Author,
			Published:   dto.PubDate.Time,

			Poster:       poster,
			PosterWidth:  posterWidth,
			PosterHeight: posterHeight,
		}

		if dto.Enclosure.Type == "audio/mpeg" {
			item.IsVideo = true
			item.Sources = []common.Source{{Name: "audio", URL: dto.Enclosure.URL}}

			if dto.Duration == "" {
				item.Duration = time.Duration(dto.Enclosure.Length) * time.Second
			} else if durationColon.MatchString(dto.Duration) {
				h, m, s := 0, 0, 0
				fmt.Sscanf(dto.Duration, "%d:%d:%d", &h, &m, &s)
				item.Duration = time.Duration(h)*time.Hour +
					time.Duration(m)*time.Minute +
					time.Duration(s)*time.Second
			} else if durationInt.MatchString(dto.Duration) {
				d, _ := strconv.Atoi(strings.TrimSuffix(dto.Duration, "\""))
				item.Duration = time.Duration(d) * time.Second
			} else {
				t.Warn("wrongDurationFormat", "s", dto.Duration)
			}
		}

		list.Items[i] = item
	}

	return list
}

var durationColon = regexp.MustCompile(`^\d\d:\d\d:\d\d$`)
var durationInt = regexp.MustCompile(`^\d+"?$`)

type rssTime struct {
	time.Time
}

func (t *rssTime) UnmarshalText(text []byte) (err error) {
	t.Time, err = time.Parse(time.RFC1123Z, string(text))
	return
}

func genID(s string) string {
	h := sha256.Sum224([]byte(s))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
