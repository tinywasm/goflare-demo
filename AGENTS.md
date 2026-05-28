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

## WASM artifacts: what to commit and what not to

- **Commit** (required by CF Git Integration): `functions/edge.wasm`,
  `functions/[[path]].mjs`, `web/public/client.wasm`.
- **Do NOT commit** (stray artifacts from builds in the parent dir): `/edge.wasm`,
  `/web/client.wasm`. Both are listed in `.gitignore`.
