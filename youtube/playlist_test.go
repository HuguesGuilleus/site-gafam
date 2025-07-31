package youtube

import (
	_ "embed"
	"testing"
	"time"

	"github.com/HuguesGuilleus/sniffle/tool"
	"github.com/HuguesGuilleus/sniffle/tool/fetch"
	"github.com/stretchr/testify/assert"
)

//go:embed playlist.html
var playlistHTML []byte

func TestPlaylist(t *testing.T) {
	_, to := tool.NewTestTool(fetch.Test(
		fetch.URL("https://www.youtube.com/playlist?list=gta").T(200, playlistHTML),
	))
	p := FetchPlaylist(to, "gta")
	assert.NotNil(t, p)
	p.Items = p.Items[:1]
	assert.Equal(t, &Playlist{
		Id:    "gta",
		Title: "GTA 5 - Boblennon",
		Items: []VideoItem{
			{
				Id:    "XGPTbxJXGI0",
				Title: "[TWITCH] Boblennon - Grand Theft Auto  V - 07/01/17",
				Author: ChannelItem{
					Name: "Les lives de Boblennon",
					Tag:  "LeslivesdeBobLennon",
					Id:   "UCEHABGxoIxfCeicnDUp1g1w",
				},
				Duration: 17154 * time.Second},
		},
	}, p)
}
