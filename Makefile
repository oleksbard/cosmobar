BIN := ./bin/cosmobar

.PHONY: build test dev fmt

build:
	go build -o $(BIN) .

test:
	go test ./...

fmt:
	go fmt ./...

# Build then wire a project-scoped test settings file at ./testsettings/.claude/settings.json
dev: build
	mkdir -p testsettings/.claude
	printf '{\n  "statusLine": {\n    "type": "command",\n    "command": "%s"\n  }\n}\n' "$(abspath $(BIN))" > testsettings/.claude/settings.json
	@echo "Wired $(abspath $(BIN)) into testsettings/.claude/settings.json"
	@echo "Run Claude Code from ./testsettings to see it live."
