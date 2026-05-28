> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan: goflare-demo — API + D1 en local (mismo endpoint que el edge)

## Objetivo

Que el servidor local sirva las **mismas rutas y handlers** que el edge
(`/api/contacto` POST/GET) contra la **D1 real**, para poder probar el flujo completo
—incluida la persistencia en la base— **antes de desplegar**.

Hoy el server local ([web/server.go](../web/server.go)) es un `http.FileServer` pelado:
nunca llama a `routes.Register`, así que las rutas API dan **404 en local**. Y el acceso
a D1 en host ([modules/contact/db_host.go](../modules/contact/db_host.go)) es un stub que
devuelve `errHostOnly`. Resultado: no se puede probar la API ni la DB en local.

## Pieza clave que YA existe

`goflare/devserver` ([devserver.go](../../goflare/devserver/devserver.go),
`//go:build !wasm`) implementa `router.Context`/`router.Router` sobre `net/http` y
expone `ListenAndServe(addr, r, staticDir)` que sirve las rutas registradas **y** los
estáticos. Ya aplica **no-cache** a los estáticos (correcto para dev: tras recompilar,
el navegador trae el `.wasm` nuevo, no uno cacheado). No hay que construir nada — solo
conectarlo desde `server.go`.

## Por qué esto permite probar la DB en local

El handler `contact.Handle`/`HandleList` es agnóstico de build. En el edge usa
`d1.New("DB")` (binding JS); en host usará `d1.NewDirect(token, accountID, dbID)` (REST).
Ambos construyen un `*orm.DB` con el **mismo `sqlt.NewCompiler()`** → el SQL generado es
idéntico. Lo que pruebas en local es exactamente lo que correrá en el edge.

> **Salvedad**: `d1.NewDirect` pega a la D1 **real** vía REST; los writes locales van a
> esa base. Para un demo está bien (persiste datos de prueba). Si quieres aislamiento,
> exporta el `D1_DATABASE_ID` de una base de dev distinta.

---

## Stages

### Stage 1 — Prerrequisito en goflare (HECHO y publicado)

`devserver.ListenAndServe` ya envuelve el servido de estáticos con headers no-cache
(cambio + test en `goflare/devserver/`). **Publicado en goflare v0.2.15** y el `go.mod`
del demo ya apunta a esa versión. No requiere acción adicional.

### Stage 2 — `web/server.go`: colapsar al dev server de goflare

Reescribir `main()` para registrar las rutas del edge y delegar todo (API + estáticos +
no-cache) a `devserver.ListenAndServe`. Conservar `lookupArg("server_port")` y
`lookupArg("server_public_dir")`. **Eliminar** el `http.FileServer`, el `gzipHandler` y
el `noCache` artesanales (el gzip en local no aporta — Cloudflare lo hace en prod; el
no-cache ahora lo da devserver).

```go
//go:build !wasm

package main

import (
	"log"
	"os"

	"github.com/tinywasm/fmt" // NO usar stdlib strings — convención tinywasm
	"github.com/tinywasm/goflare/devserver"
	"github.com/tinywasm/goflare-demo/routes"
)

// lookupArg lee -key=value o -key value de os.Args. Usa tinywasm/fmt, no strings.
func lookupArg(key string) string {
	prefix := "-" + key + "="
	args := os.Args[1:]
	for i, arg := range args {
		if fmt.HasPrefix(arg, prefix) {
			return fmt.Convert(arg).TrimPrefix(prefix).String()
		}
		if arg == "-"+key && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func main() {
	port := lookupArg("server_port")
	if port == "" {
		port = "6060"
	}
	publicDir := lookupArg("server_public_dir")
	if publicDir == "" {
		publicDir = "web/public"
	}

	r := devserver.NewRouter()
	routes.Register(r) // MISMAS rutas/handlers que el edge (edge/main.go)

	log.Printf("Dev server on :%s — static: %s, API: /api/*", port, publicDir)
	if err := devserver.ListenAndServe(":"+port, r, publicDir); err != nil {
		log.Fatal(err)
	}
}
```

`lookupArg` se reescribe con `tinywasm/fmt` (la versión actual usa `strings`). Se
**borran** del `server.go` actual: el `struct gzipResponseWriter`, las funcs
`gzipHandler`/`noCache`, el `http.FileServer` y el `mux`, con sus imports huérfanos
(`compress/gzip`, `io`, `net/http`, `strings`).

> `routes.Register` es el MISMO que usa el edge ([edge/main.go](../edge/main.go)) → un
> solo set de rutas y handlers para ambos entornos.

### Stage 3 — `db_host.go`: D1 real vía REST

Reemplazar los stubs por una implementación que espeja `db_wasm.go`, cambiando solo
cómo se obtiene el `*orm.DB` (credenciales desde env vars):

```go
//go:build !wasm

package contact

import (
	"os"

	"github.com/tinywasm/goflare/d1"
	"github.com/tinywasm/orm"
)

func hostDB() (*orm.DB, error) {
	return d1.NewDirect(
		os.Getenv("CLOUDFLARE_API_TOKEN"),
		os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		os.Getenv("D1_DATABASE_ID"),
	)
}

func saveSubmission(sub *Contact) error {
	db, err := hostDB()
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.CreateTable(sub); err != nil {
		return err
	}
	return db.Create(sub)
}

func listSubmissions() (*ContactList, error) {
	db, err := hostDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	qb := db.Query(&Contact{}).OrderBy("id").Desc()
	return ReadAllContact(qb)
}
```

### Stage 4 — Correr en local con DB

El dev server de tinywasm ejecuta `server.go`. Para que el host alcance la D1, exportar
las credenciales antes de arrancar (NO en `.env` — son secretos):

```bash
export CLOUDFLARE_API_TOKEN=cfat_...
export CLOUDFLARE_ACCOUNT_ID=...
export D1_DATABASE_ID=...
```

Verificación (vía MCP, sin compilar a mano — ver [AGENTS.md](../AGENTS.md)):
- `browser_get_console` / `browser_get_errors`: ya no hay 404 ni "expected array" en `/api/contacto`.
- Enviar el formulario → POST 200 → recargar → la lista muestra la submission (leída de D1).
- `app_get_logs`: sin panics de build.

---

## Resumen de archivos

| Archivo | Acción |
|---|---|
| `goflare/devserver/devserver.go` | **HECHO y publicado** (goflare v0.2.15) — no-cache en `ListenAndServe` + test |
| `goflare-demo/go.mod` | **HECHO** — ya en goflare v0.2.15 |
| `web/server.go` | Colapsar a `devserver.NewRouter()` + `routes.Register` + `ListenAndServe`; `lookupArg` con tinywasm/fmt; borrar gzip/noCache/mux/strings |
| `modules/contact/db_host.go` | D1 real vía `d1.NewDirect` (env vars), espejando `db_wasm.go` |

---

## Verification

- POST `/api/contacto` en local → 200 + `{"message":"¡Gracias!..."}` y registro en D1.
- GET `/api/contacto` en local → array JSON con las submissions reales.
- El frontend muestra la lista en `localhost:6060` igual que en producción.
- El edge sigue intacto: `edge/main.go` usa `pages` (wasm) + `d1.New("DB")`.
- El dev server de tinywasm auto-compila; verificar con el MCP, no con `go build`.
