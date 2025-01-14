package common

import (
	"bytes"
	"crypto/sha256"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"slices"
	"sniffle/tool"
	"sniffle/tool/fetch"
	"strconv"
	"strings"
)

func FetchPoster(t *tool.Tool, url string) (poster []byte, width, height string) {
	if url == "" {
		return
	}

	poster = tool.FetchAll(t, fetch.URL(url))

	config, _, _ := image.DecodeConfig(bytes.NewReader(poster))
	width = strconv.Itoa(config.Width)
	height = strconv.Itoa(config.Height)

	return
}

func TextPoster(text string) (poster []byte, width, height string) {
	img := image.NewNRGBA(image.Rect(0, 0, 255, 255))

	h := sha256.Sum256([]byte(text))
	copy(img.Pix, h[:4])
	img.Pix[4] = 255
	for n := 4; n < len(img.Pix); {
		n += copy(img.Pix[n:], img.Pix[:n])
	}

	buff := bytes.Buffer{}
	png.Encode(&buff, img)

	return buff.Bytes(), "255", "255"
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

// Stable sort items by Published date.
func SortByDate(items []*Item) {
	slices.SortStableFunc(items, func(a, b *Item) int {
		return b.Published.Compare(a.Published)
	})
}
