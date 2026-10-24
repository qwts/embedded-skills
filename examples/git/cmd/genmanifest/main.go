// genmanifest regenerates ../../manifest.json from ../../skills/.
// The manifest is generated, never hand-edited (docs/03). Run via
// `make manifest` after adding or editing a skill; CI fails if the
// checked-in copy is stale, and `gh skills check` verifies checksums
// at runtime.
//
// This tool parses only the manifest-relevant frontmatter fields, in the
// same restricted subset the shim parses (scalars + flow lists). It is a
// build-time tool; it is never installed or run by agents.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Skill struct {
	Name          string   `json:"name"`
	Path          string   `json:"path"`
	Description   string   `json:"description"`
	Level         string   `json:"level"`
	Tags          []string `json:"tags"`
	Commands      []string `json:"commands"`
	Prerequisites []string `json:"prerequisites"`
	Next          []string `json:"next"`
	Checksum      string   `json:"checksum"`
}

type Manifest struct {
	Version       int    `json:"version"`
	Target        string `json:"target"`
	TargetVersion string `json:"target_version"`
	VersionProbe  struct {
		Args  []string `json:"args"`
		Regex string   `json:"regex"`
	} `json:"version_probe"`
	Skills []Skill `json:"skills"`
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1] // example dir; default: run from examples/gh
	}
	skillsDir := filepath.Join(root, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		fatal(err)
	}
	m := Manifest{Version: 1, Target: "git", TargetVersion: ">=2.54"}
	m.VersionProbe.Args = []string{"--version"}
	m.VersionProbe.Regex = `(\d+\.\d+(\.\d+)?)`

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		rel := filepath.Join("skills", dir, "SKILL.md")
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			fatal(err)
		}
		fields := frontmatter(string(data))
		sum := sha256.Sum256(data)
		s := Skill{
			Name:          fields["name"],
			Path:          filepath.ToSlash(rel),
			Description:   fields["description"],
			Level:         orDefault(fields["level"], "beginner"),
			Tags:          list(fields["tags"]),
			Commands:      list(fields["commands"]),
			Prerequisites: list(fields["prerequisites"]),
			Next:          list(fields["next"]),
			Checksum:      "sha256:" + hex.EncodeToString(sum[:]),
		}
		if s.Name == "" || s.Description == "" {
			fatal(fmt.Errorf("%s: name and description are required", rel))
		}
		if s.Name != dir {
			fatal(fmt.Errorf("%s: name %q != directory %q", rel, s.Name, dir))
		}
		m.Skills = append(m.Skills, s)
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), append(out, '\n'), 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("manifest.json: %d skills for %s %s\n", len(m.Skills), m.Target, m.TargetVersion)
}

func frontmatter(raw string) map[string]string {
	out := map[string]string{}
	lines := strings.Split(raw, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return out
	}
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			break
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		v := strings.TrimSpace(value)
		if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
			v = v[1 : len(v)-1]
		}
		out[strings.TrimSpace(key)] = v
	}
	return out
}

func list(v string) []string {
	v = strings.TrimSpace(v)
	if !strings.HasPrefix(v, "[") {
		return []string{}
	}
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(v, "["), "]"))
	if inner == "" {
		return []string{}
	}
	var out []string
	for _, item := range strings.Split(inner, ",") {
		out = append(out, strings.Trim(strings.TrimSpace(item), `"'`))
	}
	return out
}

func orDefault(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "genmanifest:", err)
	os.Exit(1)
}
