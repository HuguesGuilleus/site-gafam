package main

import (
	"context"
	"flag"
	"frontend-gafam/service"
	"log/slog"
	"os"
	"sniffle/myhandler"
	"sniffle/tool"
	"sniffle/tool/fetch"
	"sniffle/tool/writefile"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

func main() {
	flag.Parse()

	fetch.ClearCache("cache", func(m *fetch.Meta) time.Duration {
		u := m.URL
		switch {
		// Instagram
		case u.Scheme == "https" && strings.HasSuffix(u.Host, ".cdninstagram.com"):
			return time.Hour * 24 * 30
		// Peertube
		case u.Scheme == "https" && strings.HasPrefix(u.Path, "/lazy-static/thumbnails/") && strings.HasSuffix(u.Path, ".jpg"):
			return time.Hour * 24 * 30
		// Tiktok
		case u.Scheme == "https" && strings.HasSuffix(u.Host, ".tiktokcdn.com"):
			return time.Hour * 24 * 30
		// Youtube
		case u.Scheme == "https" && u.Host == "img.youtube.com":
			return time.Hour * 24 * 30
		// Twitch
		case u.Scheme == "https" && u.Host == "static-cdn.jtvnw.net" && strings.HasSuffix(u.Path, ".jpg"):
			return time.Hour * 24 * 30
		default:
			return time.Hour * 2
		}
	})

	t := tool.New(&tool.Config{
		Logger:    slog.New(myhandler.New(os.Stderr, slog.LevelInfo)),
		HostURL:   "localhost",
		Writefile: writefile.Os("public"),
		Fetcher: []fetch.Fetcher{
			fetch.Cache("cache"),
			fetch.Net(nil, "cache", time.Millisecond*100*0),
		},
		LongTasksCache: writefile.Os("cache"),
		LongTasksMap:   map[string]func([]byte) ([]byte, error){},
	})

	defer func(begin time.Time) {
		t.Logger.Log(context.Background(), slog.LevelInfo+2, "duration", "d", time.Since(begin))
	}(time.Now())

	targets := map[string][]string{}
	for _, arg := range flag.Args() {
		data, err := os.ReadFile(arg)
		if err != nil {
			t.Logger.Error("err.read", "file", arg, "err", err)
			continue
		} else if err := toml.Unmarshal(data, &targets); err != nil {
			t.Logger.Error("err.toml", "file", arg, "err", err)
			continue
		}
	}
	service.Do(t, targets)
}
