package main

import (
	"context"
	"flag"
	"frontend-gafam/service"
	"log/slog"
	"os"
	"sniffle/tool"
	"sniffle/tool/fetch"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

func main() {
	t := tool.New(tool.CLI(map[string]time.Duration{
		"":                  time.Second / 2,
		"api-cdn.arte.tv":   time.Second * 3,
		"api.arte.tv":       time.Second * 30,
		"www.instagram.com": time.Second * 2,

		"actionpopulaire.fr":             0,
		"arte-uhd-cmafhls.akamaized.net": 0,
		"cdninstagram.com":               0,
		"static-cdn.jtvnw.net":           0,
		"tiktokcdn-eu.com":               0,
		"tiktokcdn-us.com":               0,
		"tiktokcdn.com":                  0,
		"youtube.com":                    0,
	}))

	defer func(begin time.Time) {
		t.Logger.Log(context.Background(), slog.LevelInfo+2, "duration", "d", time.Since(begin))
	}(time.Now())

	fetch.ClearCache(flag.CommandLine.Lookup("cache").Value.String(), canClearCache)

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

func canClearCache(m *fetch.Meta) time.Duration {
	duration := map[string]time.Duration{
		"arte.tv":           time.Hour * 24 * 2,
		"www.instagram.com": time.Hour * 24 * 2,
		"www.tiktok.com":    time.Hour * 24 * 2,

		"api-cdn.arte.tv":           time.Hour * 24 * 365,
		"api.arte.tv":               time.Hour * 24 * 365,
		"img.youtube.com":           time.Hour * 24 * 365,
		"lh3.googleusercontent.com": time.Hour * 24 * 365,
		"static-cdn.jtvnw.net":      time.Hour * 24 * 365,
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
	case strings.HasSuffix(m.URL.Host, ".tiktokcdn.com") || strings.HasSuffix(m.URL.Host, ".tiktokcdn-eu.com") || strings.HasSuffix(m.URL.Host, ".tiktokcdn-us.com"): // Tiktok
		return time.Hour * 24 * 30
	}

	return time.Hour * 2
}
