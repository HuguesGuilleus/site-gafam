package instagram

import (
	"bytes"
	"frontend-gafam/service/common"
	"sniffle/tool"
	"sniffle/tool/fetch"
	"strings"
)

func WithThread(t *tool.Tool, id string) (list *common.List) {
	list = fetchChannel(t, id)
	if list == nil {
		return nil
	}

	list.Items = fetchItems(t, id)
	list.Items = append(list.Items, fetchThreads(t, fetchThreadsPostRequest(id))...)
	list.Items = append(list.Items, fetchThreads(t, fetchThreadsRequestRepost(id))...)
	common.SortByDate(list.Items)

	list.JSON = bytes.Join([][]byte{
		[]byte(`{"channel":`),
		tool.FetchAll(t, fetchChannelRequest(id)),
		[]byte(`,"instaItems":`),
		tool.FetchAll(t, fetchItemsRequest(id)),
		[]byte(`,"threadsPost":`),
		tool.FetchAll(t, fetchThreadsPostRequest(id)),
		[]byte(`,"threadsRepost":`),
		tool.FetchAll(t, fetchThreadsRequestRepost(id)),
		[]byte(`}`),
	}, nil)

	return list
}

func fetchThreads(t *tool.Tool, r *fetch.Request) []*common.Item {
	dto := struct {
		Data struct {
			MediaData struct {
				Edges []struct {
					Node struct {
						Thread_items [1]struct {
							Post struct {
								Code               string
								Text_post_app_info struct {
									Text_fragments struct {
										Fragments []struct {
											Plaintext string
										}
									}
								}
								Image_versions2 struct {
									Candidates []struct {
										Height int
										URL    string
									}
								}
							}
						}
					}
				}
			}
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, r) {
		return nil
	}

	list := make([]*common.Item, len(dto.Data.MediaData.Edges))
	for i, edge := range dto.Data.MediaData.Edges {
		post := edge.Node.Thread_items[0].Post

		texts := make([]string, len(post.Text_post_app_info.Text_fragments.Fragments))
		for i, frag := range post.Text_post_app_info.Text_fragments.Fragments {
			texts[i] = frag.Plaintext
		}

		item := &common.Item{
			Host:        "threads",
			ID:          post.Code,
			URL:         "https://www.threads.net/post/" + post.Code,
			Description: texts,
		}

		imgURL := ""
		imgHeight := 0
		for _, img := range post.Image_versions2.Candidates {
			if imgHeight < img.Height {
				imgHeight = img.Height
				imgURL = img.URL
			}
		}
		if imgURL != "" {
			item.Poster, item.PosterWidth, item.PosterHeight = common.FetchPoster(t, imgURL)
		} else {
			item.Poster, item.PosterWidth, item.PosterHeight = common.TextPoster(strings.Join(item.Description, "\n"))
		}

		list[i] = item
	}

	return list
}

func fetchThreadsPostRequest(id string) *fetch.Request {
	return fetch.Rs("POST", `https://www.threads.net/api/graphql`,
		`&variables=%7B%22userID%22%3A%22`+id+`%22%2C%22__relay_internal__pv__BarcelonaIsLoggedInrelayprovider%22%3Afalse%2C%22__relay_internal__pv__BarcelonaShareableListsrelayprovider%22%3Atrue%2C%22__relay_internal__pv__BarcelonaIsInlineReelsEnabledrelayprovider%22%3Atrue%2C%22__relay_internal__pv__BarcelonaOptionalCookiesEnabledrelayprovider%22%3Atrue%2C%22__relay_internal__pv__BarcelonaShowReshareCountrelayprovider%22%3Atrue%2C%22__relay_internal__pv__BarcelonaQuotedPostUFIEnabledrelayprovider%22%3Afalse%2C%22__relay_internal__pv__BarcelonaIsCrawlerrelayprovider%22%3Afalse%2C%22__relay_internal__pv__BarcelonaHasDisplayNamesrelayprovider%22%3Afalse%2C%22__relay_internal__pv__BarcelonaCanSeeSponsoredContentrelayprovider%22%3Afalse%2C%22__relay_internal__pv__BarcelonaShouldShowFediverseM075Featuresrelayprovider%22%3Afalse%2C%22__relay_internal__pv__BarcelonaIsInternalUserrelayprovider%22%3Afalse%7D`+
			`&server_timestamps=true`+
			`&doc_id=9061509793918288`,
		"User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0",
		"Content-Type", "application/x-www-form-urlencoded",
		"Sec-Fetch-Site", "same-origin",
	)
}

func fetchThreadsRequestRepost(id string) *fetch.Request {
	return fetch.Rs("POST", `https://www.threads.net/api/graphql`,
		`&variables=%7B%22userID%22%3A%22`+id+`%22%2C%22__relay_internal__pv__BarcelonaIsLoggedInrelayprovider%22%3Afalse%2C%22__relay_internal__pv__BarcelonaShareableListsrelayprovider%22%3Atrue%2C%22__relay_internal__pv__BarcelonaIsInlineReelsEnabledrelayprovider%22%3Atrue%2C%22__relay_internal__pv__BarcelonaOptionalCookiesEnabledrelayprovider%22%3Atrue%2C%22__relay_internal__pv__BarcelonaShowReshareCountrelayprovider%22%3Atrue%2C%22__relay_internal__pv__BarcelonaQuotedPostUFIEnabledrelayprovider%22%3Afalse%2C%22__relay_internal__pv__BarcelonaIsCrawlerrelayprovider%22%3Afalse%2C%22__relay_internal__pv__BarcelonaHasDisplayNamesrelayprovider%22%3Afalse%2C%22__relay_internal__pv__BarcelonaCanSeeSponsoredContentrelayprovider%22%3Afalse%2C%22__relay_internal__pv__BarcelonaShouldShowFediverseM075Featuresrelayprovider%22%3Afalse%2C%22__relay_internal__pv__BarcelonaIsInternalUserrelayprovider%22%3Afalse%7D`+
			`&doc_id=9876187475731224`,
		"User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0",
		"Content-Type", "application/x-www-form-urlencoded",
		"Sec-Fetch-Site", "same-origin",
	)
}
