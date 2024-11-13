package youtube

import (
	"cmp"
	"frontend-gafam/asset"
	"slices"
	"sniffle/tool"
	"sniffle/tool/fetch"
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
		tool.FetchAll(t, fetch.R("", "https://img.youtube.com/vi/"+id+"/hqdefault.jpg", nil)))
}

func renderAll(t *tool.Tool, base string, news []IndexVideoItem, index []*Index) {
	t.WriteFile(base+"index.html", render.Merge(render.Na("html", "lang", "fr").N(
		render.N("head", asset.Begin, render.N("title", "Index")),
		render.N("body",
			render.N("header.c", render.N("div.title", "Index")),
			render.N("main.and2toc",
				render.N("ul.toc",
					render.N("li", "(Total: ", len(index), ")"),
					render.S(index, "", func(index *Index) render.Node {
						return render.N("li", render.Na("a", "href", "#"+index.Id).N(index.Title))
					}),
				),
				render.N("div",
					videosCarousel(news),
					render.S(index, "", func(index *Index) render.Node {
						return render.N("",
							render.Na("h1", "id", index.Id).N(
								render.Na("a.copy", "href", "https://youtube.com/channel/"+index.Id).N(index.Id),
								" ", index.Title, " ",
								render.Na("a", "href", index.OutID+".html").N("~>"),
							),
							videosCarousel(index.Items),
						)
					}),
				),
			),
		),
	)))
}

func videosCarousel(videos []IndexVideoItem) render.Node {
	slices.SortFunc(videos, func(a, b IndexVideoItem) int { return b.Published.Compare(a.Published) })
	return render.N("div.imgs", render.S(videos, "", func(video IndexVideoItem) render.Node {
		return render.Na("a.copy.wi", "href", "https://youtu.be/"+video.Id).N(
			render.Na("img", "src", "vi/"+video.Id+".jpg").
				A("loading", "lazy").
				A("width", "480").
				A("height", "360").
				A("title",
					video.Title+
						" @"+video.AuthorName+
						" ["+video.Published.Format("2006-01-02")+
						"] vue: "+strconv.FormatUint(uint64(video.View), 10),
				),
		)
	}))
}
