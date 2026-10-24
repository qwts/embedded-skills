# embedded-skills

Agent skills embedded in a shim binary that stands in front of any CLI:
`<cli> skills` is to an agent what `<cli> --help` is to a human — discover
the workflows, load the one that fits, then run the real commands.

- **Design:** [`docs/`](docs/) — the `skills` contract, skill format,
  shim pattern (the only delivery mechanism: a binary named exactly
  `<target>` at the front of `PATH`), progressive learning, alignment,
  discovery, authoring, and evals. Start at [`docs/README.md`](docs/README.md).
- **Runnable example:** [`examples/gh/`](examples/gh/) — a Go shim named
  `gh` with a 4-skill corpus for GitHub's CLI. `make install` there puts
  it on your `PATH`; every non-`skills` invocation execs your real `gh`
  untouched.

```console
$ cd examples/gh && make install
$ export PATH="$HOME/.local/share/skills-shims/bin:$PATH"
$ gh skills            # from the shim
$ gh skills pr-flow    # one workflow, verbatim
$ gh pr status         # the real gh
```
