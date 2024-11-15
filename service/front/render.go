package front

import (
	"fmt"
	"frontend-gafam/service/common"
	"maps"
	"slices"
	"sniffle/tool"
	"sniffle/tool/render"
	"strings"
	"time"
)

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

func Render(t *tool.Tool, base string, index *common.Index) {
	base = strings.TrimRight(base, "/")

	renderIndex(t, base, index)

	for _, list := range index.Lists {
		t.WriteFile(base+"/"+list.ID+".json", list.JSON)
		renderChannel(t, base, list)
		for _, item := range list.Items {
			t.WriteFile("/_icon/"+item.Host+"_"+item.ID+".jpg", item.Poster)
			for i, data := range item.PosterAnnex {
				t.WriteFile(fmt.Sprintf("/_icon/%s_%s_%d.jpg", item.Host, item.ID, i), data)
			}
			renderOne(t, item)
		}
	}
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
						return render.N("li", render.Na("a", "href", "#"+list.ID).N(list.Title, " [", list.Host, "]"))
					}),
				),
				render.N("div",
					carouselLatest(index.News),
					render.S(index.Lists, "", func(list *common.List) render.Node {
						return render.N("",
							render.Na("h1", "id", list.ID).N(
								render.Na("a.copy", "href", list.URL).N(list.ID, " [", list.Host, "]"),
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
						render.Na("img", "src", "../_icon/"+item.Host+"_"+item.ID+".jpg").
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
							" (",
							item.Published,
							")",
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

func carouselLatest(items []*common.Item) []render.Node {
	const future = "Future"
	m := make(map[string][]*common.Item)
	now := time.Now()
	for _, item := range items {
		date := future
		if item.Published.Before(now) {
			date = item.Published.Local().Format(time.DateOnly)
		}
		m[date] = append(m[date], item)
	}
	if len(m[future]) > 0 {
		m[future] = slices.CompactFunc(m[future], func(a, b *common.Item) bool {
			return a.URL == b.URL
		})
	}

	keys := slices.Collect(maps.Keys(m))
	slices.Sort(keys)
	slices.Reverse(keys)
	return render.S(keys, "", func(date string) render.Node {
		return render.N("",
			render.N("h2", date),
			carousel(m[date]),
		)
	})
}
func carousel(items []*common.Item) render.Node {
	return render.N("div.imgs", render.S(items, "", func(item *common.Item) render.Node {
		href := "../_" + item.Host + "_" + item.ID + ".html"
		src0 := "../_icon/" + item.Host + "_" + item.ID + ".jpg"
		if len(item.PosterAnnex) == 0 {
			if item.IsVideo {
				return render.Na("a.wi.isVideo", "href", href).N(carrouselOne(item, src0))
			}
			return render.Na("a.wi", "href", href).N(carrouselOne(item, src0))
		}
		return render.Na("a.wi.slides", "href", href).N(
			carrouselOne(item, src0),
			render.S2(item.PosterAnnex, "", func(i int, _ []byte) render.Node {
				return carrouselOne(item, fmt.Sprintf("../_icon/%s_%s_%d.jpg", item.Host, item.ID, i))
			}),
		)
	}))
}
func carrouselOne(item *common.Item, src string) render.Node {
	return render.Na("img.open", "src", src).
		A("width", item.PosterWidth).
		A("height", item.PosterHeight).
		A("loading", "lazy").
		A("title", fmt.Sprintf("%s @%s [%s] vue: %d", item.Title, item.Author, item.Published.Format(time.DateOnly), item.View)).N()
}

func renderOne(t *tool.Tool, item *common.Item) {
	t.WriteFile("/_"+item.Host+"_"+item.ID+".html", render.Merge(render.Na("html", "lang", "fr").N(
		render.N("head", begin, render.N("title", item.Title)),
		render.N("body",
			render.N("header",
				render.N("div.title", render.Na("a", "href", "index.html").N("<~"), " ", item.Title),
				render.N("p", render.Na("a.copy", "href", item.URL).N(item.ID)),
				render.N("div",
					"[ like: ", item.Like, " | vue: ", item.View,
					render.IfS(item.Duration != 0, render.N("", " | ", item.Duration)),
					" ] @",
					item.Author,
					" (",
					item.Published,
					")",
				),
			),
			render.N("main",
				render.N("div.imgs",
					carrouselOne(item, "_icon/"+item.Host+"_"+item.ID+".jpg"),
					render.S2(item.PosterAnnex, "", func(i int, _ []byte) render.Node {
						return carrouselOne(item, fmt.Sprintf("_icon/%s_%s_%d.jpg", item.Host, item.ID, i))
					}),
				),
				render.N("br"),
				render.N("div", render.S(item.Sources, " ", func(s common.Source) render.Node {
					return render.Na("a.copy", "href", s.URL).
						N(render.Int(s.Height), "p")
				})),
				renderDescription(item.Description),
			),
		),
	)))
}

func renderDescription(description []string) []render.Node {
	return render.S(description, "", func(line string) render.Node {
		if line == "" {
			return render.N("div.emptyline")
		}
		return render.N("p", line)
	})
}
