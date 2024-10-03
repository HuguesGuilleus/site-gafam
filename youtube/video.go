package youtube

import (
	"context"
	"sniffle/tool"
)

// Fetch and save the thunbail of the video
// Save in `/youtube/icon/{id}.jpg`
func FetchVideoImage(t *tool.Tool, id string) {
	t.WriteFile("/youtube/icon/"+id+".jpg", tool.FetchAll(context.Background(), t, "", "https://img.youtube.com/vi/"+id+"/hqdefault.jpg", nil, nil))
}
