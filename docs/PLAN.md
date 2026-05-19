# PLAN — Migración de `goflare-demo` a Pages Functions

> Objetivo: reescribir `goflare-demo` **in-place** para que sea el demo end-to-end de la nueva funcionalidad **Pages Functions** de `tinywasm/goflare` (ver [/home/cesar/Dev/Project/tinywasm/goflare/docs/PLAN.md](../../goflare/docs/PLAN.md)). Este demo es el vehículo de validación E2E del MVP.

## 1. Contexto

El demo actual está pensado para **Cloudflare Workers** (`edge/main.go` con `workers.Handle`) y **nunca se ha probado en producción**. El usuario lo ha indicado explícitamente.

Mientras tanto, [veltylabs/website](../../../veltylabs/website) corre en producción usando el patrón **Cloudflare Pages Functions** con `functions/api/contacto.js` (JS escrito a mano). Ese es el modelo que `goflare` automatizará para Go.

`goflare-demo` debe pasar a usar ese modelo para:

- Validar que la nueva API `pages.Serve(http.Handler)` funciona end-to-end.
- Servir como ejemplo canónico en el README de `goflare`.
- Detectar gaps reales (CORS, errores, status codes, secrets, fetch outbound) antes de declarar el MVP listo.

veltylabs **no se toca**; este demo es independiente y vive en su propio repo.

## 2. Estado actual del demo

```
goflare-demo/
├── edge/main.go            # ← Workers handler con workers.Handle (a borrar)
├── modules/contact/        # ← Modelo compartido (se conserva)
├── web/                    # ← Frontend WASM + estáticos (se conserva)
│   ├── client.go
│   ├── server.go
│   └── public/
└── docs/
```

## 3. Decisiones

| # | Decisión | Por qué |
|---|---|---|
| D1 | **Reescribir in-place** (no branch ni repo nuevo) | El demo nunca se probó en prod; no hay nada que preservar. Mantener un solo source of truth simplifica docs. |
| D2 | **Caso de uso copiado de veltylabs**: form de contacto → validación → fetch a Resend → respuesta JSON | Es un caso real, no un hello-world. Si esto funciona, el MVP funciona. |
| D3 | **Frontend WASM existente se conserva** | El demo ya valida `tinywasm/dom` + `tinywasm/form` integrados; agregar Pages Functions complementa, no reemplaza. |
| D4 | **`modules/contact/` compartido entre frontend y handler** | Modelo + validación reutilizados — uno de los argumentos centrales de "Go full-stack". |
| D5 | **CF Git Integration + artefactos commiteados** (sin GitHub Actions, sin tokens) | Alineado con D7/D8 del [PLAN.md de goflare](../../goflare/docs/PLAN.md): el demo muestra el flujo recomendado. Setup one-time de CF dashboard, después `git push` despliega. Cero secrets distribuidos. |
| D6 | **Único entrypoint `edge/main.go`** (no `pages/main.go`); modo inferido de sus imports | Alineado con D10/D11 de goflare. Reduce ruido en el repo; mismo path que el modo Workers actual. `edge/main.go` importa `tinywasm/goflare/pages` → goflare infiere modo `pages-functions` sin necesidad de variable en `.env`. |

## 4. Estructura objetivo

```
goflare-demo/
├── .env                       # PROJECT_NAME, CLOUDFLARE_ACCOUNT_ID  (sin MODE — se infiere de imports)
├── .env.example
├── go.mod
├── routes/
│   └── routes.go              # NUEVO — build-agnóstico: func Register(r router.Router)
├── modules/
│   └── contact/
│       ├── model.go           # ContactForm + Validate (sin cambios, sin build tags)
│       ├── model_orm.go       # generado por ormc
│       └── handler.go         # NUEVO — build-agnóstico: func Handle(w, r)
├── web/
│   ├── client.go              # //go:build wasm — frontend (sin cambios)
│   ├── server.go              # //go:build !wasm — dev local; ahora llama routes.Register
│   └── public/
│       ├── index.html         # dev
│       ├── client.wasm        # producido por el framework tinywasm desde web/client.go, COMMITEADO
│       ├── script.js          # producido por el framework tinywasm (assetmin), COMMITEADO
│       └── style.css          # producido por el framework tinywasm (assetmin), COMMITEADO
├── edge/
│   └── main.go                # //go:build wasm — ÚNICO entrypoint (5 líneas: routes.Register + pages.Serve)
└── functions/                 # ÚNICO output propio de goflare, COMMITEADO
    ├── [[path]].mjs           # ← generado por goflare build
    └── edge.wasm              # ← generado por goflare build (TinyGo de edge/main.go)
```

**Cambios respecto al diseño anterior** (alineado con goflare/PLAN.md D7-D11):

