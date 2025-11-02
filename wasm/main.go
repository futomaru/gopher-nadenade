//go:build js && wasm

package main

import (
	_ "embed"
	"fmt"
	"math"
	"strconv"
	"syscall/js"
)

//go:embed source/gopher_neutral.svg
var neutralSVG []byte

//go:embed source/gopher_love.svg
var loveSVG []byte

type pupil struct {
	node      js.Value
	baseX     float64
	baseY     float64
	maxOffset float64
}

func main() {
	done := make(chan struct{})

	doc := js.Global().Get("document")
	root := doc.Call("getElementById", "app")
	console := js.Global().Get("console")

	parserConstructor := js.Global().Get("DOMParser")
	if !parserConstructor.Truthy() {
		console.Call("error", "DOMParser is not available in this environment.")
		return
	}

	parser := parserConstructor.New()

	neutralNode, err := importSVG(doc, parser, string(neutralSVG))
	if err != nil {
		console.Call("error", err.Error())
		return
	}

	loveNode, err := importSVG(doc, parser, string(loveSVG))
	if err != nil {
		console.Call("error", err.Error())
		return
	}

	setSVGAttributes(neutralNode)
	setSVGAttributes(loveNode)

	container := doc.Call("createElement", "div")
	style := container.Get("style")
	style.Call("setProperty", "cursor", "pointer", "")
	style.Call("setProperty", "max-width", "90vw", "")
	style.Call("setProperty", "width", "300px", "")
	style.Call("setProperty", "height", "auto", "")
	style.Call("setProperty", "touch-action", "none", "")

	container.Call("appendChild", neutralNode)
	container.Call("appendChild", loveNode)

	loveNode.Get("style").Call("setProperty", "display", "none", "")

	root.Call("appendChild", container)

	pupils := collectPupils(neutralNode)
	if len(pupils) == 0 {
		console.Call("warn", "No pupils were detected in the neutral SVG.")
	}

	var neutralVisible = true

	pointerMove := js.FuncOf(func(this js.Value, args []js.Value) any {
		if !neutralVisible || len(pupils) == 0 {
			return nil
		}
		event := args[0]
		updatePupils(neutralNode, pupils, event)
		return nil
	})
	js.Global().Call("addEventListener", "pointermove", pointerMove)
	defer js.Global().Call("removeEventListener", "pointermove", pointerMove)

	pointerEnter := js.FuncOf(func(this js.Value, args []js.Value) any {
		neutralVisible = false
		neutralNode.Get("style").Call("setProperty", "display", "none", "")
		loveNode.Get("style").Call("setProperty", "display", "block", "")
		resetPupils(pupils)
		return nil
	})

	pointerLeave := js.FuncOf(func(this js.Value, args []js.Value) any {
		neutralVisible = true
		loveNode.Get("style").Call("setProperty", "display", "none", "")
		neutralNode.Get("style").Call("setProperty", "display", "block", "")
		resetPupils(pupils)
		return nil
	})

	container.Call("addEventListener", "pointerenter", pointerEnter)
	container.Call("addEventListener", "pointerleave", pointerLeave)
	defer container.Call("removeEventListener", "pointerenter", pointerEnter)
	defer container.Call("removeEventListener", "pointerleave", pointerLeave)

	defer pointerMove.Release()
	defer pointerEnter.Release()
	defer pointerLeave.Release()

	<-done
}

func importSVG(doc, parser js.Value, source string) (js.Value, error) {
	parsed := parser.Call("parseFromString", source, "image/svg+xml")
	if !parsed.Truthy() {
		return js.Value{}, fmt.Errorf("failed to parse SVG markup")
	}

	errorEls := parsed.Call("getElementsByTagName", "parsererror")
	if errorEls.Truthy() && errorEls.Get("length").Int() > 0 {
		return js.Value{}, fmt.Errorf("parsererror reported while parsing SVG")
	}

	svg := parsed.Call("querySelector", "svg")
	if !svg.Truthy() {
		return js.Value{}, fmt.Errorf("parsed SVG document did not contain an <svg> root element")
	}

	return doc.Call("importNode", svg, true), nil
}

func setSVGAttributes(svg js.Value) {
	svg.Call("setAttribute", "role", "img")
	svg.Call("setAttribute", "width", "300")
	svg.Call("setAttribute", "height", "360")

	style := svg.Get("style")
	style.Call("setProperty", "width", "100%", "")
	style.Call("setProperty", "height", "auto", "")
	style.Call("setProperty", "display", "block", "")
	style.Call("setProperty", "max-width", "100%", "")
}

func collectPupils(svg js.Value) []pupil {
	nodes := svg.Call("querySelectorAll", "[id^='pupil']")
	length := 0
	if nodes.Truthy() {
		length = nodes.Get("length").Int()
	}

	results := make([]pupil, 0, length)
	for i := 0; i < length; i++ {
		node := nodes.Call("item", i)
		if !node.Truthy() {
			continue
		}

		cxAttr := node.Call("getAttribute", "cx")
		cyAttr := node.Call("getAttribute", "cy")
		rAttr := node.Call("getAttribute", "r")

		if !cxAttr.Truthy() || !cyAttr.Truthy() || !rAttr.Truthy() {
			continue
		}

		cx, err := strconv.ParseFloat(cxAttr.String(), 64)
		if err != nil {
			continue
		}

		cy, err := strconv.ParseFloat(cyAttr.String(), 64)
		if err != nil {
			continue
		}

		radius, err := strconv.ParseFloat(rAttr.String(), 64)
		if err != nil {
			continue
		}

		results = append(results, pupil{
			node:      node,
			baseX:     cx,
			baseY:     cy,
			maxOffset: radius * 0.7,
		})
	}

	return results
}

func updatePupils(svg js.Value, pupils []pupil, event js.Value) {
	point := svg.Call("createSVGPoint")
	if !point.Truthy() {
		return
	}

	point.Set("x", event.Get("clientX").Float())
	point.Set("y", event.Get("clientY").Float())

	for i := range pupils {
		p := &pupils[i]

		ctm := p.node.Call("getScreenCTM")
		if !ctm.Truthy() {
			continue
		}

		inverse := ctm.Call("inverse")
		if !inverse.Truthy() {
			continue
		}

		local := point.Call("matrixTransform", inverse)
		if !local.Truthy() {
			continue
		}

		targetX := local.Get("x").Float()
		targetY := local.Get("y").Float()

		dx := targetX - p.baseX
		dy := targetY - p.baseY
		dist := math.Hypot(dx, dy)

		if dist > 0 && dist > p.maxOffset {
			scale := p.maxOffset / dist
			dx *= scale
			dy *= scale
		}

		p.node.Call("setAttribute", "cx", formatFloat(p.baseX+dx))
		p.node.Call("setAttribute", "cy", formatFloat(p.baseY+dy))
	}
}

func resetPupils(pupils []pupil) {
	for _, p := range pupils {
		p.node.Call("setAttribute", "cx", formatFloat(p.baseX))
		p.node.Call("setAttribute", "cy", formatFloat(p.baseY))
	}
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}
