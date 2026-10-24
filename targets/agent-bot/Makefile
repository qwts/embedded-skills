# Build and install the agent-bot skills shim.
#
#   make build     compile ./dist/agent-bot (a binary named agent-bot — that name is the mechanism)
#   make install   build + copy to $(SHIM_DIR)/agent-bot, print the PATH line to add
#   make test      unit tests (target corpus + the shared shim module)
#   make check     build + validate the embedded corpus against your real agent-bot
#   make manifest  regenerate manifest.json after editing skills/
#   make uninstall remove $(SHIM_DIR)/agent-bot

GO      ?= go
SHIM_DIR ?= $(HOME)/.local/share/skills-shims/bin

.PHONY: build install uninstall test check manifest clean

build:
	$(GO) build -trimpath -o dist/agent-bot .

# Staleness is caught by `make test` (checksum gate) and `make check`,
# which refuse to run against a manifest that disagrees with skills/.
manifest:
	$(GO) run embedded-skills/shim/cmd/genmanifest --root . --target agent-bot --target-version ">=0.12.0 <0.13.0"

test:
	$(GO) test ./...
	cd ../../shim && $(GO) test ./...

check: build
	./dist/agent-bot skills check

install: build
	mkdir -p "$(SHIM_DIR)"
	cp dist/agent-bot "$(SHIM_DIR)/agent-bot"
	@echo ""
	@echo "Installed: $(SHIM_DIR)/agent-bot"
	@echo "Add the shim directory to the FRONT of PATH, e.g. in your shell profile:"
	@echo '  export PATH="$(SHIM_DIR):$$PATH"'
	@echo "Then verify:  which agent-bot && agent-bot skills"

uninstall:
	rm -f "$(SHIM_DIR)/agent-bot"
	@echo "Removed $(SHIM_DIR)/agent-bot — the real agent-bot is now first on PATH again."

clean:
	rm -rf dist
