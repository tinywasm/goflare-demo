---
message: "feat: the demo proves four APIs — router, D1, files and real Google authentication"
---

> Este plan se despacha vía el flujo CodeJob. Ver skill: agents-workflow.
> Orquestado por `tinywasm/docs/DEMO_FOUR_APIS_MASTER_PLAN.md` — **Fase D (aceptación)**.

# PLAN — cola de ejecución de `goflare-demo`

> Si te han dicho *"ejecuta el plan descrito en docs/PLAN.md"*, ejecuta el **primer plan
> pendiente** de la tabla. Es autocontenido.

| Orden | Plan | Estado | Asunto |
|-------|------|--------|--------|
| 1 | [PLAN_THREE_APIS.md](PLAN_THREE_APIS.md) | ✅ **COMPLETADA** (destapó dos bugs de librería) | Reparar `go.mod`, migrar al contrato `tinywasm/router`, marcar las rutas `.Public()`, conectar la subida a R2. |
| 2 | [PLAN_FOUR_APIS.md](PLAN_FOUR_APIS.md) | ☐ **BLOQUEADA** — espera Fases E y F | Login real con Google, el formulario pasa a ser registro, y **un archivo por usuario que se reemplaza**. Demuestra las **cuatro** APIs: router, D1, archivos y **autenticación**. |

## Para qué existe este repo

Es la **prueba de aceptación** del ecosistema: aquí se demuestra que las APIs funcionan **en
Cloudflare de verdad**, no solo en tests. Y está cumpliendo su función — cada plan que se
ejecuta aquí destapa una mentira en las librerías:

**El plan 1 encontró tres:**

1. **`goflare/edge` ejecutaba la verja RBAC ANTES de los middlewares** → ningún llamante podía
   identificarse jamás → **toda ruta con `.Requires()` era un 403 eterno**, y con ella la API
   de archivos entera. Sus tests pasaban porque el fake tenía un `SetUserID` vacío.
2. **`server/httpd` registraba en el `ServeMux` sin el método** → tres métodos sobre un mismo
   path eran el mismo patrón → **panic al arrancar**. El demo lo había rodeado con un dispatch
   de método a mano.
3. **La causa de fondo de ambas**: `tinywasm/router` publicaba una interfaz (tipos) **sin
   arnés** (comportamiento). Dos implementaciones divergentes y nada que las obligara a
   coincidir. Se cerró con `router/conformance`, la suite ejecutable que toda implementación
   debe pasar (ver `tinywasm/docs/ROUTER_CONFORMANCE_MASTER_PLAN.md`).

**El plan 2, antes siquiera de empezar, ha encontrado la cuarta:**

4. **`tinywasm/user` se declaraba "edge-ready" y no lo es.** Compila a wasm con el compilador
   de Go, pero **el borde se compila con TinyGo**, y ahí `golang.org/x/oauth2` → `net/http` no
   existe. El módulo de autenticación **no puede entrar en un Worker**. Lo arregla la Fase E.

**La lección, y vale para todo el ecosistema: en el borde, TinyGo es el compilador que
decide.** `go build` y `GOOS=js go build` dan verde sobre código que TinyGo rechaza.
