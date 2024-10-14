package front

import (
	"embed"
	"sniffle/tool/fronttool"
	"sniffle/tool/render"
)

//go:embed asset.css
var fsysCSS embed.FS

//go:embed asset.js
var assetJS []byte

var begin = func() render.H {
	return `<style>` +
		render.H(fronttool.CSS(fsysCSS, nil)) +
		`</style>` +
		fronttool.InlineJs(assetJS)
}()
