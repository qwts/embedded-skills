// gh skills shim (example) — a PATH-shadow binary named `gh` that serves
// an embedded skill corpus and proxies everything else to the real gh.
// Design: ../../docs/06-shim-pattern.md, surface: ../../docs/01-cli-surface.md.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const target = "git"

func main() { os.Exit(run(os.Args)) }

func run(argv []string) int {
	invokedAs := filepath.Base(argv[0])
	args := argv[1:]

	// Management mode: invoked under any name other than the target's
	// (e.g. the build artifact before install, or `go run .`).
	if invokedAs != target {
		return management(args)
	}

	// Skills surface.
	if len(args) > 0 && args[0] == "skills" {
		return skillsCmd(args[1:])
	}
	// Optional --skill convenience, only as a leading run of flags.
	if len(args) > 0 && (args[0] == "--skill" || strings.HasPrefix(args[0], "--skill=")) {
		return skillLoad(args)
	}
	return proxy(args)
}

func management(args []string) int {
	if len(args) > 0 && args[0] == "skills" {
		return skillsCmd(args[1:]) // handy during development: `go run . skills`
	}
	if len(args) > 0 && args[0] == "doctor" {
		return doctor()
	}
	fmt.Fprintf(os.Stderr, "%s: this binary shadows %q when installed under that name.\n", filepath.Base(os.Args[0]), target)
	fmt.Fprintf(os.Stderr, "usage: doctor | skills ...   (see README.md for install)\n")
	return 2
}

func doctor() int {
	self, _ := os.Executable()
	real, err := resolveRealBinary(target)
	fmt.Printf("shim binary:   %s\n", self)
	if err != nil {
		fmt.Printf("real %s:       NOT FOUND (%v)\n", target, err)
		return 1
	}
	cat, err := loadCatalog()
	if err != nil {
		fmt.Printf("corpus:        BROKEN (%v)\n", err)
		return 1
	}
	fmt.Printf("real %s:       %s\n", target, real)
	fmt.Printf("detected:      %s (corpus written for %s)\n", probeVersion(real, cat.Manifest.VersionProbe), cat.Manifest.TargetVersion)
	fmt.Printf("skills:        %d embedded\n", len(cat.Skills))
	return 0
}

func proxy(args []string) int {
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
func skillLoad(args []string) int {
	cat, err := loadCatalog()
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
	return proxy(rest)
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

// --- skills subcommand ---

func skillsCmd(args []string) int {
	cat, err := loadCatalog()
	if err != nil {
		return corpusError(err)
	}
	if len(cat.Skills) == 0 {
		fmt.Fprintln(os.Stderr, "no skills embedded for this target")
		return 3
	}
	if len(args) == 0 {
		return printList(cat, cat.Skills, false)
	}
	switch args[0] {
	case "--json":
		return printList(cat, filterArgs(cat, args[1:]), true)
	case "--level", "--tag":
		return printList(cat, filterArgs(cat, args), hasJSON(args))
	case "search":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: skills search <query>")
			return 2
		}
		return printList(cat, cat.search(strings.Join(args[1:], " ")), false)
	case "export":
		return exportCmd(cat, args[1:])
	case "check":
		return checkCmd(cat)
	default:
		return showSkill(cat, args)
	}
}

func hasJSON(args []string) bool {
	for _, a := range args {
		if a == "--json" {
			return true
		}
	}
	return false
}

func filterArgs(cat *Catalog, args []string) []*Skill {
	skills := cat.Skills
	for i := 0; i+1 < len(args); i += 2 {
		skills = filter(skills, args[i], args[i+1])
	}
	return skills
}

func filter(skills []*Skill, flag, value string) []*Skill {
	var out []*Skill
	for _, s := range skills {
		switch flag {
		case "--level":
			if s.Level == value {
				out = append(out, s)
			}
		case "--tag":
			for _, t := range s.Tags {
				if t == value {
					out = append(out, s)
				}
			}
		default:
			return skills
		}
	}
	return out
}

func (c *Catalog) search(query string) []*Skill {
	q := strings.ToLower(query)
	var out []*Skill
	for _, s := range c.Skills {
		hay := strings.ToLower(strings.Join(append([]string{s.Name, s.Description}, append(s.Tags, s.Triggers...)...), " "))
		if strings.Contains(hay, q) {
			out = append(out, s)
		}
	}
	return out
}

// detectedVersion probes the real binary once, skills path only.
func (c *Catalog) detectedVersion() string {
	real, err := resolveRealBinary(target)
	if err != nil {
		return ""
	}
	return probeVersion(real, c.Manifest.VersionProbe)
}

func (c *Catalog) versionOK(s *Skill, detected string) bool {
	if detected == "" || s.MinVersion == "" {
		return true // unknown is not a violation; detected_version is null
	}
	return versionGE(detected, s.MinVersion)
}

func printList(cat *Catalog, skills []*Skill, asJSON bool) int {
	skills = cat.sorted(skills)
	detected := cat.detectedVersion()
	if asJSON {
		type entry struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Level       string   `json:"level"`
			Tags        []string `json:"tags"`
			Commands    []string `json:"commands"`
			VersionOK   bool     `json:"version_ok"`
		}
		out := struct {
			Version         int     `json:"version"`
			DetectedVersion *string `json:"detected_version"`
			Skills          []entry `json:"skills"`
		}{Version: 1}
		if detected != "" {
			out.DetectedVersion = &detected
		}
		for _, s := range skills {
			out.Skills = append(out.Skills, entry{s.Name, s.Description, s.Level, orEmpty(s.Tags), orEmpty(s.Commands), cat.versionOK(s, detected)})
		}
		if out.Skills == nil {
			out.Skills = []entry{}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			return 1
		}
		return 0
	}
	wName, wLevel := len("NAME"), len("LEVEL")
	for _, s := range skills {
		if len(s.Name) > wName {
			wName = len(s.Name)
		}
		if len(s.Level) > wLevel {
			wLevel = len(s.Level)
		}
	}
	fmt.Printf("%-*s  %-*s  DESCRIPTION\n", wName, "NAME", wLevel, "LEVEL")
	for _, s := range skills {
		fmt.Printf("%-*s  %-*s  %s\n", wName, s.Name, wLevel, s.Level, s.Description)
		if !cat.versionOK(s, detected) {
			fmt.Fprintf(os.Stderr, "warning: %s needs %s >= %s (detected %s)\n", s.Name, target, s.MinVersion, detected)
		}
	}
	fmt.Printf("\nLoad one with: %s skills <name>\n", target)
	return 0
}

func orEmpty(v []string) []string {
	if v == nil {
		return []string{}
	}
	return v
}

func showSkill(cat *Catalog, args []string) int {
	name := args[0]
	s := cat.find(name)
	if s == nil {
		return unknownSkill(cat, name)
	}
	if hasJSON(args[1:]) {
		out := map[string]any{
			"version": 1, "name": s.Name, "description": s.Description,
			"level": s.Level, "tags": orEmpty(s.Tags), "commands": orEmpty(s.Commands),
			"prerequisites": orEmpty(s.Prerequisites), "next": orEmpty(s.Next),
			"body": strings.TrimPrefix(s.Body, "\n"), "references": orEmpty(s.References),
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			return 1
		}
		return 0
	}
	fmt.Print(ensureNewline(s.Raw)) // verbatim: the document agents already parse
	return 0
}
