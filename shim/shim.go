// Package shim implements the generic embedded-skills PATH-shadow shim.
// One codebase, many targets: the corpus describes the target's workflows;
// the manifest records which target and which target version it was built
// against. See docs/06-shim-pattern.md.
package shim

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Run is the shim entry point. The target binary calls it with the target
// name, the embedded corpus (an embed.FS containing skills/ and manifest.json),
// and os.Args. It returns the process exit code.
func Run(target string, corpus embed.FS, argv []string) int {
	invokedAs := filepath.Base(argv[0])
	args := argv[1:]

	// Management mode: invoked under any name other than the target's
	// (e.g. the build artifact before install, or `go run .`).
	if invokedAs != target {
		return management(target, corpus, args)
	}

	// Skills surface.
	if len(args) > 0 && args[0] == "skills" {
		return skillsCmd(target, corpus, args[1:])
	}
	// Optional --skill convenience, only as a leading run of flags.
	if len(args) > 0 && (args[0] == "--skill" || strings.HasPrefix(args[0], "--skill=")) {
		return skillLoad(target, corpus, args)
	}
	return proxy(target, args)
}

func management(target string, corpus embed.FS, args []string) int {
	if len(args) > 0 && args[0] == "skills" {
		return skillsCmd(target, corpus, args[1:]) // handy during development: `go run . skills`
	}
	if len(args) > 0 && args[0] == "doctor" {
		return doctor(target, corpus)
	}
	fmt.Fprintf(os.Stderr, "%s: this binary shadows %q when installed under that name.\n", filepath.Base(os.Args[0]), target)
	fmt.Fprintf(os.Stderr, "usage: doctor | skills ...   (see README.md for install)\n")
	return 2
}

func doctor(target string, corpus embed.FS) int {
	self, _ := os.Executable()
	real, err := resolveRealBinary(target)
	fmt.Printf("shim binary:   %s\n", self)
	if err != nil {
		fmt.Printf("real %s:       NOT FOUND (%v)\n", target, err)
		return 1
	}
	cat, err := loadCatalog(corpus)
	if err != nil {
		fmt.Printf("corpus:        BROKEN (%v)\n", err)
		return 1
	}
	fmt.Printf("real %s:       %s\n", target, real)
	fmt.Printf("detected:      %s (corpus written for %s)\n", probeVersion(real, cat.Manifest.VersionProbe), cat.Manifest.TargetVersion)
	fmt.Printf("skills:        %d embedded\n", len(cat.Skills))
	return 0
}

func proxy(target string, args []string) int {
	real, err := resolveRealBinary(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "skills-shim: %v\n", err)
		return 1
	}
	if err := execReal(real, append([]string{real}, args...)); err != nil {
		fmt.Fprintf(os.Stderr, "skills-shim: exec %s: %v\n", real, err)
		return 1
	}
	return 0 // unreachable on Unix
}

// skillLoad implements the optional --skill convenience (doc 01): print
// delimited skill blocks, then run the remaining command via the proxy.
func skillLoad(target string, corpus embed.FS, args []string) int {
	cat, err := loadCatalog(corpus)
	if err != nil {
		return corpusError(err)
	}
	var names, rest []string
	i := 0
	for i < len(args) {
		switch {
		case args[i] == "--skill" && i+1 < len(args):
			names = append(names, args[i+1])
			i += 2
		case strings.HasPrefix(args[i], "--skill="):
			names = append(names, strings.TrimPrefix(args[i], "--skill="))
			i++
		default:
			rest = args[i:]
			i = len(args)
		}
	}
	if len(rest) > 0 && rest[0] == "--" {
		rest = rest[1:] // everything after -- belongs to the command
	}
	for _, name := range names {
		s := cat.find(name)
		if s == nil {
			return unknownSkill(cat, name)
		}
		fmt.Printf("<!-- skill: %s -->\n%s<!-- /skill: %s -->\n", s.Name, ensureNewline(s.Raw), s.Name)
	}
	return proxy(target, rest)
}

func ensureNewline(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}

func corpusError(err error) int {
	fmt.Fprintf(os.Stderr, "skills-shim: embedded corpus is broken: %v\n", err)
	return 1
}

func unknownSkill(cat *Catalog, name string) int {
	fmt.Fprintf(os.Stderr, "unknown skill %q\n", name)
	if sug := cat.suggestions(name); len(sug) > 0 {
		fmt.Fprintf(os.Stderr, "did you mean: %s?\n", strings.Join(sug, ", "))
	}
	return 2
}
