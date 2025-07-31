package youtube

import (
	"github.com/HuguesGuilleus/sniffle/tool"
	"github.com/HuguesGuilleus/sniffle/tool/fetch"
)

// Fetch and save the thunbail of the video
// Save in `/youtube/icon/{id}.jpg`
func FetchVideoImage(t *tool.Tool, id string) {
	t.WriteFile("/youtube/icon/"+id+".jpg", tool.FetchAll(t, fetch.R("", "https://img.youtube.com/vi/"+id+"/hqdefault.jpg", nil)))
}
