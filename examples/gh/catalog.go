package main

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed skills manifest.json
var corpus embed.FS

type VersionProbe struct {
	Args  []string `json:"args"`
	Regex string   `json:"regex"`
}

type ManifestSkill struct {
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
	Version       int             `json:"version"`
	Target        string          `json:"target"`
	TargetVersion string          `json:"target_version"`
	VersionProbe  VersionProbe    `json:"version_probe"`
	Skills        []ManifestSkill `json:"skills"`
}

// Catalog is the loaded corpus: manifest + parsed skills + references.
type Catalog struct {
	Manifest Manifest
	Skills   []*Skill
	byName   map[string]*Skill
}

func loadCatalog() (*Catalog, error) {
	raw, err := corpus.ReadFile("manifest.json")
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("manifest.json: %w", err)
	}
	c := &Catalog{Manifest: m, byName: map[string]*Skill{}}
	for _, ms := range m.Skills {
		data, err := corpus.ReadFile(ms.Path)
		if err != nil {
			return nil, fmt.Errorf("manifest path %s: %w", ms.Path, err)
		}
		s, err := parseSkill(ms.Path, string(data))
		if err != nil {
			return nil, err
		}
		refs, err := skillReferences(path.Dir(ms.Path))
		if err != nil {
			return nil, err
		}
		s.References = refs
		if _, dup := c.byName[s.Name]; dup {
			return nil, fmt.Errorf("duplicate skill name %q", s.Name)
		}
		c.byName[s.Name] = s
		c.Skills = append(c.Skills, s)
	}
	return c, nil
}

func skillReferences(skillDir string) ([]string, error) {
	refsDir := skillDir + "/references"
	entries, err := fs.ReadDir(corpus, refsDir)
	if err != nil {
		return nil, nil // no references/ is normal
	}
	var refs []string
	for _, e := range entries {
		if !e.IsDir() {
			refs = append(refs, "references/"+e.Name())
		}
	}
	sort.Strings(refs)
	return refs, nil
}

func (c *Catalog) find(name string) *Skill { return c.byName[name] }

func levelRank(level string) int {
	switch level {
	case "beginner":
		return 0
	case "intermediate":
		return 1
	case "advanced":
		return 2
	}
	return 3
}

// sorted returns skills in catalog order: level, then name.
func (c *Catalog) sorted(skills []*Skill) []*Skill {
	out := append([]*Skill(nil), skills...)
	sort.SliceStable(out, func(i, j int) bool {
		if levelRank(out[i].Level) != levelRank(out[j].Level) {
			return levelRank(out[i].Level) < levelRank(out[j].Level)
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func checksum(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func (c *Catalog) suggestions(name string) []string {
	var out []string
	for _, s := range c.Skills {
		if strings.Contains(s.Name, name) || strings.HasPrefix(name, strings.Split(s.Name, "-")[0]) {
			out = append(out, s.Name)
		}
	}
	sort.Strings(out)
	return out
}
