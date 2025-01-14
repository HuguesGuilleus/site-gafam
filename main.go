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
		duration := map[string]time.Duration{
			"api-cdn.arte.tv":           time.Hour * 24 * 365,
			"api.arte.tv":               time.Hour * 24 * 30,
			"www.instagram.com":         time.Hour * 24,
			"www.tiktok.com":            time.Hour * 24,
			"img.youtube.com":           time.Hour * 24 * 30,
			"lh3.googleusercontent.com": time.Hour * 24 * 30,
			"static-cdn.jtvnw.net":      time.Hour * 24 * 30,
		}[m.URL.Host]
		if duration > 0 {
			return duration
		}

		switch {
		case strings.HasSuffix(m.URL.Host, ".akamaized.net"): // Arte
			return time.Hour * 24 * 365
		case strings.HasSuffix(m.URL.Host, ".cdninstagram.com"): // Instagram
			return time.Hour * 24 * 30
		case strings.HasPrefix(m.URL.Path, "/lazy-static/thumbnails/") && strings.HasSuffix(m.URL.Path, ".jpg"): // Peertube
			return time.Hour * 24 * 60
		case strings.HasSuffix(m.URL.Host, ".tiktokcdn.com"): // Tiktok
			return time.Hour * 24 * 30
		}

		return time.Hour * 2
	})

	t := tool.New(&tool.Config{
		Logger:    slog.New(myhandler.New(os.Stderr, slog.LevelInfo)),
		Writefile: writefile.Os("public"),
		Fetcher: []fetch.Fetcher{
			fetch.Cache("cache"),
			fetch.Net(nil, "cache", map[string]time.Duration{
				"":                   time.Second / 2,
				"actionpopulaire.fr": 0,
				"api-cdn.arte.tv":    time.Second * 3,
				"api.arte.tv":        time.Second * 30,
				"www.instagram.com":  time.Second * 2,
				"www.youtube.com":    time.Second / 100 * 0,

				"arte-uhd-cmafhls.akamaized.net":   0 * time.Second / 20,
				"img.youtube.com":                  0 * time.Second / 20,
				"p16-sign-useast2a.tiktokcdn.com":  0 * time.Second / 20,
				"scontent-cdg4-1.cdninstagram.com": 0 * time.Second / 20,
				"scontent-cdg4-2.cdninstagram.com": 0 * time.Second / 20,
				"scontent-cdg4-3.cdninstagram.com": 0 * time.Second / 20,
				"scontent-cdg4-4.cdninstagram.com": 0 * time.Second / 20,
			}),
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
