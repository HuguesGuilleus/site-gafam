package service

import (
	"cmp"
	"frontend-gafam/service/common"
	"frontend-gafam/service/front"
	"frontend-gafam/service/peertube"
	"frontend-gafam/service/twitch"
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
		proto, id, _ := strings.Cut(u, ":")
		switch proto {
		case "yt.ch":
			index.Lists = append(index.Lists, youtube.ChannelRSS(t, id))
		case "yt.pl":
			index.Lists = append(index.Lists, youtube.PlaylistRSS(t, id))
		case "peertube.a":
			index.Lists = append(index.Lists, peertube.User(t, id))
		case "peertube.c":
			index.Lists = append(index.Lists, peertube.Channel(t, id))
		case "twitch.ch":
			index.Lists = append(index.Lists, twitch.Channel(t, id))
		case "twitch.te":
			index.Lists = append(index.Lists, twitch.Team(t, id)...)
		default:
			t.Warn("unknown.urlproto", "proto", proto, "id", id)
			continue
		}
	}
	index.Lists = slices.DeleteFunc(index.Lists, func(list *common.List) bool { return list == nil })
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
