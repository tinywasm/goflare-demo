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

`goflare/pages/devserver` ([devserver.go](../../goflare/pages/devserver/devserver.go),
`//go:build !wasm`) ya implementa `router.Context` y `router.Router` sobre `net/http`.
No hay que construir el adaptador — solo conectarlo. Falta exponer su `http.Handler`
para poder componerlo con el middleware de dev (no-cache, gzip) y el servido de
estáticos del demo.

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

### Stage 1 — goflare: exponer el handler de devserver

En `goflare/pages/devserver/devserver.go`, añadir un accesor para componer el mux con
middleware externo (sin forzar `ListenAndServe`):

```go
import "net/http"

// Handler returns the router's configured http.Handler so callers can compose it
// with their own middleware (no-cache, gzip) and static file serving.
func Handler(r router.Router) http.Handler {
	return r.(*nativeRouter).mux
}
```

Publicar nueva versión de goflare y actualizar `go.mod` del demo a esa versión.

### Stage 2 — demo `web/server.go`: montar las rutas del edge

Reescribir `main()` para componer: rutas API vía devserver + estáticos con el
middleware de dev existente. Conservar `lookupArg("server_port")` y
`lookupArg("server_public_dir")`, y el `noCache`/`gzipHandler` actuales.

```go
import (
	"net/http"
	"github.com/tinywasm/goflare/pages/devserver"
	"github.com/tinywasm/goflare-demo/routes"
)

// ...dentro de main(), tras resolver port y publicDir...

// 1. Router del edge montado sobre net/http (mismos handlers).
apiRouter := devserver.NewRouter()
routes.Register(apiRouter)

mux := http.NewServeMux()
// 2. Delegar /api/* al handler de devserver (matchea "POST /api/contacto", etc.).
mux.Handle("/api/", devserver.Handler(apiRouter))
// 3. El resto: estáticos con no-cache + gzip (igual que hoy).
mux.Handle("/", noCache(gzipHandler(fs)))

server := &http.Server{Addr: ":" + port, Handler: mux}
log.Printf("Starting server on port %s", port)
if err := server.ListenAndServe(); err != nil {
	log.Fatal(err)
}
```

> `routes.Register` es el MISMO que usa el edge ([edge/main.go](../edge/main.go)) → un
> solo set de rutas y handlers para ambos entornos.

### Stage 3 — demo `db_host.go`: D1 real vía REST

Reemplazar los stubs por una implementación que espeja `db_wasm.go`, cambiando solo
cómo se obtiene el `*orm.DB` (credenciales desde env vars):

```go
//go:build !wasm

package contact

import (
	"os"
	"github.com/tinywasm/goflare/d1"
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

(Requiere importar `github.com/tinywasm/orm` para el tipo de retorno de `hostDB`.)

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
| `goflare/pages/devserver/devserver.go` | **goflare** — añadir `Handler(r) http.Handler`; publicar versión |
| `goflare-demo/go.mod` | Actualizar goflare a la versión publicada |
| `web/server.go` | Montar `devserver.NewRouter()` + `routes.Register` + `Handler` en `/api/`; estáticos con no-cache/gzip en `/` |
| `modules/contact/db_host.go` | D1 real vía `d1.NewDirect` (env vars), espejando `db_wasm.go` |

---

## Verification

```bash
# host: el dev server auto-compila; verificar con MCP (no go build manual)
# edge: la compilación wasm sigue usando d1.New (binding) — sin cambios
```

- POST `/api/contacto` en local → 200 + `{"message":"¡Gracias!..."}` y registro en D1.
- GET `/api/contacto` en local → array JSON con las submissions reales.
- El frontend muestra la lista en `localhost:6060` igual que en producción.
- El edge sigue intacto: `edge/main.go` usa `pages` (wasm) + `d1.New("DB")`.
