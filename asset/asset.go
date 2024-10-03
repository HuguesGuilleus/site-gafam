package asset

import (
	"bytes"
	_ "embed"
	"sniffle/tool/render"

	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)

//go:embed body.css
var _cssBytes []byte

//go:embed app.mjs
var _jsBytes []byte

var cssBytes = func() []byte {
	out := bytes.Buffer{}
	if err := css.Minify(nil, &out, bytes.NewReader(_cssBytes), nil); err != nil {
		panic(err.Error())
	}
	return out.Bytes()
}()

var jsBytes = func() []byte {
	out := bytes.Buffer{}
	if err := js.Minify(nil, &out, bytes.NewReader(_jsBytes), nil); err != nil {
		panic(err.Error())
	}
	return out.Bytes()
}()

var Begin render.H = `<meta charset=utf-8>` +
	`<meta name=viewport content="width=device-width,initial-scale=1.0">` +
	`<style>` + render.H(cssBytes) + `</style>` +
	`<script type=module>` + render.H(jsBytes) + `</script>`

var Back = render.Back
