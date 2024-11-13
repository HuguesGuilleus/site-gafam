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

func FetchPoster(t *tool.Tool, url string) (data []byte, width, height string) {
	data = tool.FetchAll(t, fetch.R("", url, nil))

	config, _, _ := image.DecodeConfig(bytes.NewReader(data))
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
