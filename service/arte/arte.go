package arte

import (
	"crypto/sha256"
	"encoding/base64"
	"frontend-gafam/service/common"
	"regexp"
	"slices"
	"sniffle/tool"
	"sniffle/tool/fetch"
	"strings"
	"time"
)

func Category(t *tool.Tool, id string) []*common.List {
	dto := struct {
		Tag   string
		Value struct {
			Zones []struct {
				ID          string
				Title       string
				Description string
				contentItems
			}
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, fetch.URL("https://www.arte.tv/api/rproxy/emac/v4/fr/web/pages/"+id+"/")) {
		return nil
	} else if dto.Tag != "Ok" {
		t.Warn("dto.err", "err", "tag is not 'Ok'", "tag", dto.Tag)
		return nil
	}

	lists := make([]*common.List, 0, len(dto.Value.Zones))
	for _, zone := range dto.Value.Zones {
		items := trItems(t, zone.contentItems)
		if len(items) == 0 {
			continue
		}
		lists = append(lists, &common.List{
			Host:        "arte",
			ID:          genID(zone.Title),
			Title:       zone.Title,
			Description: []string{zone.Description},
			Items:       items,
		})
	}

	return lists
}

func List(t *tool.Tool, id string) []*common.List {
	dto := struct {
		Tag   string
		Value struct {
			Zones []struct {
				ID          string
				Title       string
				Description string
				contentItems
			}
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, requestChannel(id)) {
		return nil
	} else if dto.Tag != "Ok" {
		t.Warn("dto.err", "err", "tag is not 'Ok'", "tag", dto.Tag)
		return nil
	}

	lists := make([]*common.List, 0, len(dto.Value.Zones))
	for _, zone := range dto.Value.Zones {
		items := trItems(t, zone.contentItems)
		if len(items) == 0 {
			continue
		}
		lists = append(lists, &common.List{
			Host:        "arte",
			ID:          genID(zone.Title),
			Title:       zone.Title,
			Description: []string{zone.Description},
			Items:       items,
		})
	}

	return lists
}
func genID(s string) string {
	h := sha256.Sum256([]byte(s))
	return base64.RawURLEncoding.EncodeToString(h[:6])
}

func Channel(t *tool.Tool, id string) *common.List {
	dto := struct {
		Tag   string
		Value struct {
			URL      string
			Metadata struct {
				Title       string
				Description string
				Publish     struct {
					Offline time.Time
				}
			}
			Zones []contentItems
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, requestChannel(id)) {
		return nil
	} else if dto.Tag != "Ok" {
		t.Warn("dto.err", "err", "tag is not 'Ok'", "tag", dto.Tag)
		return nil
	}

	items := make([]*common.Item, 0)
	for _, zone := range dto.Value.Zones {
		items = append(items, trItems(t, zone)...)
	}
	slices.SortStableFunc(items, func(a, b *common.Item) int {
		return b.Published.Compare(a.Published)
	})

	meta := dto.Value.Metadata
	return &common.List{
		Host:  "arte",
		ID:    id,
		URL:   "http://www.arte.tv" + dto.Value.URL,
		Title: meta.Title,
		Description: append(
			common.Description(meta.Description),
			"",
			"offline: "+meta.Publish.Offline.Format(time.DateOnly),
		),
		Items: items,
		JSON:  tool.FetchAll(t, requestChannel(id)),
	}
}
func requestChannel(id string) *fetch.Request {
	return fetch.URL("https://www.arte.tv/api/rproxy/emac/v4/fr/web/collections/" + id + "/")
}

type contentItems struct {
	Content struct {
		Data []struct {
			Kind struct {
				Code string
			}
			ProgramId string
			Title     string
			URL       string

			Subtitle         string
			ShortDescription string

			Availability struct {
				Start time.Time
				End   time.Time
			}
			Duration uint

			MainImage struct {
				URL string
			}
		}
	}
}

func trItems(t *tool.Tool, zone contentItems) []*common.Item {
	items := make([]*common.Item, 0, len(zone.Content.Data))
	for _, entry := range zone.Content.Data {
		if entry.Kind.Code != "SHOW" {
			continue
		}
		i := &common.Item{
			Host:  "arte",
			ID:    entry.ProgramId,
			URL:   "https://www.arte.tv" + entry.URL,
			Title: entry.Title,
			Description: []string{
				entry.Subtitle,
				entry.ShortDescription,
				"deadline: " + entry.Availability.End.Format(time.DateOnly),
			},
			Published: entry.Availability.Start,
			Duration:  time.Second * time.Duration(entry.Duration),
			IsVideo:   true,
		}
		fetchSources(t, i)
		items = append(items, i)
	}
	return items
}

// Get poster and sources
func fetchSources(t *tool.Tool, item *common.Item) {
	dto := struct {
		Data struct {
			Attributes struct {
				Metadata struct {
					Images [1]struct {
						URL string
					}
				}
				Streams []struct {
					URL string
				}
			}
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, fetch.URL("https://api.arte.tv/api/player/v2/config/fr/"+item.ID)) {
		return
	}

	poster, width, height := common.FetchPoster(t, dto.Data.Attributes.Metadata.Images[0].URL)
	item.Poster = poster
	item.PosterWidth = width
	item.PosterHeight = height

	var m3u8Urls = regexp.MustCompile(`https\://[\w/\-+\.]+\.(m3u8|vtt)`)
	var m3u8Name = regexp.MustCompile(`.*/[0-9A-Z\-]+_([\w\-]+)\.(mp4|vtt)$`)
	var m3u8vtt = regexp.MustCompile(`.*/[0-9A-Z\-]+_st[\w\-]+\.vtt$`)
	for _, streams := range dto.Data.Attributes.Streams {
		item.Sources = append(item.Sources, common.Source{
			Name: "m3u8",
			URL:  streams.URL,
		})
		m3u8 := string(tool.FetchAll(t, fetch.URL(streams.URL)))
		for _, u := range m3u8Urls.FindAllString(m3u8, -1) {
			if strings.HasSuffix(u, "_iframe_index.m3u8") ||
				strings.HasSuffix(u, "_SPR.vtt") {
				continue
			}
			if m3u8vtt.MatchString(u) {
				u = strings.Replace(u, ".m3u8", ".vtt", 1)
			} else if uu, ok := strings.CutSuffix(u, ".m3u8"); ok {
				u = uu + ".mp4"
			}
			item.Sources = append(item.Sources, common.Source{
				Name: m3u8Name.ReplaceAllString(u, "$1"),
				URL:  u,
			})
		}
	}
}
