package rss

import (
	"context"
	"encoding/xml"
	"frontend-gafam/asset"
	"sniffle/tool"
	"sniffle/tool/render"
	"time"
)

func Do(t *tool.Tool) {
	Fetch(t, "https://video.blast-info.fr/feeds/videos.xml?videoChannelId=2")
}

func Fetch(t *tool.Tool, u string) any {
	x := tool.FetchAll(context.Background(), t, "", u, nil, nil)

	dto := struct {
		Channel struct {
			Title       string   `xml:"title"`
			Description string   `xml:"description"`
			URL         []string `xml:"link"`
			Items       []struct {
				Title       string  `xml:"title"`
				Description string  `xml:"description"`
				URL         string  `xml:"link"`
				Pub         rssTime `xml:"pubDate"`
				Thumbnail   struct {
					Url string `xml:"url,attr"`
				} `xml:"thumbnail"`
				Sources struct {
					Content []struct {
						Height int    `xml:"height,attr"`
						Url    string `xml:"url,attr"`
					} `xml:"content"`
				} `xml:"group"`
			} `xml:"item"`
		} `xml:"channel"`
	}{}
	if err := xml.Unmarshal(x, &dto); err != nil {
		t.Warn("xml.decode", "url", u, "err", err.Error())
		return nil
	}

	ch := Channel{
		Title:       dto.Channel.Title,
		Description: dto.Channel.Description,
		URL:         dto.Channel.URL[0],
		Items:       make([]Video, len(dto.Channel.Items)),
	}
	for i, item := range dto.Channel.Items {
		src := make([]Source, len(item.Sources.Content))
		for i, s := range item.Sources.Content {
			src[i] = Source{Height: 0, Url: s.Url}
			src[i] = Source{Height: s.Height, Url: s.Url}
		}
		ch.Items[i] = Video{
			Title:       item.Title,
			Description: item.Description,
			URL:         item.URL,
			Pub:         item.Pub.Time,
			Poster:      item.Thumbnail.Url, ///////////////////
			Sources:     src,
		}
	}

	WriteChannel(t, "yt", &ch)

	return nil
}

func WriteChannel(t *tool.Tool, base string, ch *Channel) {
	t.WriteFile(base+"/rss.html", render.Merge(render.No("html", render.A("lang", "fr"),
		render.N("head", asset.Begin, render.N("title", ch.Title)),
		render.N("body",
			render.N("header",
				render.N("div.title", ch.Title),
				render.N("p", ch.Description),
				render.N("p", render.No("a.copy", render.A("href", ch.URL), ch.URL)),
			),
			render.N("main",
				render.N("ul.items", render.Slice(ch.Items, func(_ int, v Video) render.Node {
					return render.N("li.item",
						render.No("img", render.A("src", v.Poster)),
						render.N("div.title", render.No("a.copy", render.A("href", v.URL), "lien"), " ", v.Title),
						render.N("div.meta", v.Pub.In(time.Local).Format(" (2006-01-02 15:04:05)")),
						render.N("div", render.SliceSeparator(v.Sources, " ", func(_ int, s Source) render.Node {
							return render.No("a.copy", render.A("href", s.Url), render.Int(s.Height), "p")
						})),
					)
				})),
			),
		),
	)))
}

type Channel struct {
	Title       string
	Description string
	URL         string
	Items       []Video
}
type Video struct {
	Title       string
	Description string
	URL         string
	Pub         time.Time
	Poster      string
	Sources     []Source
}
type Source struct {
	Url    string
	Height int
}

type rssTime struct {
	time.Time
}

func (t *rssTime) UnmarshalText(text []byte) (err error) {
	t.Time, err = time.Parse("Mon, 02 Jan 2006 15:04:05 MST", string(text))
	return
}
