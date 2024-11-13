package instagram

import (
	"bytes"
	"frontend-gafam/service/common"
	"net/url"
	"sniffle/tool"
	"sniffle/tool/fetch"
	"strconv"
	"time"
)

func User(t *tool.Tool, id string) (list *common.List) {
	list = fetchChannel(t, id)
	if list == nil {
		return nil
	}

	request := fetch.Request{URL: &url.URL{
		Scheme:   "https",
		Host:     "www.instagram.com",
		Path:     "/graphql/query",
		RawQuery: `query_hash=56a7068fea504063273cc2120ffd54f3&variables={"id":"` + id + `","first":"24"}`,
	}}

	list.JSON = bytes.Join([][]byte{
		[]byte("["),
		list.JSON,
		[]byte(","),
		tool.FetchAll(t, &request),
		[]byte("]"),
	}, nil)

	list.Items = trItems(t, &request)

	return list
}

func fetchChannel(t *tool.Tool, id string) *common.List {
	request := fetch.Request{URL: &url.URL{
		Scheme:   "https",
		Host:     "www.instagram.com",
		Path:     "/graphql/query",
		RawQuery: `doc_id=9539110062771438&variables={"id":"` + id + `","render_surface":"PROFILE"}`,
	}}

	dto := struct {
		Data struct {
			User struct {
				Username     string
				Full_name    string
				External_url string
				Biography    string
			}
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, &request) {
		return nil
	}
	user := dto.Data.User

	description := common.Description(user.Biography)
	if user.External_url != "" {
		description = append(description, user.External_url)
	}
	description = append(description, "raw ID: "+id)

	return &common.List{
		Host:        "insta",
		ID:          user.Username,
		URL:         "https://www.instagram.com/" + user.Username + "/",
		Title:       user.Full_name,
		Description: description,
		JSON:        tool.FetchAll(t, &request),
	}
}

func trItems(t *tool.Tool, request *fetch.Request) []*common.Item {
	dto := struct {
		Status string
		Data   struct {
			User struct {
				Edge_owner_to_timeline_media struct {
					Edges []struct {
						Node struct {
							ID          string
							Shortcode   string
							Display_url string
							Dimensions  struct {
								Height int
								Width  int
							}
							Edge_media_to_caption struct {
								Edges [1]struct {
									Node struct {
										Text string
									}
								}
							}
							Edge_sidecar_to_children struct {
								Edges []struct {
									Node struct {
										Display_url string
									}
								}
							}
							Video_url               string
							Taken_at_timestamp      int64
							Edge_media_preview_like struct {
								Count int
							}
							Owner struct {
								Username string
							}
						}
					}
				}
			}
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, request) {
		return nil
	}

	edges := dto.Data.User.Edge_owner_to_timeline_media.Edges
	items := make([]*common.Item, len(edges))
	for i, edge := range edges {
		node := edge.Node
		poster := tool.FetchAll(t, fetch.R("", edge.Node.Display_url, nil))

		posterAnnex := ([][]byte)(nil)
		if len(edge.Node.Edge_sidecar_to_children.Edges) > 1 {
			posterAnnex = make([][]byte, len(edge.Node.Edge_sidecar_to_children.Edges)-1)
			for i, annex := range edge.Node.Edge_sidecar_to_children.Edges[1:] {
				posterAnnex[i] = tool.FetchAll(t, fetch.R("", annex.Node.Display_url, nil))
			}
		}

		sources := ([]common.Source)(nil)
		if edge.Node.Video_url != "" {
			sources = []common.Source{{
				URL:    edge.Node.Video_url,
				Height: node.Dimensions.Height,
			}}
		}

		owner := node.Owner.Username
		items[i] = &common.Item{
			ID:          edge.Node.ID,
			Host:        "instagram",
			URL:         "https://www.instagram.com/" + owner + "/p/" + node.Shortcode + "/",
			Description: common.Description(node.Edge_media_to_caption.Edges[0].Node.Text),
			Author:      owner,
			Published:   time.Unix(node.Taken_at_timestamp, 0),
			Like:        node.Edge_media_preview_like.Count,
			IsVideo:     len(sources) != 0,

			Poster:       poster,
			PosterAnnex:  posterAnnex,
			PosterWidth:  strconv.Itoa(node.Dimensions.Width),
			PosterHeight: strconv.Itoa(node.Dimensions.Height),
			Sources:      sources,
		}
	}
	return items
}
