package shim

import (
	"context"
	"fmt"
	"io/fs"
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
func exportCmd(fsys fs.FS, cat *Catalog, args []string) int {
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
		data, err := fs.ReadFile(fsys, filepath.ToSlash(filepath.Join(filepath.Dir(s.Path), ref)))
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
// this implementation). Command-path checks run `<real> <path...> --help`
// when a real binary is resolvable; flag-level validation against help
// output is the production design's job (docs/06 help adapters) and out of
// scope here.
func checkCmd(fsys fs.FS, target string, cat *Catalog) int {
	var problems []string
	real, realErr := resolveRealBinary(target)
	for _, ms := range cat.Manifest.Skills {
		data, err := fs.ReadFile(fsys, ms.Path)
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		if ms.Checksum != "" && Checksum(data) != ms.Checksum {
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
	// Collision guard (docs/06): if the real target answers `skills --help`,
	// the shim would shadow a native skills surface. The runtime does not
	// probe on every call (the skills path is the agent's hot path); corpus
	// validation catches the collision instead.
	if realErr == nil && commandExists(real, "skills") {
		problems = append(problems, fmt.Sprintf("real %s answers `skills --help`; the shim would shadow a native skills surface (docs/06 collision policy)", target))
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
