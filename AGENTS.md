# AGENTS.md — goflare-demo

## Do NOT run `go build` manually

The tinywasm dev server (started via the tinywasm MCP tool) watches files and
recompiles automatically on save — both the frontend (`web/`, wasm) and the edge
(`edge/`). Never run:

- `go build ./...`
- `GOOS=js GOARCH=wasm go build ...`
- any manual build command

It is redundant with the watcher and pollutes the tree with stale artifacts.

## Verification: use the tinywasm MCP tools

To check whether a change works, use the MCP tools instead of compiling:

- `app_get_logs` — build/compile errors, WASM panics, server state.
- `browser_get_console` — `console.log` output and network errors.
- `browser_get_errors` — JS exceptions / WASM panics at runtime.
- `browser_get_content` — rendered page content.
- `browser_navigate` — navigate to a URL.

Typical flow after editing: save → `app_get_logs` (build OK) → `browser_get_console`
/ `browser_get_errors` (runtime OK).

## Use `tinywasm/fmt` — never `log` or `strings`

Binary size is a first-class constraint in this project. The stdlib `log` and `strings`
packages pull in large dependency trees that bloat both the host binary and the WASM
output. Use `tinywasm/fmt` instead for all string operations and output:

- `fmt.Println(...)` instead of `log.Print(...)`  / `log.Printf(...)`
- `fmt.Errf(...)` instead of `fmt.Errorf(...)` or `errors.New(...)`
- `fmt.HasPrefix(s, prefix)` instead of `strings.HasPrefix`
- `fmt.Convert(s).TrimPrefix(p).String()` instead of `strings.TrimPrefix`

This applies to **all files** in the project: `web/server.go`, `edge/main.go`,
`modules/`, `routes/`, etc.

## DB: initialize once at startup, inject via closure

Never open a DB connection inside a handler or on every request. The DB must be
constructed once at process/isolate startup and passed to handlers via closure:

- `web/server.go` → `d1.NewLocal(":memory:")` + `db.CreateTable(...)` → `routes.Register(r, db)`
- `edge/main.go` → `d1.NewEdge("DB")` + `db.CreateTable(...)` → `routes.Register(r, db)`

Handlers receive `*orm.DB` as a constructor parameter and return `router.HandlerFunc`.
`db_host.go` and `db_wasm.go` are **removed** — DB construction does not belong in
the business logic layer.

## SQLite local path: always use `:memory:`

Never use a relative file path (e.g. `"goflare-local.db"`) for the local SQLite DB.
The tinywasm dev harness does not guarantee that the binary's CWD is the project root,
so relative paths fail silently (502 on every request). Use `:memory:` for local dev —
data is ephemeral per process restart, which is acceptable for a contact form demo.

## API routes are served locally.

Since CHECK_PLAN.md Stage 2, `web/server.go` uses `devserver.ListenAndServe` with
`routes.Register`, so `/api/contacto` (POST and GET) is served locally at
`localhost:6060`. A 404 on `/api/*` in local dev is a bug, not expected behavior.

## CI/CD — single source of truth: `workflow/spec.go`

All CI/CD configuration lives in **`workflow/spec.go`**. Never edit
`.github/workflows/deploy.yml` or the Docker test install steps directly.

- `workflow.InstallScript(version)` — commands to install the `goflare` binary in CI.
- `workflow.DockerImage` — the container image used for both CI and local tests.
- `workflow.ReadGoflareVersion(gomodPath)` — reads the goflare version from `go.mod`.

**When goflare version changes:**
1. `go get github.com/tinywasm/goflare@vX.Y.Z`
2. `go generate ./workflow/` — regenerates `.github/workflows/deploy.yml`
3. Commit both `go.mod`, `go.sum`, and the updated `deploy.yml`.

**NEVER use `git tag` + `git push origin vX.Y.Z` alone to publish a goflare release.**
That creates a git tag without the binary attached to the GitHub Release, causing CI to fail
with `curl: (22) The requested URL returned error: 404` when trying to download the binary.
Always publish via `gorelease` (creates tag + compiles + attaches binary in one step).
If `gorelease` fails with "no tag was created" (nothing to commit), manually build and publish:
```bash
GOOS=linux GOARCH=amd64 go build -o /tmp/goflare-linux-amd64 ./cmd/goflare
git tag vX.Y.Z <commit-sha>
git push origin vX.Y.Z
gh release create vX.Y.Z /tmp/goflare-linux-amd64 --title "vX.Y.Z" --notes "..."
```

**To test CI locally** (requires Docker):
```bash
go test -tags=integration -run TestCIBuild_Docker ./tests/ -v
```
The test runs in a non-root Docker container matching the GitHub Actions environment.
If it passes locally, CI will pass.

## WASM artifacts: what to commit and what not to

- **Commit** (required by CF Git Integration): `functions/edge.wasm`,
  `functions/[[path]].mjs`, `web/public/client.wasm`.
- **Do NOT commit** (stray artifacts from builds in the parent dir): `/edge.wasm`,
  `/web/client.wasm`. Both are listed in `.gitignore`.
