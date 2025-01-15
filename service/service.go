package service

import (
	"cmp"
	"context"
	"frontend-gafam/service/arte"
	"frontend-gafam/service/common"
	"frontend-gafam/service/front"
	"frontend-gafam/service/instagram"
	"frontend-gafam/service/lfi"
	"frontend-gafam/service/peertube"
	"frontend-gafam/service/rss"
	"frontend-gafam/service/tiktok"
	"frontend-gafam/service/twitch"
	"frontend-gafam/service/youtube"
	"log/slog"
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
		t.Log(context.Background(), slog.LevelInfo+2, "target", "u", u)
		protoAndHandle, id, _ := strings.Cut(u, ":")
		proto, _, _ := strings.Cut(protoAndHandle, "#")
		switch proto {
		case "arte.cat":
			index.Lists = append(index.Lists, arte.Category(t, id)...)
		case "arte.ch":
			index.Lists = append(index.Lists, arte.Channel(t, id))
		case "arte.li":
			index.Lists = append(index.Lists, arte.List(t, id)...)
		case "insta.ch":
			index.Lists = append(index.Lists, instagram.User(t, id))
		case "insta.tr+ch":
			index.Lists = append(index.Lists, instagram.WithThread(t, id))
		case "lfi.g":
			index.Lists = append(index.Lists, lfi.Group(t, id))
		case "peertube.a":
			index.Lists = append(index.Lists, peertube.User(t, id))
		case "peertube.c":
			index.Lists = append(index.Lists, peertube.Channel(t, id))
		case "rss":
			index.Lists = append(index.Lists, rss.Fetch(t, id))
		case "tiktok.ch":
			index.Lists = append(index.Lists, tiktok.Channel(t, id))
		case "twitch.ch":
			index.Lists = append(index.Lists, twitch.Channel(t, id))
		case "twitch.te":
			index.Lists = append(index.Lists, twitch.Team(t, id)...)
		case "yt.charts.titles":
			index.Lists = append(index.Lists, youtube.ChartsTitles(t, id))
		case "yt.ch":
			index.Lists = append(index.Lists, youtube.ChannelRSS(t, id))
		case "yt.pl":
			index.Lists = append(index.Lists, youtube.PlaylistRSS(t, id))
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
	slices.SortFunc(index.News, func(a, b *common.Item) int { return a.Published.Compare(b.Published) })
	index.News = slices.CompactFunc(index.News, func(a, b *common.Item) bool {
		return a.URL == b.URL
	})

	return
}
