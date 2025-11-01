//go:build js && wasm

package main

import (
	_ "embed"
	"encoding/base64"
	"syscall/js"
)

//go:embed source/gopher_neutral.svg
var neutralSVG []byte

//go:embed source/gopher_smile.svg
var smileSVG []byte

func main() {
	done := make(chan struct{})

	doc := js.Global().Get("document")
	root := doc.Call("getElementById", "app")

	img := doc.Call("createElement", "img")
	img.Call("setAttribute", "alt", "gopher")
	img.Call("setAttribute", "width", "300")
	img.Call("setAttribute", "height", "360")
	img.Call("setAttribute", "style", "cursor:pointer;max-width:90vw;height:auto;")

	neutral := dataURI(neutralSVG)
	smiling := dataURI(smileSVG)

	img.Set("src", neutral)

	onOver := js.FuncOf(func(this js.Value, args []js.Value) any {
		img.Set("src", smiling)
		return nil
	})
	defer onOver.Release()

	onOut := js.FuncOf(func(this js.Value, args []js.Value) any {
		img.Set("src", neutral)
		return nil
	})
	defer onOut.Release()

	img.Call("addEventListener", "pointerover", onOver)
	img.Call("addEventListener", "pointerout", onOut)

	root.Call("appendChild", img)

	<-done
}

func dataURI(svg []byte) string {
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(svg)
}
