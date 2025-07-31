package twitch

import (
	"net/http"
	"strconv"

	"github.com/HuguesGuilleus/site-gafam/service/common"
	"github.com/HuguesGuilleus/sniffle/tool"
	"github.com/HuguesGuilleus/sniffle/tool/fetch"
)

const (
	teamBodyBegin = `[{` +
		`"operationName":"TeamLandingMemberList",` +
		`"variables":{` +
		`"teamName":`
	teamBodyEnd = `,` +
		`"withLiveMembers":false,` +
		`"withMembers":true` +
		`},` +
		`"extensions":{` +
		`"persistedQuery":{` +
		`"version":1,` +
		`"sha256Hash":"ee7d5bb7aeb195ac05164b6f306f1eb51db407c59f4398cbaa7901a3c3ba833d"` +
		`}` +
		`}` +
		`}]`
)

func Team(t *tool.Tool, id string) (list []*common.List) {
	body := teamBodyBegin + strconv.Quote(id) + teamBodyEnd
	dto := [1]struct {
		Data struct {
			Team struct {
				Members struct {
					Edges []struct {
						Node struct {
							Login string
						}
					}
				}
			}
		}
	}{}
	if tool.FetchJSON(t, nil, &dto, fetch.Rs(http.MethodPost, endpointURL, body, "Client-ID", clientID)) {
		return nil
	}

	for _, edge := range dto[0].Data.Team.Members.Edges {
		list = append(list, Channel(t, edge.Node.Login))
	}

	return
}
