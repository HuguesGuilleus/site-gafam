package main

import (
	"context"
	"frontend-gafam/youtube"
	"log/slog"
	"os"
	"sniffle/myhandler"
	"sniffle/tool"
	"sniffle/tool/fetch"
	"sniffle/tool/writefile"
	"time"
)

var ytTodo = youtube.Todo{}

func main() {
	t := tool.New(&tool.Config{
		Logger:    slog.New(myhandler.New(os.Stderr, slog.LevelInfo)),
		HostURL:   "localhost",
		Writefile: writefile.Os("public"),
		Fetcher: []fetch.Fetcher{
			fetch.CacheOnly("cache"),
			fetch.Net(nil, "cache", 1, time.Millisecond*100),
		},
		LongTasksCache: writefile.Os("cache"),
		LongTasksMap:   map[string]func([]byte) ([]byte, error){},
	})

	defer func(begin time.Time) {
		t.Logger.Log(context.Background(), slog.LevelInfo+2, "duration", "d", time.Since(begin))
	}(time.Now())

	youtube.Do(t, &ytTodo)
}
