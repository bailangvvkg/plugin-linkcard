//go:build wasip1

// Command plugin-linkcard is the linkcard plugin's module: it turns the card
// code blocks of a rendered page into link cards, so a page carries the cards
// themselves and no script has to draw them.
package main

import "github.com/extism/go-pdk"

func main() {}

//go:wasmexport transform_html
func transformHTML() int32 {
	var in struct {
		Settings struct {
			DefaultSize string `json:"default_size"`
			Target      string `json:"target"`
			Links       bool   `json:"links"`
		} `json:"settings"`
		HTML string `json:"html"`
	}
	if err := pdk.InputJSON(&in); err != nil {
		pdk.SetError(err)
		return 1
	}
	size, target := "regular", "_blank"
	if knownSize(in.Settings.DefaultSize) {
		size = in.Settings.DefaultSize
	}
	if in.Settings.Target == "_self" {
		target = "_self"
	}

	out := cards.ReplaceAllStringFunc(in.HTML, func(tag string) string {
		card, ok := draw(tag, size, target)
		if !ok {
			return tag // not a card this can draw: leave the block as it is
		}
		return card
	})
	out = hyperlinks.ReplaceAllStringFunc(out, func(tag string) string {
		card, ok := draw(tag, size, target)
		if !ok {
			return tag
		}
		return card
	})
	if in.Settings.Links {
		out = bareCards(out, size, target)
	}
	if out == in.HTML {
		return 0
	}
	if err := pdk.OutputJSON(map[string]string{"html": out}); err != nil {
		pdk.SetError(err)
		return 1
	}
	return 0
}
