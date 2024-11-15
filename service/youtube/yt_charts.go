package youtube

import (
	"frontend-gafam/service/common"
	"sniffle/tool"
	"sniffle/tool/fetch"
	"sniffle/tool/render"
	"strings"
	"time"
)

func ChartsTitles(t *tool.Tool, country string) *common.List {
	r := fetch.Rs("POST", "https://charts.youtube.com/youtubei/v1/browse?alt=json", `{`+
		`"context":{"client":{"clientName":"WEB_MUSIC_ANALYTICS","clientVersion":"2.0","hl":"en","gl":"US"}},`+
		`"browseId":"FEmusic_analytics_charts_home",`+
		`"query":"perspective=CHART_DETAILS&chart_params_country_code=`+country+`&chart_params_chart_type=TRACKS&chart_params_period_type=WEEKLY"`+
		`}`)

	dto := struct {
		Contents struct {
			SectionListRenderer struct {
				Contents [1]struct {
					MusicAnalyticsSectionRenderer struct {
						Content struct {
							TrackTypes [1]struct {
								TrackViews []struct {
									Name             string
									View             uint `json:"viewCount,string"`
									EncryptedVideoId string
									Artists          []struct {
										Name string
									}
									ReleaseDate struct {
										Year, Month, Day int
									}
									Thumbnail struct {
										Thumbnails [1]struct {
											URL string
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, r) {
		return nil
	}

	tracks := dto.Contents.SectionListRenderer.Contents[0].MusicAnalyticsSectionRenderer.Content.TrackTypes[0].TrackViews
	items := make([]*common.Item, len(tracks))
	for i, track := range tracks {
		id := track.EncryptedVideoId
		poster, width, height := []byte(nil), "", ""
		if strings.HasPrefix(track.Thumbnail.Thumbnails[0].URL, "https://lh3.googleusercontent.com/") {
			poster, width, height = common.FetchPoster(t, track.Thumbnail.Thumbnails[0].URL)
		} else {
			poster, width, height = fetchPoster(t, id)
		}

		artists := make([]string, len(track.Artists))
		for i, artist := range track.Artists {
			artists[i] = artist.Name
		}

		items[i] = &common.Item{
			Host:      "yt",
			ID:        id,
			URL:       "https://www.youtube.com/watch?v=" + id,
			Title:     track.Name,
			Author:    strings.Join(artists, " & "),
			View:      track.View,
			Published: time.Date(track.ReleaseDate.Year, time.Month(track.ReleaseDate.Month), track.ReleaseDate.Day, 0, 0, 0, 0, render.DateZone),

			Poster:       poster,
			PosterWidth:  width,
			PosterHeight: height,
		}
	}

	return &common.List{
		Host:  "yt",
		ID:    "charts.title." + country,
		URL:   "https://charts.youtube.com/charts/TopSongs/" + country + "/weekly",
		Title: country,
		Items: items,
		JSON:  tool.FetchAll(t, r),
	}
}
