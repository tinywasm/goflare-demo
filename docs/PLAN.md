---
message: "feat: migrate to tinywasm/router contract and exercise D1 + router + R2 uploads"
---

> Este plan se despacha vía el flujo CodeJob. Ver skill: agents-workflow.

# PLAN — cola de ejecución de `goflare-demo`

> Si te han dicho *"ejecuta el plan descrito en docs/PLAN.md"*, ejecuta el plan de la tabla.
> Es autocontenido.

| Orden | Plan | Asunto |
|-------|------|--------|
| 1 | [PLAN_THREE_APIS.md](PLAN_THREE_APIS.md) | Reparar el `go.mod`, migrar al contrato `tinywasm/router`, marcar las rutas `.Public()`, y ejercitar la subida de archivos a R2. Termina con la **verificación real** en Cloudflare. |

## ✅ Compuerta — abierta

Este plan dependía de que `tinywasm/goflare` publicara sus dos etapas (router y archivos).
**Ya están publicadas en `goflare v0.4.1`** (2026-07-13): trae `goflare/edge`, `goflare/r2`,
`goflare/files` y el logging obligatorio del borde (todo 4xx/5xx sale con su causa, y un
pánico se recupera en vez de tumbar el Worker con un 1101). Ya **no** trae `goflare/pages`
ni `goflare/router`.

Sube la dependencia a `v0.4.1` o superior en el paso 1. Si tu `go.mod` sigue en `v0.3.6`,
los imports de `goflare/edge` no resuelven.

## Por qué existe este repo

`goflare-demo` es la **prueba de aceptación** de `goflare`: aquí se demuestra que las tres
APIs que de verdad se usan —**D1**, **router** y **archivos**— funcionan juntas **en
Cloudflare de verdad**, no solo en tests.

Estado verificado **2026-07-12**: el repo **no compila**. `modules/contact/list_handler.go`
importa `github.com/tinywasm/model` sin declararlo en `go.mod`, y las dependencias van muy
por detrás del ecosistema. Esa deuda se salda en el paso 1 del plan.
