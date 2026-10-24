package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSkill(t *testing.T) {
	raw := "---\nname: demo\ndescription: Does a thing. Use when things\nlevel: advanced\ntags: [a, b]\nrequires:\n  - \"check one\"\n---\n\n# Demo\n"
	s, err := parseSkill("skills/demo/SKILL.md", raw)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "demo" || s.Level != "advanced" || len(s.Tags) != 2 || len(s.Requires) != 1 {
		t.Fatalf("bad parse: %+v", s)
	}
	if !strings.Contains(s.Body, "# Demo") {
		t.Fatalf("body missing: %q", s.Body)
	}
}

func TestCorpusIsValid(t *testing.T) {
	cat, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Skills) == 0 {
		t.Fatal("empty corpus")
	}
	for _, s := range cat.Skills {
		if len(s.Description) > 200 {
			t.Errorf("%s: description %d chars > 200", s.Name, len(s.Description))
		}
		for _, link := range append(s.Prerequisites, s.Next...) {
			if cat.find(link) == nil {
				t.Errorf("%s: dangling chain link %q", s.Name, link)
			}
		}
	}
	// Manifest checksums must match embedded bytes (staleness gate).
	for _, ms := range cat.Manifest.Skills {
		data, err := corpus.ReadFile(ms.Path)
		if err != nil {
			t.Fatal(err)
		}
		if got := checksum(data); got != ms.Checksum {
			t.Errorf("%s: checksum %s != manifest %s (run make manifest)", ms.Path, got, ms.Checksum)
		}
	}
}

func TestVersionGE(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"2.103.1", "2.40", true},
		{"2.40", "2.40", true},
		{"2.39.9", "2.40", false},
		{"3.0", "2.40", true},
	}
	for _, tc := range cases {
		if got := versionGE(tc.a, tc.b); got != tc.want {
			t.Errorf("versionGE(%s, %s) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestResolveSkipsSelf(t *testing.T) {
	// PATH = [dir holding a copy of this test binary named gh, dir2 with
	// a fake gh]. Resolution must skip the copy (SameFile) and find dir2's.
	// os.Executable is the test binary, not the copy, so simulate by
	// symlink: a symlink named gh -> os.Executable resolves to the same file.
	self, err := os.Executable()
	if err != nil {
		t.Skip("no executable path")
	}
	d1 := t.TempDir()
	d2 := t.TempDir()
	if err := os.Symlink(self, filepath.Join(d1, "git")); err != nil {
		t.Skipf("symlink: %v", err)
	}
	fake := filepath.Join(d2, "git")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", d1+string(os.PathListSeparator)+d2)
	t.Setenv("SKILLS_REAL_BIN", "")
	t.Setenv("SKILLS_REAL_GIT", "")
	got, err := resolveRealBinary("git")
	if err != nil {
		t.Fatal(err)
	}
	if got != fake {
		t.Fatalf("resolveRealBinary = %q, want %q (skip-self failed)", got, fake)
	}
}

func TestResolveSkipsVisitedWrapper(t *testing.T) {
	// Simulates stacked shims: another wrapper at d1 already handed off
	// to us (its resolved path is in the visited set), so resolution must
	// skip it too and reach the binary behind it in d2.
	d1, d2 := t.TempDir(), t.TempDir()
	wrapper := filepath.Join(d1, "git")
	real := filepath.Join(d2, "git")
	for _, p := range []string{wrapper, real} {
		if err := os.WriteFile(p, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", d1+string(os.PathListSeparator)+d2)
	t.Setenv("SKILLS_REAL_BIN", "")
	t.Setenv("SKILLS_REAL_GIT", "")
	t.Setenv(visitedEnv, resolvedPath(wrapper))
	got, err := resolveRealBinary("git")
	if err != nil {
		t.Fatal(err)
	}
	if got != real {
		t.Fatalf("resolveRealBinary = %q, want %q (visited wrapper not skipped)", got, real)
	}
}

func TestChildEnvExtendsVisited(t *testing.T) {
	t.Setenv(visitedEnv, "/a/gh")
	env := childEnv("/b/gh")
	joined := strings.Join(env, "\n")
	want := visitedEnv + "=/a/gh" + string(os.PathListSeparator) + "/b/gh"
	if !strings.Contains(joined, want) {
		t.Fatalf("childEnv missing %q", want)
	}
}