- **`edge/main.go`** en lugar de `pages/main.go` — único entrypoint, mismo path que en modo Workers (D10).
- **Modo inferido de imports** (D11): `edge/main.go` importa `tinywasm/goflare/pages` → goflare sabe que es `pages-functions`. Sin variable `MODE` en `.env`.
- **`functions/` y `web/public/*` se commitean** (D8). `.gitignore` solo cubre `.env`, `.build/`, archivos de IDE.
- **Sin GitHub Actions**: CF Git Integration despliega lo commiteado en cada push (D7).

**Regla**: la lógica de negocio de cada endpoint vive en `modules/<feature>/handler.go` SIN build tags. `routes/routes.go` solo registra rutas URL→handler (aggregator). `edge/main.go` y `web/server.go` son entrypoints triviales que llaman `routes.Register(r)`. Cero duplicación, cero drift dev↔prod.

## 5. Diseño — tres archivos clave

> **Restricción D12 de goflare**: el código en `//go:build wasm` (handlers, edge/main.go, modules/*/handler.go cuando se llaman desde wasm) NO puede importar stdlib pesada (`fmt`, `strings`, `errors`, `encoding/json`, `net/http`, `log`, etc.). Solo `tinywasm/*` y primitivas mínimas. Stdlib infla el binario +80% y excede 1 MiB de CF Free.

### 5.1 `modules/contact/handler.go` (build-agnóstico — la lógica real)

```go
package contact

import (
    "github.com/tinywasm/goflare/cloudflare"
    "github.com/tinywasm/goflare/router"
    "github.com/tinywasm/json"   // ← tinywasm, NO encoding/json
)

func Handle(ctx router.Context) {
    ctx.SetHeader("Content-Type", "application/json")
    ctx.SetHeader("Access-Control-Allow-Origin", "*")

    if ctx.Method() == "OPTIONS" {
        ctx.SetHeader("Access-Control-Allow-Methods", "POST, OPTIONS")
        ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type")
        ctx.WriteStatus(204)
        return
    }
    if ctx.Method() != "POST" {
        ctx.WriteStatus(405)
        ctx.Write([]byte(`{"error":"method not allowed"}`))
        return
    }

    var data ContactForm
    if err := json.Decode(ctx.Body(), &data); err != nil {
        ctx.WriteStatus(400)
        ctx.Write([]byte(`{"error":"invalid json"}`))
        return
    }
    if err := data.Validate(0); err != nil {
        ctx.WriteStatus(422)
        ctx.Write([]byte(`{"error":"` + err.Error() + `"}`))
        return
    }

    if err := sendEmail(data, cloudflare.Env("RESEND_API_KEY")); err != nil {
        ctx.WriteStatus(502)
        ctx.Write([]byte(`{"error":"email delivery failed"}`))
        return
    }

    ctx.WriteStatus(200)
    ctx.Write([]byte(`{"message":"¡Gracias! Hemos recibido tu solicitud."}`))
}

func sendEmail(data ContactForm, apiKey string) error {
    // TODO: tinywasm/fetch para POST a https://api.resend.com/emails
    return nil
}
```

### 5.2 `routes/routes.go` (build-agnóstico — aggregator central de rutas URL→handler)

```go
package routes

import (
    "github.com/tinywasm/goflare/router"
    "github.com/tinywasm/goflare-demo/modules/contact"
)

func Register(r router.Router) {
    r.Post("/api/contacto", contact.Handle)
    r.Options("/api/contacto", contact.Handle) // CORS preflight
}
```

### 5.3 `edge/main.go` (edge, wasm — trivial; mismo path que modo Workers)

```go
//go:build wasm

package main

import (
    "github.com/tinywasm/goflare/pages"
    "github.com/tinywasm/goflare-demo/routes"
)

func main() {
    r := pages.NewRouter()
    routes.Register(r)
    pages.Serve(r)
}
```

### 5.4 `web/server.go` (dev local — invoca la misma `routes.Register`)

```go
//go:build !wasm

package main

import (
    "github.com/tinywasm/goflare/pages/devserver" // adapter sobre net/http (stdlib OK acá)
    "github.com/tinywasm/goflare-demo/routes"
)

func main() {
    r := devserver.NewRouter()
    routes.Register(r)
    devserver.ListenAndServe(":8080", r, "web/public")
}
```

**Misma URL, mismo handler, mismo `routes.Register`** — la única diferencia es qué implementación de `router.Router` se inyecta. El handler no sabe si está corriendo en wasm o en stdlib.

## 6. Frontend: cambios mínimos

- `web/public/index.html`: eliminar `window.WORKER_URL` — ya no hace falta. El form postea a `/api/contacto` (mismo origin).
- `web/client.go`: si tenía URL absoluta para POST, cambiar a path relativo `/api/contacto`.

## 7. Roadmap

### Fase 0 — Bloqueado por goflare
Esperar a que `tinywasm/goflare` complete las fases 1-3 de su [PLAN.md](../../goflare/docs/PLAN.md) (Runtime Go, Build pipeline, Deploy): paquete `pages/`, `cloudflare/`, y `goflare build` con modo Pages-Functions.

### Fase 1 — Migración estructural
- [ ] Reescribir `edge/main.go`: contenido nuevo con `routes.Register(mux) + pages.Serve(mux)` (§5.3). Mismo path que antes, código distinto.
- [ ] Crear `modules/contact/handler.go` con `Handle(w, r)` build-agnóstico (§5.1).
- [ ] Crear `routes/routes.go` con `Register(r router.Router)` (§5.2).
- [ ] Modificar `web/server.go` para llamar `routes.Register(mux)` antes del fileserver (§5.4).
- [ ] Actualizar `go.mod` (eliminar `goflare/workers`, agregar `goflare/pages` + `goflare/cloudflare`).
- [ ] Escribir `.env` con `PROJECT_NAME`, `CLOUDFLARE_ACCOUNT_ID` (sin `MODE` — se infiere de los imports de `edge/main.go`, D11). `.env.example` documenta `RESEND_API_KEY` (que vive en CF dashboard, no en `.env`).
- [ ] Ajustar `.gitignore`: incluir `.env` y `.build/`; **NO** ignorar `functions/` ni `web/public/*.wasm/css/js`.
- [ ] Verificar: `go run web/server.go` levanta API local + estáticos en `:8080` sin tocar wasm.

### Fase 2 — Frontend
- [ ] Eliminar `window.WORKER_URL` de `index.html`.
- [ ] Path relativo `/api/contacto` en `web/client.go`.

### Fase 3 — Integración real con Resend
- [ ] Implementar `sendEmail()` usando `tinywasm/fetch` (o helper que invoque el `fetch` global del runtime JS).
- [ ] Validar dominio en Resend (instrucciones en README).

### Fase 4 — Validación E2E
- [ ] `goflare build` produce `functions/[[path]].mjs` + `functions/edge.wasm`, sin otros archivos.
- [ ] `goflare deploy` despliega a Cloudflare Pages.
- [ ] Test manual: submit del form → recibo email → respuesta 200 al navegador.
- [ ] Test de error: payload inválido → 422 con mensaje claro.
- [ ] Test CORS preflight.

### Fase 5 — CF Git Integration (setup one-time)
- [ ] En `dash.cloudflare.com` → Pages → Create project → Connect to Git → seleccionar repo del demo.
- [ ] Build command: **vacío**. Build output directory: `web/public`. Production branch: `main`.
- [ ] CF Pages → Settings → Environment variables → agregar `RESEND_API_KEY` (Production + Preview).
- [ ] Verificar: `goflare build && git add . && git commit && git push` dispara deploy automático en CF, sin intervención manual ni tokens.

### Fase 6 — Docs
- [ ] README reescrito: nuevo flujo `goflare build && git push` → CF despliega. Sin Workers, sin `wrangler`, sin Actions.
- [ ] Diagrama actualizado en `docs/img/`.
- [ ] Capturas del form en producción.

## 8. Criterios de éxito (DoD)

1. `tree -L 2` muestra `edge/`, `routes/`, `modules/`, `functions/` (commiteado).
2. El comando `goflare build` produce **exactamente 2 archivos** dentro de `functions/`: `[[path]].mjs` y `edge.wasm`.
3. `git status` tras `goflare build` muestra `functions/` y `web/public/*.wasm/css/js` como cambios commiteables (no ignorados).
4. Un `git push` a `main` dispara deploy automático de CF Pages, sin Actions ni tokens en local.
5. Un POST real al endpoint en producción entrega un email vía Resend.
6. El demo replica funcionalmente al endpoint de [veltylabs/website/functions/api/contacto.js](../../../veltylabs/website/functions/api/contacto.js), pero escrito en Go.
7. Un dev nuevo puede clonar el repo, modificar código, `goflare build`, `git push` — y desplegar sin que nadie le entregue secrets.

## 9. Out of scope

- D1/KV/R2: no aporta al caso form-de-contacto.
- Tests automatizados de integración contra Cloudflare (mucho overhead para un demo; test manual basta).
- Multi-endpoint (newsletter, etc.) — fuera del MVP; el ejemplo se mantiene minimal.
- **Conservar el código Workers anterior**: por D1 (reescribir in-place), `edge/main.go` se borra. Si en el futuro hace falta un demo de Workers puro, será un proyecto aparte.
