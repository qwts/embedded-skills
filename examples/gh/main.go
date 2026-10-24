// gh skills shim (example) — a PATH-shadow binary named `gh` that serves
// an embedded skill corpus and proxies everything else to the real gh.
// Design: ../../docs/06-shim-pattern.md, surface: ../../docs/01-cli-surface.md.
//
// The shim logic lives in the shared ../../shim package; this file is a
// thin per-target wrapper that embeds the corpus and calls shim.Run.
package main

import (
	"embed"
	"os"

	"embedded-skills/shim"
)

//go:embed skills manifest.json
var corpus embed.FS

func main() { os.Exit(shim.Run("gh", corpus, os.Args)) }
