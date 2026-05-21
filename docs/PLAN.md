# PLAN — Eliminar `SSRInstance()` del demo

## Objetivo

Quitar `func SSRInstance() *ContactForm { return &ContactForm{} }` del módulo
de contacto. El contrato real es la firma de `RenderCSS()`; la función
accesoria sólo existe para que el extractor externo construya el receiver, lo
cual puede deducirse de la firma del método.

## Justificación

`goflare-demo` es la referencia que copian los nuevos adoptantes de
`tinywasm/app`. Cada línea de boilerplate aquí se replica en cada proyecto
derivado. Quitar `SSRInstance` muestra el contrato mínimo real: sólo
`RenderCSS()` (y opcionalmente `RenderHTML`, `RenderJS`).

## Precondición técnica

Borrar `SSRInstance` rompe la compilación del extractor SSR upstream mientras
éste invoque `pkg.SSRInstance()` en el `main.go` que genera. Aplicar este plan
sólo cuando el extractor ya descubra el tipo receiver automáticamente desde la
firma del método (o acepte su ausencia como fallback). Verificación previa:

```bash
# El módulo del demo debe seguir compilando dentro del flujo del extractor.
go build ./...
```

## Archivos afectados

| Archivo | Cambio |
|---|---|
| [modules/contact/ssr.go:8-11](../modules/contact/ssr.go#L8-L11) | Borrar comentario y función `SSRInstance` |

`grep -rn SSRInstance .` en el repo del demo debe quedar vacío tras los cambios.

## Tests y validación

- `go build ./...` y `go test ./...` verdes.
- Arrancar dev server vía `mcp__tinywasm__start_development` y validar con
  `browser_screenshot` que el formulario de contacto se renderiza con los
  estilos correctos (light/dark) — sin regresiones visuales.

## Documentación

- Si el README del demo o `docs/CHECK_PLAN.md` referencia `SSRInstance` como
  parte del "patrón de módulo", actualizarlo para reflejar el contrato reducido.

## Migración adicional: `RenderJS()` → `[]*js.Script` (breaking change)

Si el demo añade un `RenderJS()` (por ejemplo para registrar un service
worker que convierta `goflare-demo` en una PWA), debe usar la firma nueva:
`func (c *ContactForm) RenderJS() []*js.Script`. Tres formas de construir los
elementos del slice (ver `github.com/tinywasm/js`):

- Bundle inline: `&js.Script{Content: rawJS}` (Name vacío).
- Standalone crudo (escape hatch): `&js.Script{Name: "raw.js", Content: rawJS}`.
- **Recomendado para SW/Worker:** constructores tipados —
  `js.ServiceWorker("sw.js", &MyAppSW{})` o
  `js.WebWorker("parser.worker.js", &ParserWorker{})`. El handler se
  implementa como interfaz Go (`OnFetch`, `OnInstall`, etc.) y tinywasm/js
  genera el JS-shim. Cero JS escrita por el usuario.

Estado actual: el demo no implementa `RenderJS()`. La acción aquí es:

- Ejemplo opcional pero recomendado: implementar un `ServiceWorkerHandler`
  Go mínimo (cache de estáticos para PWA offline) y usar
  `js.ServiceWorker("sw.js", &handler)` — valida E2E el shim generado y
  sirve de referencia copiable para adoptantes.

Precondición técnica: `tinywasm/js`, `tinywasm/dom`, `tinywasm/assetmin` y
`tinywasm/site` publicados con el contrato `[]*js.Script`. Verificación:

```bash
go list -m github.com/tinywasm/js github.com/tinywasm/dom github.com/tinywasm/assetmin github.com/tinywasm/site
```

## Stages

| # | Tarea | Done |
|---|---|---|
| 1 | Confirmar precondición técnica (`go build ./...` verde) | [ ] |
| 2 | Borrar `SSRInstance` en `modules/contact/ssr.go` | [ ] |
| 3 | Confirmar precondición `[]*js.Script` publicada en js/dom/assetmin/site | [ ] |
| 4 | (Opcional) Añadir `RenderJS()` de ejemplo con un `sw.js` standalone | [ ] |
| 5 | `go test ./...` y `go build ./...` verde | [ ] |
| 6 | Screenshot SSR vía MCP browser — sin regresiones | [ ] |
| 7 | Validar registro de SW vía DevTools si se añadió el ejemplo | [ ] |
| 8 | Actualizar README/CHECK_PLAN si referencia `SSRInstance` o `RenderJS` | [ ] |
