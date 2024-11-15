package youtube

import (
	"encoding/xml"
	"frontend-gafam/service/common"
	"sniffle/tool"
	"sniffle/tool/fetch"
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
	data := tool.FetchAll(t, fetch.R("", dataURL, nil))
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
				V int `xml:"count,attr"`
			} `xml:"group>community>starRating"`
		} `xml:"entry"`
	}{}
	if err := xml.Unmarshal(data, &dto); err != nil {
		t.Warn("rss.err", "url", dataURL, "err", err.Error())
		return nil
	}

	items := make([]*common.Item, len(dto.Entry))
	for i, entry := range dto.Entry {
		poster, width, height := fetchPoster(t, entry.ID)
		items[i] = &common.Item{
			Host:         "youtube",
			ID:           entry.ID,
			URL:          "https://www.youtube.com/watch?v=" + entry.ID,
			Title:        entry.Title,
			Description:  common.Description(entry.Description),
			Author:       entry.AuthorName,
			Published:    entry.Published,
			Updated:      entry.Updated,
			Like:         entry.Like.V,
			View:         entry.View.V,
			IsVideo:      true,
			Poster:       poster,
			PosterWidth:  width,
			PosterHeight: height,
		}
	}

	return &common.List{
		Host:  "yt",
		ID:    dto.ID,
		URL:   humanURL,
		Title: dto.Title,
		Items: items,
		JSON:  common.Xml2Json(data),
	}
}

func fetchPoster(t *tool.Tool, id string) (poster []byte, width, height string) {
	return common.FetchPoster(t, "https://img.youtube.com/vi/"+id+"/hqdefault.jpg")
}
