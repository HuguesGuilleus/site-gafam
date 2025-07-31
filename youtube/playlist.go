package youtube

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/HuguesGuilleus/site-gafam/asset"
	"github.com/HuguesGuilleus/sniffle/tool"
	"github.com/HuguesGuilleus/sniffle/tool/fetch"
	"github.com/HuguesGuilleus/sniffle/tool/render"
)

type Playlist struct {
	Id          string
	Title       string
	Description string
	Items       []VideoItem
}

func FetchPlaylist(t *tool.Tool, id string) *Playlist {
	body := tool.FetchAll(t, fetch.R("", "https://www.youtube.com/playlist?list="+id, nil))

	_, src, ok := bytes.Cut(body, []byte("var ytInitialData = "))
	if !ok {
		t.Warn("wrong HTML format for playlist", "id", id)
		return nil
	}
	src, _, ok = bytes.Cut(src, []byte(";</script>"))
	if !ok {
		t.Warn("wrong HTML format for playlist", "id", id)
		return nil
	}

	t.WriteFile("/youtube/list/"+id+".json", src)

	dto := struct {
		Metadata struct {
			PlaylistMetadataRenderer struct {
				Title       string
				Description string
			}
		}
		Contents struct {
			TwoColumnBrowseResultsRenderer struct {
				Tabs []struct {
					TabRenderer struct {
						Content struct {
							SectionListRenderer struct {
								Contents []struct {
									ItemSectionRenderer struct {
										Contents []struct {
											PlaylistVideoListRenderer struct {
												Contents []struct {
													PlaylistVideoRenderer playlistItemDTO
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}{}
	if err := json.Unmarshal(src, &dto); err != nil {
		t.Warn("err", "id", id, "err", err.Error())
		return nil
	}

	items := []VideoItem{}
	for _, v := range dto.Contents.TwoColumnBrowseResultsRenderer.Tabs {
		for _, v := range v.TabRenderer.Content.SectionListRenderer.Contents {
			for _, v := range v.ItemSectionRenderer.Contents {
				for _, v := range v.PlaylistVideoListRenderer.Contents {
					items = append(items, v.PlaylistVideoRenderer.VideoItem())
				}
			}
		}
	}

	return &Playlist{
		Id:          id,
		Title:       dto.Metadata.PlaylistMetadataRenderer.Title,
		Description: dto.Metadata.PlaylistMetadataRenderer.Description,
		Items:       items,
	}
}

type ChannelItem struct {
	Name string
	Tag  string
	Id   string
}

type VideoItem struct {
	Id          string
	Title       string
	Author      ChannelItem
	Duration    time.Duration
	Description string
}

type playlistItemDTO struct {
	VideoId string
	Title   struct {
		Runs [1]struct{ Text string }
	}
	ShortBylineText struct {
		Runs [1]struct {
			Text               string
			NavigationEndpoint struct {
				BrowseEndpoint struct {
					BrowseId         string
					CanonicalBaseUrl string
				}
			}
		}
	}
	LengthSeconds int `json:"lengthSeconds,string"`
}

func (dto *playlistItemDTO) VideoItem() VideoItem {
	author := dto.ShortBylineText.Runs[0]
	return VideoItem{
		Id:    dto.VideoId,
		Title: dto.Title.Runs[0].Text,
		Author: ChannelItem{
			Name: author.Text,
			Tag:  strings.TrimLeft(author.NavigationEndpoint.BrowseEndpoint.CanonicalBaseUrl, "/@"),
			Id:   author.NavigationEndpoint.BrowseEndpoint.BrowseId,
		},
		Duration: time.Duration(dto.LengthSeconds) * time.Second,
	}
}

// Fetch playlist item and render it.
func RenderPlaylist(t *tool.Tool, id string) {
	p := FetchPlaylist(t, id)
	if p == nil {
		return
	}

	for _, video := range p.Items {
		FetchVideoImage(t, video.Id)
	}

	t.WriteFile("/youtube/list/"+id+".html", render.Merge(render.Na("html", "lang", "fr").N(
		render.N("head",
			asset.Begin,
			render.N("title", p.Title),
		),
		render.N("body",
			render.N("header",
				render.N("div.title", p.Title),
				render.Na("a.copy", "href", "https://www.youtube.com/playlist?list="+id).N(p.Id),
				render.If(p.Description != "", func() render.Node {
					return render.N("p", p.Description)
				}),
			),
			render.N("ul.items", render.S(p.Items, "", func(video VideoItem) render.Node {
				return render.N("li.item",
					render.Na("img", "src", "../icon/"+video.Id+".jpg"),
					render.N("div",
						render.N("div",
							render.Na("a.copy", "href", "https://youtu.be/"+video.Id).N(video.Id),
							" ", video.Title),
						"@", video.Author.Tag, " (", video.Duration, ")",
					),
				)
			})),
		)),
	))
}
