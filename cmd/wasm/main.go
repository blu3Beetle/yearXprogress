//go:build js && wasm

// Command wasm exposes internal/progress to JavaScript for the GitHub Pages
// site, so the page computes the same year-progress values as the bot -
// live, in the visitor's browser, using their own clock.
package main

import (
	"syscall/js"
	"time"

	"github.com/blu3Beetle/yearXprogress/internal/progress"
)

func computeProgress(this js.Value, args []js.Value) any {
	snap := progress.Compute(time.Now())

	result := js.Global().Get("Object").New()
	result.Set("year", snap.Year)
	result.Set("percentage", snap.Percentage)
	result.Set("bar", snap.Bar)
	result.Set("text", snap.Text)
	return result
}

func main() {
	js.Global().Set("computeProgress", js.FuncOf(computeProgress))
	select {} // keep the Go runtime alive so JS can keep calling back into it
}
