package service

import (
	"cmp"
	"frontend-gafam/service/common"
	"frontend-gafam/service/front"
	"frontend-gafam/service/peertube"
	"frontend-gafam/service/youtube"
	"slices"
	"sniffle/tool"
	"strings"
	"time"
)

func Do(t *tool.Tool, targets map[string][]string) {
	for title, urls := range targets {
		index := fetchAll(t, title, urls)
		front.Render(t, strings.ToLower(title), &index)
	}
	front.RenderTitles(t, targets)
}

func fetchAll(t *tool.Tool, title string, urls []string) (index common.Index) {
	index.Title = title
	index.Lists = make([]*common.List, 0, len(urls))

	for _, u := range urls {
		list := (*common.List)(nil)
		proto, id, _ := strings.Cut(u, ":")
		switch proto {
		case "yt.ch":
			list = youtube.ChannelRSS(t, id)
		case "yt.pl":
			list = youtube.PlaylistRSS(t, id)
		case "peertube.ch":
			list = peertube.Channel(t, id)
		default:
			t.Warn("unknown.urlproto", "proto", proto, "id", id)
			continue
		}
		if list != nil {
			index.Lists = append(index.Lists, list)
		}
	}
	slices.SortStableFunc(index.Lists, func(a, b *common.List) int {
		return cmp.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title))
	})

	after := time.Now().Add(-time.Hour * 24 * 3)
	index.News = make([]*common.Item, 0)
	for _, list := range index.Lists {
		for _, item := range list.Items {
			if item.Published.After(after) {
				index.News = append(index.News, item)
			}
		}
	}
	slices.SortFunc(index.News, func(a, b *common.Item) int { return b.Published.Compare(a.Published) })

	return
}
