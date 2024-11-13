package tiktok

import (
	"encoding/json"
	"frontend-gafam/service/common"
	"slices"
	"sniffle/tool"
	"sniffle/tool/fetch"
	"strconv"
	"strings"
	"time"
)

func Channel(t *tool.Tool, id string) *common.List {
	url := "https://www.tiktok.com/@" + id
	html := string(tool.FetchAll(t, fetch.R("", url, nil)))

	_, jsonAndEnd, ok := strings.Cut(html, `<script id="__UNIVERSAL_DATA_FOR_REHYDRATION__" type="application/json">`)
	if !ok {
		t.Warn("tiktok", "err", "missing data script begin markup")
		return nil
	}
	j, _, ok := strings.Cut(jsonAndEnd, string(`</script>`))
	if !ok {
		t.Warn("tiktok", "err", "missing data script end markup")
		return nil
	}

	dto := struct {
		Scope struct {
			UserDetail struct {
				UserInfo struct {
					User struct {
						Nickname  string
						Signature string
						SecUid    string
					}
				}
			} `json:"webapp.user-detail"`
		} `json:"__DEFAULT_SCOPE__"`
	}{}
	json.Unmarshal([]byte(j), &dto)

	user := dto.Scope.UserDetail.UserInfo.User
	items := append(
		fetchList(t, "https://www.tiktok.com/api/creator/item_list/?aid=1988&count=15&cursor="+
			strconv.FormatInt(time.Now().Round(time.Hour).Unix()*1000, 10)+
			"&type=0&secUid="+user.SecUid),
		fetchList(t, "https://www.tiktok.com/api/repost/item_list/?aid=1988&count=30&cursor=0&secUid="+user.SecUid)...,
	)
	slices.SortStableFunc(items, func(a, b *common.Item) int {
		return b.Published.Compare(a.Published)
	})
	return &common.List{
		ID:          id,
		URL:         url,
		Title:       user.Nickname,
		Description: common.Description(user.Signature),
		Items:       items,
		JSON:        []byte(j),
	}
}

func fetchList(t *tool.Tool, url string) []*common.Item {
	dto := struct {
		ItemList []struct {
			Author struct {
				UniqueId string
			}
			ID         string
			CreateTime int64
			Desc       string
			Video      struct {
				Cover string
			}
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, fetch.R("", url, nil)) {
		return nil
	}

	items := make([]*common.Item, 0, len(dto.ItemList))
	for _, dto := range dto.ItemList {
		if dto.CreateTime == 0 {
			continue
		}
		poster, width, height := common.FetchPoster(t, dto.Video.Cover)

		author := dto.Author.UniqueId
		items = append(items, &common.Item{
			Host:      "tiktok",
			ID:        dto.ID,
			URL:       "https://tiktok.com/@" + author + "/video/" + dto.ID,
			Title:     dto.Desc,
			Author:    author,
			Published: time.Unix(dto.CreateTime, 0),
			IsVideo:   true,

			Poster:       poster,
			PosterWidth:  width,
			PosterHeight: height,
		})
	}
	return items
}
