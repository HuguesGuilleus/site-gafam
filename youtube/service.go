package youtube

import (
	"cmp"
	"context"
	"frontend-gafam/asset"
	"slices"
	"sniffle/tool"
	"sniffle/tool/render"
	"strconv"
	"strings"
	"time"
)

type Todo struct {
	Base     string
	Since    time.Duration
	Channel  []string
	Playlist []string
}

func Do(t *tool.Tool, todo *Todo) {
	todo.Base = strings.TrimRight(todo.Base, "/") + "/" // be sure to add only one '/'

	index := []*Index{}
	for _, id := range todo.Channel {
		index = append(index, FetchRSS(t, true, id))
	}
	for _, id := range todo.Playlist {
		index = append(index, FetchRSS(t, false, id))
	}
	index = slices.DeleteFunc(index, func(index *Index) bool { return index == nil })
	slices.SortStableFunc(index, func(a, b *Index) int { return cmp.Compare(a.TitleLower, b.TitleLower) })

	after := time.Now().Add(-todo.Since)
	news := make([]IndexVideoItem, 0)
	for _, index := range index {
		for _, video := range index.Items {
			SaveVideoImage(t, todo.Base, video.Id)
			if video.Published.After(after) {
				news = append(news, video)
			}
		}
		RenderIndex(t, todo.Base, index)
	}

	renderAll(t, todo.Base, news, index)
}

// Fetch and save the thunbail of the video
// Save in `/youtube/icon/{id}.jpg`
func SaveVideoImage(t *tool.Tool, base, id string) {
	t.WriteFile(base+"vi/"+id+".jpg",
		tool.FetchAll(context.Background(), t, "", "https://img.youtube.com/vi/"+id+"/hqdefault.jpg", nil, nil))
}

func renderAll(t *tool.Tool, base string, news []IndexVideoItem, index []*Index) {
	t.WriteFile(base+"index.html", render.Merge(render.No("html", render.A("lang", "fr"),
		render.N("head", asset.Begin, render.N("title", "Index")),
		render.N("body",
			render.N("header.c", render.N("div.title", "Index")),
			render.N("main.and2toc",
				render.N("ul.toc",
					render.N("li", "(Total: ", len(index), ")"),
					render.Slice(index, func(_ int, index *Index) render.Node {
						return render.N("li", render.No("a", render.A("href", "#"+index.Id), index.Title))
					}),
				),
				render.N("div",
					videosCarousel(news),
					render.Slice(index, func(_ int, index *Index) render.Node {
						return render.N("",
							render.No("h1", render.A("id", index.Id),
								render.No("a.copy", render.A("href", "https://youtube.com/channel/"+index.Id), index.Id),
								" ", index.Title, " ",
								render.No("a", render.A("href", index.OutID+".html"), "~>")),
							videosCarousel(index.Items),
						)
					}),
				),
			),
		)),
	))
}

func videosCarousel(videos []IndexVideoItem) render.Node {
	slices.SortFunc(videos, func(a, b IndexVideoItem) int { return b.Published.Compare(a.Published) })
	return render.N("div.imgs", render.Slice(videos, func(_ int, video IndexVideoItem) render.Node {
		return render.No("img.copy", render.
			A("src", "vi/"+video.Id+".jpg").
			A("loading", "lazy").
			A("title", video.Title+" @"+video.AuthorName+" ["+video.Published.Format("2006-01-02")+"] vue: "+strconv.FormatUint(uint64(video.View), 10)).
			A("data-href", "https://youtu.be/"+video.Id),
		)
	}))
}
