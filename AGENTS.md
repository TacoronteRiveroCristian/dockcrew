# Repository Guidelines

## Project Structure & Module Organization
- `cmd/dockcrew/main.go` provides the CLI entrypoint; keep startup logic thin and delegate to internal packages.
- `internal/config`, `internal/docker`, `internal/ui`, `internal/logs`, `internal/update`, and `internal/telemetry` hold domain logic; group new code by capability, not by layer.
- `docs/` houses the PRD, SPEC, and other planning artefacts. Update these when behaviour changes.
- Build outputs live in `bin/`; nothing under `docs/` or `bin/` should be imported by production code.

## Build, Test, and Development Commands
- `make build` compiles a release-like binary with version metadata and writes it to `bin/dockcrew`.
- `make run` rebuilds then invokes the binary with `--version`, useful for smoke checks.
- `make test` runs `go test ./...`; use it before every push.
- `make fmt` and `make vet` enforce canonical Go formatting and catch static issues. Prefer running both before opening a PR.
- Without Make, run `go build -ldflags "-X main.version=0.0.0-dev" -o bin/dockcrew ./cmd/dockcrew` and `go test ./...`.

## Coding Style & Naming Conventions
- Follow standard Go style: tabs for indentation, lower-case package names, and descriptive exported identifiers with comments when behaviour is non-obvious.
- Keep CLI flag names kebab-case (e.g., `--all`), and prefer short, action-oriented command verbs.
- Any generated assets should live under `internal/ui` or `internal/docker/templates` with `_gen.go` suffixes.

## Testing Guidelines
- Place tests alongside code as `*_test.go`; use table-driven tests for parsing and Docker interaction logic like `internal/docker/parse_test.go`.
- Mock external Docker calls using interfaces from `internal/docker` so tests remain hermetic.
- Strive to cover new branches touching container lifecycle commands; add regression tests before fixing bugs.

## Commit & Pull Request Guidelines
- Follow Conventional Commits (`feat:`, `chore:`, `fix:`) as seen in `git log`; scope nouns should map to package names when possible.
- Each PR should describe behaviour changes, link relevant TODOs or issues, and include manual test notes or screenshots for TUI changes.
- Do not commit binaries; add new tooling or scripts under `bin/` only if generated during CI.
- Request review once CI (build + test) passes locally and documentation in `docs/` reflects any user-facing updates.
