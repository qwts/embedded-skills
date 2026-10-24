package shim

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const exportMarker = ".embedded-skills-export"

// exportCmd: <target> skills export <name> [--to <dir>] [--strict]
// The only skills command that writes (docs/01).
func exportCmd(cat *Catalog, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: skills export <name> [--to <dir>] [--strict]")
		return 2
	}
	s := cat.find(args[0])
	if s == nil {
		return unknownSkill(cat, args[0])
	}
	to := filepath.Join(".embedded-skills", s.Name)
	strict := false
	for i := 1; i < len(args); i++ {
		switch {
		case args[i] == "--strict":
			strict = true
		case args[i] == "--to" && i+1 < len(args):
			to = args[i+1]
			i++
		default:
			fmt.Fprintf(os.Stderr, "unknown export flag %q\n", args[i])
			return 2
		}
	}
	if _, err := os.Stat(to); err == nil {
		if _, err := os.Stat(filepath.Join(to, exportMarker)); err != nil {
			fmt.Fprintf(os.Stderr, "refusing to overwrite %s: not an embedded-skills export\n", to)
			return 1
		}
	}
	doc := s.Raw
	if strict {
		// Strict harness validators: spec fields only (docs/02 Conformance).
		doc = fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n%s", s.Name, s.Description, s.Body)
	}
	if err := os.MkdirAll(to, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := os.WriteFile(filepath.Join(to, "SKILL.md"), []byte(doc), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, ref := range s.References {
		data, err := cat.corpus.ReadFile(filepath.ToSlash(filepath.Join(filepath.Dir(s.Path), ref)))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		dest := filepath.Join(to, ref)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	_ = os.WriteFile(filepath.Join(to, exportMarker), []byte("exported by embedded-skills shim\n"), 0o644)
	fmt.Println(to)
	return 0
}

var refPattern = regexp.MustCompile("references/[A-Za-z0-9._/-]+")
var namePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// checkCmd validates the embedded corpus (docs/02, rules as applicable to
// this example). Command-path checks run `<real> <path...> --help` when a
// real binary is resolvable; flag-level validation against help output is
// the production design's job (docs/06 help adapters) and out of scope here.
func checkCmd(target string, cat *Catalog) int {
	var problems []string
	real, realErr := resolveRealBinary(target)
	for _, ms := range cat.Manifest.Skills {
		data, err := cat.corpus.ReadFile(ms.Path)
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		if ms.Checksum != "" && checksum(data) != ms.Checksum {
			problems = append(problems, fmt.Sprintf("%s: stale manifest checksum (run: make manifest)", ms.Path))
		}
		s := cat.find(ms.Name)
		if s == nil {
			problems = append(problems, fmt.Sprintf("%s: manifest name %q not parsed", ms.Path, ms.Name))
			continue
		}
		if !namePattern.MatchString(s.Name) {
			problems = append(problems, fmt.Sprintf("%s: name %q fails grammar", ms.Path, s.Name))
		}
		if filepath.Base(filepath.Dir("/"+ms.Path)) != s.Name {
			problems = append(problems, fmt.Sprintf("%s: name %q != directory", ms.Path, s.Name))
		}
		if len(s.Description) > 200 {
			problems = append(problems, fmt.Sprintf("%s: description %d chars > 200", s.Name, len(s.Description)))
		}
		if n := strings.Count(s.Raw, "\n") + 1; n > 200 || len(s.Raw) > 20*1024 {
			problems = append(problems, fmt.Sprintf("%s: SKILL.md %d lines / %d bytes over budget", s.Name, n, len(s.Raw)))
		}
		switch s.Level {
		case "beginner", "intermediate", "advanced":
		default:
			problems = append(problems, fmt.Sprintf("%s: bad level %q", s.Name, s.Level))
		}
		switch s.Risk {
		case "read-only", "mutating", "destructive":
		default:
			problems = append(problems, fmt.Sprintf("%s: bad risk %q", s.Name, s.Risk))
		}
		for _, link := range append(s.Prerequisites, s.Next...) {
			if cat.find(link) == nil {
				problems = append(problems, fmt.Sprintf("%s: prerequisite/next %q does not exist", s.Name, link))
			}
		}
		for _, ref := range refPattern.FindAllString(s.Body, -1) {
			found := false
			for _, have := range s.References {
				if have == ref {
					found = true
				}
			}
			if !found {
				problems = append(problems, fmt.Sprintf("%s: body links missing %s", s.Name, ref))
			}
		}
		if realErr == nil {
			for _, cmdPath := range s.Commands {
				if !commandExists(real, cmdPath) {
					problems = append(problems, fmt.Sprintf("%s: command %q not in real %s", s.Name, cmdPath, target))
				}
			}
		}
	}
	// Catalog budget: rendered layer-1 list ~ name+description per skill.
	total := 0
	for _, s := range cat.Skills {
		total += len(s.Name) + len(s.Description) + len(s.Level) + 4
	}
	if total > 8*1024 {
		problems = append(problems, fmt.Sprintf("catalog %d bytes > 8 KB budget", total))
	}
	if realErr != nil {
		fmt.Fprintf(os.Stderr, "note: real %s not resolvable; command-path checks skipped\n", target)
	}
	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, "check:", p)
		}
		return 1
	}
	fmt.Printf("ok: %d skills, catalog %d bytes, target %s %s\n", len(cat.Skills), total, target, cat.Manifest.TargetVersion)
	return 0
}

func commandExists(real, cmdPath string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	args := append(strings.Fields(cmdPath), "--help")
	cmd := exec.CommandContext(ctx, real, args...)
	cmd.Env = childEnv(real)
	return cmd.Run() == nil
}

// --- skills subcommand ---

func skillsCmd(target string, corpus embed.FS, args []string) int {
	cat, err := loadCatalog(corpus)
	if err != nil {
		return corpusError(err)
	}
	if len(cat.Skills) == 0 {
		fmt.Fprintln(os.Stderr, "no skills embedded for this target")
		return 3
	}
	if len(args) == 0 {
		return printList(target, cat, cat.Skills, false)
	}
	switch args[0] {
	case "--json":
		return printList(target, cat, filterArgs(cat, args[1:]), true)
	case "--level", "--tag":
		return printList(target, cat, filterArgs(cat, args), hasJSON(args))
	case "search":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: skills search <query>")
			return 2
		}
		return printList(target, cat, cat.search(strings.Join(args[1:], " ")), false)
	case "export":
		return exportCmd(cat, args[1:])
	case "check":
		return checkCmd(target, cat)
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
func (c *Catalog) detectedVersion(target string) string {
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

func printList(target string, cat *Catalog, skills []*Skill, asJSON bool) int {
	skills = cat.sorted(skills)
	detected := cat.detectedVersion(target)
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
