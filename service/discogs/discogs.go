package discogs

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strconv"

	"github.com/HuguesGuilleus/site-gafam/service/common"
	"github.com/HuguesGuilleus/sniffle/tool"
	"github.com/HuguesGuilleus/sniffle/tool/fetch"
)

func ArtistExtra(t *tool.Tool, id string) (list *common.List) {
	return fetchArtist(t, id, false)
}

func ArtistStrict(t *tool.Tool, id string) (list *common.List) {
	return fetchArtist(t, id, true)
}

func fetchArtist(t *tool.Tool, id string, strict bool) (list *common.List) {
	list = fetchInfo(t, id)
	if list == nil {
		return nil
	}

	dto := struct{ Pagination struct{ Items uint } }{}
	if tool.FetchJSON(t, nil, &dto, fetch.Fmt("https://api.discogs.com/artists/%s/releases?per_page=100&page=0", id)) {
		return nil
	}

	for page := uint(0); page < dto.Pagination.Items; page += 100 {
		artistPage(t, list, strict, page)
	}
	slices.SortFunc(list.Items, func(a, b *common.Item) int { return cmp.Compare(b.Title, a.Title) })

	return
}

func artistPage(t *tool.Tool, list *common.List, strict bool, page uint) {
	dto := struct {
		Releases []struct {
			Artist string
			ID     uint
			Type   string
			Role   string
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, fetch.Fmt("https://api.discogs.com/artists/%s/releases?per_page=100&page=%d", list.ID, page)) {
		return
	}

	for _, r := range dto.Releases {
		if strict && r.Role != "Main" {
			continue
		}
		it := (*common.Item)(nil)
		switch r.Type {
		case "release":
			it = fetchItem(t, fetch.Fmt("https://api.discogs.com/releases/%d", r.ID))
		case "master":
			it = fetchItem(t, fetch.Fmt("https://api.discogs.com/masters/%d", r.ID))
		default:
			t.Warn("discogs.unknwon_realse_type", "type", r.Type)
		}
		if it != nil {
			list.Items = append(list.Items, it)
		}
	}
}

func fetchInfo(t *tool.Tool, id string) *common.List {
	dto := struct {
		Name string
		URI  string
	}{}
	if tool.FetchJSON(t, nil, &dto, fetch.Fmt("https://api.discogs.com/artists/%s", id)) {
		return nil
	}

	return &common.List{
		Host:  "discogs",
		ID:    id,
		Title: dto.Name,
		URL:   dto.URI,
		Items: make([]*common.Item, 0),
	}
}

func fetchItem(t *tool.Tool, r *fetch.Request) *common.Item {
	dto := struct {
		ID        int
		Title     string
		URI       string
		Year      uint
		Tracklist []struct {
			Title        string
			Duration     string
			Extraartists []struct {
				Name string
				Role string
			}
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, r) {
		return nil
	}

	desc := make([]string, len(dto.Tracklist))
	for i, track := range dto.Tracklist {
		feat := ""
		for _, extra := range track.Extraartists {
			feat = fmt.Sprintf("%s ||| %s [%s]", feat, extra.Name, extra.Role)
		}
		desc[i] = fmt.Sprintf("%02d. %s [%s]%s", i+1, track.Title, track.Duration, feat)
	}

	title := ""
	if dto.Year != 0 {
		title = fmt.Sprintf("(%d) %s", dto.Year, dto.Title)
	} else {
		title = " " + dto.Title
	}

	it := &common.Item{
		Host:        "discogs",
		ID:          strconv.Itoa(dto.ID),
		URL:         dto.URI,
		Title:       title,
		IsVideo:     true,
		Description: desc,
	}
	getImage(t, it)
	return it
}

var openGraphImage = regexp.MustCompile(`<meta data-rh="" property="og:image" content="([^"]+)"/`)

func getImage(t *tool.Tool, item *common.Item) {
	r := fetch.R("", item.URL, nil,
		"User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:141.0) Gecko/20100101 Firefox/141.0",
		"Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language", "fr-FR,en-US;q=0.8,en;q=0.5=",
		"Sec-GPC", "1",
		"Connection", "keep-alive",
		"Upgrade-Insecure-Requests", "1",
		"Sec-Fetch-Dest", "document",
		"Sec-Fetch-Mode", "navigate",
		"Sec-Fetch-Site", "same-origin",
		"Sec-Fetch-User", "?1",
		"Priority", "u=0, i",
		"TE", "trailers",
	)

	html := tool.FetchAll(t, r)
	target := openGraphImage.FindSubmatch(html)
	if target == nil {
		return
	}

	item.Poster, item.PosterWidth, item.PosterHeight = common.FetchPoster(t, string(target[1]))
}
