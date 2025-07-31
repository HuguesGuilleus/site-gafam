package twitch

import (
	"net/http"
	"strconv"
	"time"

	"github.com/HuguesGuilleus/site-gafam/service/common"
	"github.com/HuguesGuilleus/sniffle/tool"
	"github.com/HuguesGuilleus/sniffle/tool/fetch"
)

const (
	endpointURL = "https://gql.twitch.tv/gql"
	clientID    = "kimne78kx3ncx6brgo4mv6wki5h1ko"

	channelBodyBegin = `[` +
		`{` +
		`"operationName":"HomeShelfVideos",` +
		`"variables":{` +
		`"channelLogin":`
	channelBodyMiddle = `,` +
		`"first":1` +
		`},` +
		`"extensions":{` +
		`"persistedQuery":{` +
		`"version":1,` +
		`"sha256Hash":"951c268434dc36a482c6f854215df953cf180fc2757f1e0e47aa9821258debf7"` +
		`}` +
		`}` +
		`},` +
		`{` +
		`"operationName":"ChannelRoot_AboutPanel",` +
		`"variables":{` +
		`"channelLogin":`
	channelBodyEnd = `,` +
		`"skipSchedule":true` +
		`},` +
		`"extensions":{` +
		`"persistedQuery":{` +
		`"version":1,` +
		`"sha256Hash":"6089531acef6c09ece01b440c41978f4c8dc60cb4fa0124c9a9d3f896709b6c6"` +
		`}` +
		`}` +
		`}` +
		`]`
)

func Channel(t *tool.Tool, id string) *common.List {
	qID := strconv.Quote(id)
	body := channelBodyBegin + qID + channelBodyMiddle + qID + channelBodyEnd
	list := &common.List{
		Host: "twitch",
		ID:   id,
		URL:  "https://www.twitch.tv/" + id,
		JSON: tool.FetchAll(t, fetch.Rs(http.MethodPost, endpointURL, body, "Client-ID", clientID)),
	}

	dto := []struct {
		Data struct {
			User struct {
				DisplayName  string
				Description  string
				VideoShelves struct {
					Edges []struct {
						Node struct {
							Items []struct {
								ID          string
								Title       string
								PosterURL   string `json:"previewThumbnailURL"`
								PublishedAt time.Time
								Duration    int64 `json:"lengthSeconds"`
								ViewCount   uint
							}
						}
					}
				}
			}
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, fetch.Rs(http.MethodPost, endpointURL, body, "Client-ID", clientID)) {
		return nil
	}

	for _, dto := range dto {
		if title := dto.Data.User.DisplayName; title != "" {
			list.Title = title
		}
		if d := dto.Data.User.Description; d != "" {
			list.Description = common.Description(d)
		}
	}
	for _, dto := range dto {
		for _, edges := range dto.Data.User.VideoShelves.Edges {
			for _, video := range edges.Node.Items {
				poster, width, height := common.FetchPoster(t, video.PosterURL)

				list.Items = append(list.Items, &common.Item{
					Host:         "twitch",
					ID:           video.ID,
					URL:          "https://www.twitch.tv/videos/" + video.ID,
					Title:        video.Title,
					Author:       list.Title,
					Published:    video.PublishedAt,
					Duration:     time.Duration(video.Duration) * time.Second,
					View:         video.ViewCount,
					IsVideo:      true,
					Poster:       poster,
					PosterWidth:  width,
					PosterHeight: height,
				})
			}
		}
	}

	return list
}
