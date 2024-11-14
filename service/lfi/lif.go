package lfi

import (
	"bytes"
	"frontend-gafam/service/common"
	"sniffle/tool"
	"sniffle/tool/fetch"
	"time"
)

func Group(t *tool.Tool, id string) *common.List {
	list := groupInfo(t, id)
	if list == nil {
		return nil
	}
	list.Items = groupFuture(t, id)

	list.JSON = bytes.Join([][]byte{
		{'['},
		tool.FetchAll(t, groupInfoRequest(id)),
		{','},
		tool.FetchAll(t, futureEventRequest(id)),
		{']'},
	}, nil)

	return list
}

func groupInfo(t *tool.Tool, id string) *common.List {
	dto := struct {
		Name            string
		TextDescription string
	}{}
	if tool.FetchJSON(t, nil, &dto, groupInfoRequest(id)) {
		return nil
	}

	return &common.List{
		Host:        "lfi",
		ID:          id,
		URL:         "https://actionpopulaire.fr/groupes/" + id + "/",
		Title:       dto.Name,
		Description: common.Description(dto.TextDescription),
	}
}
func groupInfoRequest(id string) *fetch.Request {
	return fetch.URL("https://actionpopulaire.fr/api/groupes/" + id + "/")
}

func groupFuture(t *tool.Tool, id string) []*common.Item {
	dto := []struct {
		ID           string
		Name         string
		StartTime    time.Time
		Illustration struct {
			Banner string
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, futureEventRequest(id)) {
		return nil
	}

	items := make([]*common.Item, len(dto))
	for i, event := range dto {
		poster, width, height := common.FetchPoster(t, event.Illustration.Banner)
		items[i] = &common.Item{
			Host:         "lfi",
			ID:           event.ID,
			URL:          "https://actionpopulaire.fr/evenements/" + event.ID + "/",
			Title:        event.Name,
			Published:    event.StartTime,
			Poster:       poster,
			PosterWidth:  width,
			PosterHeight: height,
		}
	}
	return items
}
func futureEventRequest(id string) *fetch.Request {
	return fetch.URL("https://actionpopulaire.fr/api/groupes/" + id + "/evenements/a-venir/")
}
