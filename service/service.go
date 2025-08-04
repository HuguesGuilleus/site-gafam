package service

import (
	"cmp"
	"context"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/HuguesGuilleus/site-gafam/service/arte"
	"github.com/HuguesGuilleus/site-gafam/service/common"
	"github.com/HuguesGuilleus/site-gafam/service/discogs"
	"github.com/HuguesGuilleus/site-gafam/service/front"
	"github.com/HuguesGuilleus/site-gafam/service/instagram"
	"github.com/HuguesGuilleus/site-gafam/service/lfi"
	"github.com/HuguesGuilleus/site-gafam/service/peertube"
	"github.com/HuguesGuilleus/site-gafam/service/rss"
	"github.com/HuguesGuilleus/site-gafam/service/tiktok"
	"github.com/HuguesGuilleus/site-gafam/service/twitch"
	"github.com/HuguesGuilleus/site-gafam/service/youtube"
	"github.com/HuguesGuilleus/sniffle/tool"
)

func Do(t *tool.Tool, targets map[string][]string) {
	data := make(map[string][]*common.List)
	for _, urls := range targets {
		for _, u := range urls {
			data[u] = nil
		}
	}
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	for u := range data {
		go func(u string) {
			defer wg.Done()
			l := fetchOne(t, u)
			mutex.Lock()
			defer mutex.Unlock()
			data[u] = l
		}(u)
	}
	wg.Add(len(data))
	wg.Wait()

	for title, urls := range targets {
		index := mergeIndex(title, urls, data)
		front.Render(t, strings.ToLower(title), &index)
	}
	front.RenderTitles(t, targets)
}

func mergeIndex(title string, urls []string, data map[string][]*common.List) (index common.Index) {
	index.Title = title
	index.Lists = make([]*common.List, 0)

	for _, u := range urls {
		index.Lists = append(index.Lists, data[u]...)
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
	index.News = slices.CompactFunc(index.News, func(a, b *common.Item) bool { return a.URL == b.URL })

	return
}

func fetchOne(t *tool.Tool, u string) []*common.List {
	defer t.Log(context.Background(), slog.LevelInfo+2, "target", "u", u)
	proto, id, _ := strings.Cut(u, ":")
	proto, _, _ = strings.Cut(proto, "#")
	id, _, _ = strings.Cut(id, "#")
	switch proto {
	case "arte.cat":
		return arte.Category(t, id)
	case "arte.ch":
		return []*common.List{arte.Channel(t, id)}
	case "arte.li":
		return arte.List(t, id)
	case "discogs":
		return []*common.List{discogs.ArtistStrict(t, id)}
	case "discogs+":
		return []*common.List{discogs.ArtistExtra(t, id)}
	case "insta.ch":
		return []*common.List{instagram.User(t, id)}
	case "insta.tr+ch":
		return []*common.List{instagram.WithThread(t, id)}
	case "lfi.g":
		return []*common.List{lfi.Group(t, id)}
	case "peertube.a":
		return []*common.List{peertube.User(t, id)}
	case "peertube.c":
		return []*common.List{peertube.Channel(t, id)}
	case "rss":
		return []*common.List{rss.Fetch(t, id)}
	case "tiktok.ch":
		return []*common.List{tiktok.Channel(t, id)}
	case "twitch.ch":
		return []*common.List{twitch.Channel(t, id)}
	case "twitch.te":
		return twitch.Team(t, id)
	case "yt.charts.titles":
		return []*common.List{youtube.ChartsTitles(t, id)}
	case "yt.ch":
		return []*common.List{youtube.ChannelRSS(t, id)}
	case "yt.pl":
		return []*common.List{youtube.PlaylistRSS(t, id)}
	default:
		t.Warn("unknown.urlproto", "proto", proto, "id", id)
		return nil
	}
}
