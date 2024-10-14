package common

import (
	"bytes"
	"context"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"sniffle/tool"
	"strconv"
	"strings"
)

func FetchPoster(t *tool.Tool, url string) (data []byte, width, height string) {
	data = tool.FetchAll(context.Background(), t, "", url, nil, nil)

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
