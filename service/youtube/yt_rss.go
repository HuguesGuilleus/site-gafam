package youtube

import (
	"context"
	"encoding/xml"
	"frontend-gafam/service/common"
	"sniffle/tool"
	"time"
)

func ChannelRSS(t *tool.Tool, id string) *common.List {
	return fetchRSS(t,
		"https://youtube.com/channel/"+id,
		"https://www.youtube.com/feeds/videos.xml?channel_id="+id)
}

func PlaylistRSS(t *tool.Tool, id string) *common.List {
	return fetchRSS(t,
		"https://www.youtube.com/playlist?list="+id,
		"https://www.youtube.com/feeds/videos.xml?playlist_id="+id)
}

func fetchRSS(t *tool.Tool, humanURL, dataURL string) *common.List {
	data := tool.FetchAll(context.Background(), t, "", dataURL, nil, nil)
	dto := struct {
		Title string `xml:"title"`
		ID    string `xml:"channelId"`
		Entry []struct {
			ID          string    `xml:"videoId"`
			Title       string    `xml:"title"`
			AuthorURI   string    `xml:"author>uri"`
			AuthorName  string    `xml:"author>name"`
			Published   time.Time `xml:"published"`
			Updated     time.Time `xml:"updated"`
			Description string    `xml:"group>description"`

			View struct {
				V uint `xml:"views,attr"`
			} `xml:"group>community>statistics"`
			Like struct {
				V uint `xml:"count,attr"`
			} `xml:"group>community>starRating"`
		} `xml:"entry"`
	}{}
	if err := xml.Unmarshal(data, &dto); err != nil {
		t.Warn("rss.err", "url", dataURL, "err", err.Error())
		return nil
	}

	items := make([]*common.Item, len(dto.Entry))
	for i, entry := range dto.Entry {
		poster, width, height := common.FetchPoster(t,
			"https://img.youtube.com/vi/"+entry.ID+"/hqdefault.jpg")
		items[i] = &common.Item{
			ID:           entry.ID,
			URL:          "https://www.youtube.com/watch?v=" + entry.ID,
			Title:        entry.Title,
			Description:  common.Description(entry.Description),
			Author:       entry.AuthorName,
			Published:    entry.Published,
			Updated:      entry.Updated,
			Like:         entry.Like.V,
			View:         entry.View.V,
			Poster:       poster,
			PosterWidth:  width,
			PosterHeight: height,
		}
	}

	return &common.List{
		ID:    dto.ID,
		URL:   humanURL,
		Title: dto.Title,
		Items: items,
		JSON:  common.Xml2Json(data),
	}
}
