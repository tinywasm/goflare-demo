# AGENTS.md — goflare-demo

## Compilación: NO correr `go build` manualmente

El servidor de desarrollo de tinywasm (lanzado con la herramienta/MCP de tinywasm)
**observa los archivos y recompila automáticamente** al guardar — tanto el frontend
(`web/`, wasm) como el edge (`edge/`). No ejecutes:

- `go build ./...`
- `GOOS=js GOARCH=wasm go build ...`
- ningún comando de compilación manual

Es redundante con el watcher y ensucia con artefactos.

## Verificación: usar el MCP de tinywasm

Para comprobar si un cambio funciona, usa las herramientas MCP en vez de compilar:

- `app_get_logs` — errores de build/compilación, panics WASM, estado del server.
- `browser_get_console` — `console.log` y errores de red del navegador.
- `browser_get_errors` — excepciones JS / panics WASM en runtime.
- `browser_get_content` — contenido renderizado de la página.
- `browser_navigate` — navegar a una URL.

Flujo típico tras editar: guardar → `app_get_logs` (build OK) → `browser_get_console`
/ `browser_get_errors` (runtime OK).

## Límite del entorno local

El server local sirve **solo estáticos desde `web/public`**. **No ejecuta las Pages
Functions** (`functions/edge.wasm` + `functions/[[path]].mjs`). Por eso las rutas API
(`/api/contacto`, etc.) devuelven **404 en local** — solo funcionan desplegadas en
Cloudflare Pages (o con un emulador de Functions). Un 404 en `/api/...` local es
esperado, no un bug del código.

## Artefactos wasm: qué se commitea y qué no

- **Commiteados (los necesita CF Git Integration)**: `functions/edge.wasm`,
  `functions/[[path]].mjs`, `web/public/client.wasm`.
- **NO commitear** (extraviados de builds desde el dir padre): `/edge.wasm`,
  `/web/client.wasm`. Están en `.gitignore`.
