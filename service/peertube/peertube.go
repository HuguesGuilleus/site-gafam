package peertube

import (
	"context"
	"encoding/xml"
	"frontend-gafam/service/common"
	"sniffle/tool"
	"strconv"
	"strings"
	"time"
)

func User(t *tool.Tool, handle string) *common.List {
	name, host, _ := strings.Cut(handle, "@")
	if host == "" || name == "" {
		t.Warn("wrongHandleFormat", "handle", handle)
		return nil
	}

	dto := struct {
		ID int
	}{}
	if tool.FetchJSON(context.Background(), t, "", "https://"+host+"/api/v1/accounts/"+handle, nil, nil, &dto) {
		return nil
	}

	return fetchData(t, handle, name, host,
		"https://"+host+"/feeds/videos.xml?accountId="+strconv.Itoa(dto.ID))
}

func Channel(t *tool.Tool, handle string) *common.List {
	name, host, _ := strings.Cut(handle, "@")
	if host == "" || name == "" {
		t.Warn("wrongHandleFormat", "handle", handle)
		return nil
	}

	dto := struct {
		ID int
	}{}
	if tool.FetchJSON(context.Background(), t, "", "https://"+host+"/api/v1/video-channels/"+handle, nil, nil, &dto) {
		return nil
	}

	return fetchData(t, handle, name, host,
		"https://"+host+"/feeds/videos.xml?videoChannelId="+strconv.Itoa(dto.ID))
}

func fetchData(t *tool.Tool, handle, handleName, host, url string) *common.List {
	x := tool.FetchAll(context.Background(), t, "", url, nil, nil)
	dto := struct {
		Channel struct {
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Entries     []struct {
				Link        string  `xml:"link"`
				Title       string  `xml:"title"`
				Description string  `xml:"description"`
				Pub         rssTime `xml:"pubDate"`
				Community   struct {
					Statistics struct {
						View string `xml:"views,attr"`
					} `xml:"statistics"`
				} `xml:"community"`
				Thumbnail struct {
					URL string `xml:"url,attr"`
				} `xml:"thumbnail"`
				Sources struct {
					Content []struct {
						Height int    `xml:"height,attr"`
						URL    string `xml:"url,attr"`
					} `xml:"content"`
				} `xml:"group"`
			} `xml:"item"`
		} `xml:"channel"`
	}{}
	if err := xml.Unmarshal(x, &dto); err != nil {
		t.Warn("xml.decode", "url", url, "err", err.Error())
		return nil
	}

	items := make([]*common.Item, 0)
	for _, entry := range dto.Channel.Entries {
		_, id, _ := strings.Cut(entry.Link, "/w/")
		if id == "" {
			continue
		}

		poster, width, height := common.FetchPoster(t, entry.Thumbnail.URL)
		view, _ := strconv.ParseUint(entry.Community.Statistics.View, 10, 32)

		sources := make([]common.Source, 0)
		for _, s := range entry.Sources.Content {
			sources = append(sources, common.Source{URL: s.URL, Height: s.Height})
		}

		items = append(items, &common.Item{
			Host:         host,
			ID:           id,
			URL:          entry.Link,
			Title:        entry.Title,
			Description:  common.Description(entry.Description),
			Author:       dto.Channel.Title,
			Published:    entry.Pub.Time,
			View:         uint(view),
			IsVideo:      true,
			Poster:       poster,
			PosterWidth:  width,
			PosterHeight: height,
			Sources:      sources,
		})
	}

	return &common.List{
		ID:          handle,
		URL:         "https://" + host + "/c/" + handleName + "/videos",
		Title:       dto.Channel.Title,
		Description: common.Description(dto.Channel.Description),
		Items:       items,
	}
}

type rssTime struct {
	time.Time
}

func (t *rssTime) UnmarshalText(text []byte) (err error) {
	t.Time, err = time.Parse("Mon, 02 Jan 2006 15:04:05 MST", string(text))
	return
}
