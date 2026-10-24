# Build and install the gh skills shim.
#
#   make build     compile ./dist/gh (a binary named gh — that name is the mechanism)
#   make install   build + copy to $(SHIM_DIR)/gh, print the PATH line to add
#   make test      unit tests
#   make check     build + validate the embedded corpus against your real gh
#   make manifest  regenerate manifest.json after editing skills/
#   make uninstall remove $(SHIM_DIR)/gh

GO      ?= go
SHIM_DIR ?= $(HOME)/.local/share/skills-shims/bin

.PHONY: build install uninstall test check manifest clean

build:
	$(GO) build -trimpath -o dist/gh .

# Staleness is caught by `make test` (checksum gate) and `make check`,
# which refuse to run against a manifest that disagrees with skills/.
manifest:
	$(GO) run ./cmd/genmanifest .

test:
	$(GO) test ./...

check: build
	./dist/gh skills check

install: build
	mkdir -p "$(SHIM_DIR)"
	cp dist/gh "$(SHIM_DIR)/gh"
	@echo ""
	@echo "Installed: $(SHIM_DIR)/gh"
	@echo "Add the shim directory to the FRONT of PATH, e.g. in your shell profile:"
	@echo '  export PATH="$(SHIM_DIR):$$PATH"'
	@echo "Then verify:  which gh && gh skills"

uninstall:
	rm -f "$(SHIM_DIR)/gh"
	@echo "Removed $(SHIM_DIR)/gh — the real gh is now first on PATH again."

clean:
	rm -rf dist
