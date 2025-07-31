package youtube

import (
	"encoding/json"
	"encoding/xml"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/HuguesGuilleus/site-gafam/asset"
	"github.com/HuguesGuilleus/sniffle/tool"
	"github.com/HuguesGuilleus/sniffle/tool/fetch"
	"github.com/HuguesGuilleus/sniffle/tool/render"
)

type Index struct {
	IsChannel  bool
	Id         string
	Title      string
	TitleLower string
	OutID      string

	data []byte

	Items []IndexVideoItem
}

type IndexVideoItem struct {
	Id          string
	Title       string
	AuthorId    string
	AuthorName  string
	Published   time.Time
	Updated     time.Time
	Duration    time.Duration
	Description []string
	View        uint
	Like        uint
}

func FetchRSS(t *tool.Tool, isChannel bool, id string) *Index {
	url := ""
	if isChannel {
		url = "https://www.youtube.com/feeds/videos.xml?channel_id=" + id
	} else {
		url = "https://www.youtube.com/feeds/videos.xml?playlist_id=" + id
	}

	data := tool.FetchAll(t, fetch.R("", url, nil))
	dto := struct {
		Title      string    `xml:"title"`
		ID         string    `xml:"channelId"`
		AuthorName string    `xml:"author>name"`
		Published  time.Time `xml:"published"`
		Entry      []struct {
			Id          string    `xml:"videoId"`
			Title       string    `xml:"title"`
			AuthorURI   string    `xml:"author>uri"`
			AuthorName  string    `xml:"author>name"`
			Published   time.Time `xml:"published"`
			Updated     time.Time `xml:"updated"`
			Description string    `xml:"group>description"`

			View struct {
				V uint `xml:"views,attr"`
			} `xml:"group>community>statistics"`
			Like struct {
				V uint `xml:"count,attr"`
			} `xml:"group>community>starRating"`
		} `xml:"entry"`
	}{}
	if err := xml.Unmarshal(data, &dto); err != nil {
		t.Warn("rss.err", "isChannel", isChannel, "id", id, "err", err.Error())
		return nil
	}

	outID := strings.ToLower(strings.Join(strings.FieldsFunc(dto.Title, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}), ""))

	items := make([]IndexVideoItem, len(dto.Entry))
	for i, entry := range dto.Entry {
		description := strings.Split(entry.Description, "\n")
		for i, d := range description {
			if !strings.ContainsFunc(d, func(r rune) bool {
				switch r {
				case '-', '_', ' ', '\t':
					return false
				}
				return true
			}) {
				description[i] = ""
			}
		}
		description = slices.CompactFunc(description, func(a, b string) bool { return a == "" && b == "" })

		items[i] = IndexVideoItem{
			Id:          entry.Id,
			Title:       entry.Title,
			AuthorId:    strings.TrimLeft(entry.AuthorURI, "https://www.youtube.com/channel/"),
			AuthorName:  entry.AuthorName,
			Published:   entry.Published,
			Updated:     entry.Updated,
			Description: description,
			View:        entry.View.V,
			Like:        entry.Like.V,
		}
	}

	return &Index{
		IsChannel:  isChannel,
		Id:         id,
		Title:      dto.Title,
		TitleLower: strings.ToLower(dto.Title),
		OutID:      outID,
		Items:      items,
		data:       data,
	}
}

func RenderIndex(t *tool.Tool, base string, index *Index) {
	saveRss2json(t, base, index)
	t.WriteFile(base+index.OutID+".html", render.Merge(render.Na("html", "lang", "fr").N(
		render.N("head", asset.Begin, render.N("title", "Index")),
		render.N("body",
			render.N("header",
				render.N("div.title", render.Na("a", "href", "index.html").N("<~"), " Index"),
				render.N("p", render.IfElse(index.IsChannel,
					func() render.Node {
						return render.Na("a.copy", "href", "https://youtube.com/channel/"+index.Id).N(index.Id)
					},
					func() render.Node {
						return render.Na("a.copy", "href", "https://www.youtube.com/playlist?list="+index.Id).N(index.Id)
					},
				)),
			),
			render.N("main",
				videosCarousel(index.Items),
				render.N("ul.items", render.S(index.Items, "", func(video IndexVideoItem) render.Node {
					return render.N("li.item",
						render.Na("img", "src", "vi/"+video.Id+".jpg"),
						render.N("div.title",
							render.Na("a.copy", "href", "https://youtu.be/"+video.Id).N(video.Id),
							" ", video.Title),
						render.N("div.meta",
							"[ like: ", video.Like,
							" | vue: ", video.View, " ] ",
							video.AuthorName,
							video.Published.In(time.Local).Format(" (2006-01-02 15:04:05)"),
						),
						render.S(video.Description, "", func(l string) render.Node {
							if l == "" {
								return render.N("div.emptyline")
							}
							return render.N("p", l)
						}),
					)
				}))),
		),
	)))
}

func saveRss2json(t *tool.Tool, base string, index *Index) {
	type xmlNode struct {
		XMLName xml.Name
		Attrs   []xml.Attr `xml:",any,attr"`
		Content string     `xml:",chardata"`
		Nodes   []xmlNode  `xml:",any"`
	}
	rootDTO := xmlNode{}
	if err := xml.Unmarshal(index.data, &rootDTO); err != nil {
		t.Warn("rss0json.err", "isChannel", index.IsChannel, "id", index.Id, "err", err.Error())
		return
	}

	var walk func(node xmlNode) any
	walk = func(node xmlNode) any {
		items := map[string]any{"!": "[" + node.XMLName.Space + "]: " + node.XMLName.Local}
		if c := strings.TrimSpace(node.Content); c != "" {
			items["%"] = c
		}
		for _, a := range node.Attrs {
			items["$"+a.Name.Local] = a.Value
		}
		if len(node.Nodes) != 0 {
			children := []any{}
			for _, c := range node.Nodes {
				children = append(children, walk(c))
			}
			items["&"] = children
		}
		return items
	}

	if j, err := json.Marshal(walk(rootDTO)); err != nil {
		t.Warn("json(fromRss).err", "isChannel", index.IsChannel, "id", index.Id, "err", err.Error())
		return
	} else {
		t.WriteFile(base+index.OutID+".json", j)
	}
}
