package shim

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// visitedEnv records resolved paths this shim chain has already handed
// off to. A machine may already have another wrapper for the target
// (e.g. an identity/token shim script): that wrapper resolves *its*
// "real" binary by scanning PATH, finds us first, and execs us back —
// two shims bouncing forever. Threading a visited set through the env
// breaks the cycle generically: every handoff records the resolved
// path handed to, and resolution skips anything already visited. The
// intermediary wrapper still runs once, so any env it sets (tokens,
// identity) is in place when the chain reaches the true binary.
const visitedEnv = "EMBEDDED_SKILLS_SHIM_VISITED"

func resolvedPath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}

func visitedSet() map[string]bool {
	out := map[string]bool{}
	for _, p := range filepath.SplitList(os.Getenv(visitedEnv)) {
		if p != "" {
			out[p] = true
		}
	}
	return out
}

// childEnv returns the environment for handing off to `real`, with the
// visited set extended by real's resolved path.
func childEnv(real string) []string {
	visited := filepath.SplitList(os.Getenv(visitedEnv))
	visited = append(visited, resolvedPath(real))
	env := os.Environ()
	kv := visitedEnv + "=" + strings.Join(visited, string(os.PathListSeparator))
	for i, e := range env {
		if strings.HasPrefix(e, visitedEnv+"=") {
			env[i] = kv
			return env
		}
	}
	return append(env, kv)
}

// resolveRealBinary finds the real target binary: SKILLS_REAL_<TARGET> /
// SKILLS_REAL_BIN override first, else walk PATH in order skipping any
// entry that is this shim (compared with os.SameFile after resolving
// symlinks, never string equality) and anything already visited in this
// handoff chain. See docs/06, "No recursion".
func resolveRealBinary(target string) (string, error) {
	if v := os.Getenv("SKILLS_REAL_" + strings.ToUpper(target)); v != "" {
		return v, nil
	}
	if v := os.Getenv("SKILLS_REAL_BIN"); v != "" {
		return v, nil
	}
	visited := visitedSet()
	self, selfErr := os.Executable()
	var selfInfo os.FileInfo
	if selfErr == nil {
		self = resolvedPath(self)
		selfInfo, _ = os.Stat(self)
		visited[self] = true
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			dir = "."
		}
		candidate := filepath.Join(dir, target)
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() || info.Mode()&0o111 == 0 {
			continue
		}
		resolved := resolvedPath(candidate)
		if visited[resolved] {
			continue
		}
		if selfInfo != nil {
			if ri, err := os.Stat(resolved); err == nil && os.SameFile(selfInfo, ri) {
				continue // that's us
			}
		}
		return candidate, nil
	}
	return "", fmt.Errorf("real %q not found on PATH (is the shim shadowing itself?)", target)
}

// probeVersion runs the manifest's version probe against the real binary.
// Skills path only — never on the proxy path. Failure is (nil, nil-ish):
// unknown version, never an error that blocks serving skills.
func probeVersion(real string, probe VersionProbe) string {
	if real == "" || probe.Regex == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, real, probe.Args...)
	cmd.Env = childEnv(real)
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		return ""
	}
	re, err := regexp.Compile(probe.Regex)
	if err != nil {
		return ""
	}
	m := re.FindSubmatch(out)
	if len(m) < 2 {
		return ""
	}
	return string(m[1])
}

// versionGE reports dotted-version a >= b ("2.103.1" >= "2.40").
func versionGE(a, b string) bool {
	pa, pb := parseVersion(a), parseVersion(b)
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if x != y {
			return x > y
		}
	}
	return true
}

func parseVersion(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), ">=")
	v = strings.TrimPrefix(v, "v")
	var out []int
	for _, p := range strings.Split(v, ".") {
		n, err := strconv.Atoi(p)
		if err != nil {
			return out
		}
		out = append(out, n)
	}
	return out
}
