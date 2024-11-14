package common

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"sniffle/tool"
	"sniffle/tool/fetch"
	"strconv"
	"strings"
)

func FetchPoster(t *tool.Tool, url string) (poster []byte, width, height string) {
	poster = tool.FetchAll(t, fetch.URL(url))

	config, _, _ := image.DecodeConfig(bytes.NewReader(poster))
	width = strconv.Itoa(config.Width)
	height = strconv.Itoa(config.Height)

	return
}

func Description(src string) (description []string) {
	description = strings.Split(src, "\n")
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
	return
}
