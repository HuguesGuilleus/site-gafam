package front

import (
	"fmt"
	"frontend-gafam/service/common"
	"sniffle/tool"
	"sniffle/tool/render"
	"strings"
	"time"
)

func Render(t *tool.Tool, base string, index *common.Index) {
	base = strings.TrimRight(base, "/")

	renderIndex(t, base, index)

	for _, list := range index.Lists {
		t.WriteFile(base+"/"+list.ID+".json", list.JSON)
		renderChannel(t, base, list)
		for _, item := range list.Items {
			t.WriteFile(base+"/_icon/"+item.ID+".jpg", item.Poster)
		}
	}
}

func RenderTitles(t *tool.Tool, titles map[string][]string) {
	t.WriteFile("/index.html", render.Merge(render.Na("html", "lang", "fr").N(
		render.N("head", begin, render.N("title", "Index")),
		render.N("body",
			render.N("header", render.N("div.title", "Index")),
			render.N("main", render.Map(titles, func(title string, _ []string) render.Node {
				return render.N("h1", title, " ", render.Na("a", "href", strings.ToLower(title)+"/index.html").N("~>"))
			})),
		),
	)))
}

func renderIndex(t *tool.Tool, base string, index *common.Index) {
	t.WriteFile(base+"/index.html", render.Merge(render.Na("html", "lang", "fr").N(
		render.N("head", begin, render.N("title", index.Title)),
		render.N("body",
			render.N("header.withToc", render.N("div.title", render.Na("a", "href", "../index.html").N("<~"), " ", index.Title)),
			render.N("main.withToc",
				render.N("ul.toc",
					render.N("li", "(Total: ", len(index.Lists), ")"),
					render.S(index.Lists, "", func(list *common.List) render.Node {
						return render.N("li", render.Na("a", "href", "#"+list.ID).N(list.Title))
					}),
				),
				render.N("div",
					carousel(index.News),
					render.S(index.Lists, "", func(list *common.List) render.Node {
						return render.N("",
							render.Na("h1", "id", list.ID).N(
								render.Na("a.copy", "href", list.URL).N(list.ID),
								" ", list.Title, " ",
								render.Na("a", "href", list.ID+".html").N("~>"),
							),
							carousel(list.Items),
						)
					}),
				),
			),
		),
	)))
}

func renderChannel(t *tool.Tool, base string, list *common.List) {
	t.WriteFile(base+"/"+list.ID+".html", render.Merge(render.Na("html", "lang", "fr").N(
		render.N("head", begin, render.N("title", list.Title)),
		render.N("body",
			render.N("header",
				render.N("div.title", render.Na("a", "href", "index.html").N("<~"), " ", list.Title),
				render.N("p", render.Na("a.copy", "href", list.URL).N(list.ID)),
				renderDescription(list.Description),
			),
			render.N("main",
				carousel(list.Items),
				render.N("ul.items", render.S(list.Items, "", func(item *common.Item) render.Node {
					return render.N("li.item",
						render.Na("img", "src", "_icon/"+item.ID+".jpg").
							A("width", item.PosterWidth).
							A("height", item.PosterHeight).
							A("loading", "lazy").
							N(),
						render.N("div.title",
							render.Na("a.copy", "href", item.URL).N(item.ID),
							" ", item.Title,
						),
						render.N("div.meta",
							"[ like: ", item.Like, " | vue: ", item.View,
							render.IfS(item.Duration != 0, render.N("", " | ", item.Duration)),
							" ] @",
							item.Author,
							item.Published.In(time.Local).Format(" (2006-01-02 15:04:05)"),
						),
						render.N("div", render.S(item.Sources, " ", func(s common.Source) render.Node {
							return render.Na("a.copy", "href", s.URL).
								N(render.Int(s.Height), "p")
						})),
						renderDescription(item.Description),
					)
				})),
			),
		),
	)))
}

func carousel(items []*common.Item) render.Node {
	return render.N("div.imgs", render.S(items, "", func(item *common.Item) render.Node {
		return render.Na("a.copy.wi", "href", item.URL).
			N(render.Na("img", "src", "_icon/"+item.ID+".jpg").
				A("width", item.PosterWidth).
				A("height", item.PosterHeight).
				A("loading", "lazy").
				A("title", fmt.Sprintf("%s @%s [%s] vue: %d", item.Title, item.Author, item.Published.Format("2006-01-02"), item.View)).N(),
			)
	}))
}

func renderDescription(description []string) []render.Node {
	return render.S(description, "", func(line string) render.Node {
		if line == "" {
			return render.N("div.emptyline")
		}
		return render.N("p", line)
	})
}
