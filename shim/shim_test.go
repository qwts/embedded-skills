package shim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// testTarget is a synthetic target name for resolution tests; the shim is
// target-agnostic, the name is arbitrary.
const testTarget = "example-cli"

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

func TestVersionGE(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"2.103.1", "2.40", true},
		{"2.40", "2.40", true},
		{"2.39.9", "2.40", false},
		{"3.0", "2.40", true},
		{"0.12.0", "0.12.0", true},
		{"0.11.9", "0.12.0", false},
	}
	for _, tc := range cases {
		if got := versionGE(tc.a, tc.b); got != tc.want {
			t.Errorf("versionGE(%s, %s) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestResolveSkipsSelf(t *testing.T) {
	// PATH = [dir holding a copy of this test binary named example-cli, dir2
	// with a fake example-cli]. Resolution must skip the copy (SameFile) and
	// find dir2's. os.Executable is the test binary, not the copy, so
	// simulate by symlink: a symlink named example-cli -> os.Executable
	// resolves to the same file.
	self, err := os.Executable()
	if err != nil {
		t.Skip("no executable path")
	}
	d1 := t.TempDir()
	d2 := t.TempDir()
	if err := os.Symlink(self, filepath.Join(d1, testTarget)); err != nil {
		t.Skipf("symlink: %v", err)
	}
	fake := filepath.Join(d2, testTarget)
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", d1+string(os.PathListSeparator)+d2)
	t.Setenv("SKILLS_REAL_BIN", "")
	t.Setenv(realOverrideEnv(testTarget), "")
	got, err := resolveRealBinary(testTarget)
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
	wrapper := filepath.Join(d1, testTarget)
	real := filepath.Join(d2, testTarget)
	for _, p := range []string{wrapper, real} {
		if err := os.WriteFile(p, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", d1+string(os.PathListSeparator)+d2)
	t.Setenv("SKILLS_REAL_BIN", "")
	t.Setenv(realOverrideEnv(testTarget), "")
	t.Setenv(visitedEnv, resolvedPath(wrapper))
	got, err := resolveRealBinary(testTarget)
	if err != nil {
		t.Fatal(err)
	}
	if got != real {
		t.Fatalf("resolveRealBinary = %q, want %q (visited wrapper not skipped)", got, real)
	}
}

func TestResolveHonorsPerTargetOverride(t *testing.T) {
	// A hyphenated target's override normalizes to underscores:
	// SKILLS_REAL_AGENT_BOT for target "agent-bot".
	t.Setenv("SKILLS_REAL_BIN", "")
	t.Setenv("SKILLS_REAL_AGENT_BOT", "/managed/agent-bot")
	got, err := resolveRealBinary("agent-bot")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/managed/agent-bot" {
		t.Fatalf("resolveRealBinary = %q, want the per-target override", got)
	}
}

func TestChildEnvExtendsVisited(t *testing.T) {
	t.Setenv(visitedEnv, "/a/"+testTarget)
	env := childEnv("/b/" + testTarget)
	joined := strings.Join(env, "\n")
	want := visitedEnv + "=/a/" + testTarget + string(os.PathListSeparator) + "/b/" + testTarget
	if !strings.Contains(joined, want) {
		t.Fatalf("childEnv missing %q", want)
	}
}

// --- interception truth table (docs/01) over a synthetic corpus ---

func testCorpusFS() fstest.MapFS {
	skill := "---\nname: demo\ndescription: Does a thing. Use when things\n---\n\n# Demo\n"
	manifest := `{"version":1,"target":"` + testTarget + `","target_version":">=1.0",` +
		`"version_probe":{"args":["--version"],"regex":"(\\d+\\.\\d+(\\.\\d+)?)"},` +
		`"skills":[{"name":"demo","path":"skills/demo/SKILL.md","description":"Does a thing. Use when things",` +
		`"level":"beginner","tags":[],"commands":[],"prerequisites":[],"next":[],"checksum":"` + Checksum([]byte(skill)) + `"}]}`
	return fstest.MapFS{
		"manifest.json":        &fstest.MapFile{Data: []byte(manifest)},
		"skills/demo/SKILL.md": &fstest.MapFile{Data: []byte(skill)},
	}
}

func TestRunSkillsList(t *testing.T) {
	if code := Run(testTarget, testCorpusFS(), []string{testTarget, "skills"}); code != 0 {
		t.Fatalf("skills list exit = %d, want 0", code)
	}
}

func TestRunShowSkillVerbatim(t *testing.T) {
	// showSkill prints the document verbatim; capture stdout.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	code := Run(testTarget, testCorpusFS(), []string{testTarget, "skills", "demo"})
	os.Stdout = old
	w.Close()
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	out := string(buf[:n])
	if code != 0 {
		t.Fatalf("skills demo exit = %d, want 0", code)
	}
	if !strings.Contains(out, "name: demo") || !strings.Contains(out, "# Demo") {
		t.Fatalf("verbatim output missing frontmatter or body: %q", out)
	}
}

func TestRunUnknownSkill(t *testing.T) {
	if code := Run(testTarget, testCorpusFS(), []string{testTarget, "skills", "nope"}); code != 2 {
		t.Fatalf("unknown skill exit = %d, want 2", code)
	}
}

func TestRunEmptyCorpus(t *testing.T) {
	fsys := fstest.MapFS{
		"manifest.json": &fstest.MapFile{Data: []byte(`{"version":1,"target":"` + testTarget + `","target_version":">=1.0","version_probe":{"args":["--version"],"regex":"(\\d+)"},"skills":[]}`)},
	}
	if code := Run(testTarget, fsys, []string{testTarget, "skills"}); code != 3 {
		t.Fatalf("empty corpus exit = %d, want 3", code)
	}
}

func TestRunManagementModePassesSkillsThrough(t *testing.T) {
	// Invoked under any other name, the binary offers doctor and passes
	// `skills ...` through for development.
	if code := Run(testTarget, testCorpusFS(), []string{"shim-build-artifact", "skills"}); code != 0 {
		t.Fatalf("management skills exit = %d, want 0", code)
	}
	if code := Run(testTarget, testCorpusFS(), []string{"shim-build-artifact"}); code != 2 {
		t.Fatalf("management usage exit = %d, want 2", code)
	}
}
