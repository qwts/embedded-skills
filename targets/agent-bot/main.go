// agent-bot skills shim — a PATH-shadow binary named `agent-bot` that
// serves an embedded skill corpus and proxies everything else to the real
// agent-bot. Design: ../../docs/06-shim-pattern.md, surface:
// ../../docs/01-cli-surface.md. The shim logic is the shared module in
// ../../shim; this target contributes only its name and its corpus.
package main

import (
	"embed"
	"os"

	shim "embedded-skills/shim"
)

//go:embed skills manifest.json
var corpus embed.FS

func main() { os.Exit(shim.Run("agent-bot", corpus, os.Args)) }
