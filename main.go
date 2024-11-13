package main

import (
	"context"
	"flag"
	"frontend-gafam/service"
	"frontend-gafam/youtube"
	"log/slog"
	"os"
	"sniffle/myhandler"
	"sniffle/tool"
	"sniffle/tool/fetch"
	"sniffle/tool/writefile"
	"time"

	"github.com/BurntSushi/toml"
)

var ytTodo = youtube.Todo{}

func main() {
	// Clear cache
	cc := flag.Bool("cc", false, "Clear the cache")
	flag.Parse()
	if *cc {
		entrys, _ := os.ReadDir("cache/https/")
		for _, entry := range entrys {
			if entry.Name() == "img.youtube.com" {
				continue
			}
			os.RemoveAll("cache/https/" + entry.Name())
		}
	}

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
