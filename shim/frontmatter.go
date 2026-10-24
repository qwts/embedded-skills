package shim

import (
	"fmt"
	"strings"
)

// Skill is one parsed SKILL.md. Field set follows docs/02-skill-format.md:
// Agent Skills spec fields (name, description) plus this project's
// declared extensions.
type Skill struct {
	Name          string
	Description   string
	Level         string
	Risk          string
	Tags          []string
	Triggers      []string
	Commands      []string
	Requires      []string
	MinVersion    string
	Prerequisites []string
	Next          []string
	CompletesWith string

	Path       string // path inside the corpus FS, e.g. skills/pr-flow/SKILL.md
	Raw        string // full document, frontmatter included, served verbatim
	Body       string // markdown after the frontmatter
	References []string
}

// parseSkill parses the restricted YAML subset this corpus uses: scalar
// `key: value`, flow lists `key: [a, b]`, and block lists. It is not a
// general YAML parser; `skills check` rejects anything outside the subset
// by failing to parse. Keeping the shim dependency-free (standard library
// only) is worth more here than full YAML — see docs/09.
func parseSkill(path, raw string) (*Skill, error) {
	lines := strings.Split(raw, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, fmt.Errorf("%s: missing frontmatter opening ---", path)
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return nil, fmt.Errorf("%s: missing frontmatter closing ---", path)
	}
	s := &Skill{Path: path, Raw: raw, Body: strings.Join(lines[end+1:], "\n"), Level: "beginner", Risk: "mutating"}
	var listKey string
	for _, line := range lines[1:end] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") {
			if listKey == "" {
				return nil, fmt.Errorf("%s: list item without a key: %q", path, line)
			}
			s.setListItem(listKey, unquote(strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))))
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("%s: unparseable frontmatter line: %q", path, line)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		listKey = ""
		switch {
		case value == "":
			listKey = key // block list follows
		case strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]"):
			inner := strings.TrimSpace(value[1 : len(value)-1])
			if inner != "" {
				for _, item := range strings.Split(inner, ",") {
					s.setListItem(key, unquote(strings.TrimSpace(item)))
				}
			}
		default:
			s.setScalar(key, unquote(value))
		}
	}
	if s.Name == "" {
		return nil, fmt.Errorf("%s: name is required", path)
	}
	if s.Description == "" {
		return nil, fmt.Errorf("%s: description is required", path)
	}
	return s, nil
}

func (s *Skill) setScalar(key, value string) {
	switch key {
	case "name":
		s.Name = value
	case "description":
		s.Description = value
	case "level":
		s.Level = value
	case "risk":
		s.Risk = value
	case "min_version":
		s.MinVersion = value
	case "completes_with":
		s.CompletesWith = value
	}
}

func (s *Skill) setListItem(key, value string) {
	switch key {
	case "tags":
		s.Tags = append(s.Tags, value)
	case "triggers":
		s.Triggers = append(s.Triggers, value)
	case "commands":
		s.Commands = append(s.Commands, value)
	case "requires":
		s.Requires = append(s.Requires, value)
	case "prerequisites":
		s.Prerequisites = append(s.Prerequisites, value)
	case "next":
		s.Next = append(s.Next, value)
	}
}

func unquote(v string) string {
	if len(v) >= 2 {
		if (v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'') {
			return v[1 : len(v)-1]
		}
	}
	return v
}
