# Plan — `goflare-demo` sobre las APIs actuales: D1 + router + archivos

> Este plan se despacha vía el flujo CodeJob. Ver skill: agents-workflow.
>
> Es la **prueba de aceptación** de todo el trabajo hecho en `tinywasm/goflare`: si las tres
> APIs no funcionan aquí, no funcionan.
>
> **Prerrequisito, ya cumplido:** `goflare` publicó sus dos etapas en **v0.4.0**
> (2026-07-13). Esa versión trae `goflare/edge`, `goflare/r2` y `goflare/files`, y **ya no
> trae `goflare/pages` ni `goflare/router`**. Usa `goflare v0.4.0` o superior.
> Contexto opcional: https://github.com/tinywasm/goflare/blob/main/docs/PLAN.md

Autocontenido, en español.

---

## Estado de partida: el demo está roto AHORA MISMO

`go build ./...` falla, y no por los cambios que vienen:

```
modules/contact/list_handler.go:3:8: no required module provides package github.com/tinywasm/model
```

Alguien añadió `import "github.com/tinywasm/model"` en
[modules/contact/list_handler.go](../modules/contact/list_handler.go) sin declararlo en
`go.mod`. Además el `go.mod` va muy por detrás del ecosistema: `goflare v0.3.6`,
`orm v0.9.18`, `fmt v0.24.6`, `json v0.5.6`, `sqlite v0.2.3`.

O sea que el demo arrastra **dos deudas a la vez**: la que ya tenía, y la que le llega de
las etapas 1 y 2. Este plan las salda juntas.

---

## Cambios

### 1. Reparar el `go.mod` y subir el ecosistema

Añadir `github.com/tinywasm/model` (hoy se importa sin declarar) y actualizar todas las
dependencias de `tinywasm/*` a las versiones publicadas. Verificar con `go build ./...`
**antes** de tocar nada más: hay que separar "estaba roto" de "lo rompió la migración".

### 2. Migrar el router al contrato `tinywasm/router` (etapa 1)

Tres archivos importan el fork borrado `github.com/tinywasm/goflare/router`:

- [routes/routes.go](../routes/routes.go)
- [modules/contact/handler.go](../modules/contact/handler.go)
- [modules/contact/list_handler.go](../modules/contact/list_handler.go)

Pasan a `github.com/tinywasm/router`. Las firmas (`router.Context`, `router.HandlerFunc`)
no cambian de nombre — cambia de dónde vienen.

Y en [edge/main.go](../edge/main.go): `goflare/pages` → `goflare/edge`, con
`edge.NewRouter()` / `edge.Serve(r)` en vez de `pages.*`. Si esto se olvida, la build
**falla ruidosamente** (así se diseñó el endurecimiento de `inferMode` en la etapa 1); antes
habría emitido el artefacto equivocado en silencio.

### 3. ⚠️ Marcar las rutas públicas — o el demo devuelve 403 entero

El contrato nuevo es **privado por defecto**: una ruta que no llama `.Public()` ni
`.Requires()` responde **403** ante una petición sin identidad. El demo **no tiene
autenticación**, así que hoy sus tres rutas son públicas de facto:

```go
// routes/routes.go — DESPUÉS
func Register(r router.Router, db *orm.DB) {
    r.Post("/api/contacto", contact.Handle(db)).Public()
    r.Get("/api/contacto", contact.HandleList(db)).Public()
    r.Options("/api/contacto", contact.Handle(db)).Public()
}
```

Sin los `.Public()`, el demo compila, despliega… y **rechaza todas las peticiones con 403**.
Es el fallo más fácil de cometer en toda la migración.

### 4. Ejercitar la API de archivos (etapa 2)

El demo **no sube archivos hoy** — no hay nada que migrar, hay que añadirlo. Es la única
forma de demostrar la tercera API.

**⚠️ NO escribas los handlers de subida y servido. Ya existen:** `goflare/files` los trae
hechos, con la validación por bytes mágicos y la generación de clave dentro. Tu trabajo aquí
es **conectarlos**, no reimplementarlos.

```go
import (
	"github.com/tinywasm/goflare/files"
	"github.com/tinywasm/goflare/r2"
)

bucket, err := r2.NewEdge("FILES")   // el binding declarado en wrangler; falla ruidoso si no está
if err != nil {
	// no lo silencies: un bucket ausente es un error de configuración
}

store, err := files.New(bucket, "/api/files/")   // el prefijo DEBE terminar en "/"
if err != nil {
	// ...
}
store.Mount(r)   // registra las dos rutas sobre el mismo router.Router
```

`Mount` registra exactamente esto, y por eso no lo escribes tú:

- `PUT /api/files/` → **privado**, exige el permiso `files`/`write`. Valida el cuerpo contra
  `filetype.Images` (PNG, JPEG, GIF, WebP): si los bytes no son una imagen ráster responde
  **415 y no escribe nada en el bucket**. La clave la **genera el servidor** (`unixid` +
  la extensión deducida de los bytes) y la devuelve en el cuerpo de la respuesta 201.
- `GET /api/files/<clave>` → **público** (un `<img src>` no puede mandar cabeceras). Sirve
  con el `Content-Type` real deducido al subir y con `X-Content-Type-Options: nosniff`.

Ajustes disponibles si los necesitas: `store.MaxSize(n)` (por defecto 10 MiB) y
`store.Allow(...)`. **Nunca metas SVG ni HTML en la lista blanca:** llevan JavaScript dentro
y, servidos desde tu dominio, se ejecutan en tu origen.

Lo que sí es trabajo tuyo:

- Declarar el binding R2 (`FILES`) en la configuración de wrangler, junto al de D1.
- **Frontend:** subir con `tinywasm/fetch` mandando **los bytes crudos** como cuerpo del
  `PUT` (nada de `multipart` — decisión de la etapa 2). La respuesta 201 trae la clave: úsala
  para mostrar la imagen de vuelta con `<img src="/api/files/{clave}">`.
- **No mandes el nombre del archivo como clave.** El servidor la ignora a propósito: el
  nombre que elige el cliente es tan poco fiable como su `Content-Type`.

### 5. D1 no se toca

Funciona y está verificado. El módulo `contact` sigue siendo la prueba de la primera API.
Lo único que le cambia es de dónde importa `router`.

---

## Cómo se prueba (no lo improvises)

`gotest`, **nunca `go test`**. El código del borde habla con `js.Global()`, no con Cloudflare:
se prueba en navegador inyectando un `context.env` falso.

Estrategia completa (la misma que rige en `goflare`):
https://github.com/tinywasm/goflare/blob/main/docs/TESTING.md

---

## ✅ Verificación REAL — el motivo de que este repo exista

Los tests demuestran que **nuestro Go** es correcto. **No demuestran que Cloudflare se
comporte como creemos.** Ese es el trabajo de este repo, y se hace en dos niveles.

### Nivel A — Local con bindings reales (`wrangler dev`, sin desplegar)

`wrangler dev --local` arranca **miniflare** en segundos, con **D1 y R2 emulados de verdad**:
sin red, sin cuenta de Cloudflare, sin desplegar nada. Es el nivel donde se caza que **el
fake mentía**.

```bash
goflare build
wrangler dev --local
```

Con el worker corriendo en local, ejecuta las cuatro comprobaciones de abajo contra
`http://localhost:8787`.

### Nivel B — Desplegado de verdad

Solo cuando el Nivel A esté verde. Es el único momento en que desplegar está justificado —
y no es un test, es la confirmación final.

### Las cuatro comprobaciones (idénticas en Nivel A y Nivel B)

**1. D1 — el formulario guarda y lista.**
```bash
curl -X POST $BASE/api/contacto -d '{"name":"Ana","email":"a@b.c"}'
curl $BASE/api/contacto     # → debe aparecer Ana
```

**2. Router — las rutas responden 200, NO 403.**
Si sale **403**, faltó un `.Public()`: el contrato es **privado por defecto**. Es el fallo
más probable de toda la migración.

**3. Archivos — ida y vuelta de una imagen REAL, byte a byte.**
```bash
curl -X PUT $BASE/api/files/ --data-binary @foto.jpg -H 'Content-Type: image/jpeg'
# → 201 + la clave generada por el servidor, p. ej. "1721...jpg"

curl -o vuelta.jpg $BASE/api/files/<clave-devuelta>
cmp foto.jpg vuelta.jpg && echo "✅ idéntica"   # debe pasar
```
**`cmp` es el criterio, no que el `PUT` devuelva 201.** Si `cmp` falla, el cuerpo binario
sigue corrompiéndose y la Etapa 2 de `goflare` no está bien hecha.
Compruébalo también abriendo la imagen en el navegador con `<img src="/api/files/…">`.

**4. Validación — un archivo ilegítimo se rechaza.**
```bash
echo '<html><script>alert(1)</script></html>' > malicioso.png   # HTML disfrazado de PNG
curl -X PUT $BASE/api/files/ --data-binary @malicioso.png -H 'Content-Type: image/png'
# → 415, y NADA escrito en el bucket
```
El `Content-Type` miente a propósito. Si esto devuelve 201, la validación por bytes mágicos
no está conectada: estarías sirviendo HTML ejecutable desde tu propio dominio.

## Criterios de aceptación

- `go build ./...` **y** `GOOS=js GOARCH=wasm go build ./...` pasan. (Hoy fallan los dos.)
- `gotest` pasa.
- No queda ninguna referencia a `github.com/tinywasm/goflare/router` ni a
  `github.com/tinywasm/goflare/pages` en el repo.
- **La subida usa `goflare/files`, no una copia.** El demo **no** debe importar
  `tinywasm/filetype` ni `tinywasm/unixid`, ni contener un `r.Put("/api/files/"...)` escrito
  a mano: eso significaría que has duplicado la política de seguridad en vez de consumirla.

  ```bash
  grep -rn "filetype\|unixid" --include=*.go .   # → vacío
  grep -rn "files.New\|store.Mount\|\.Mount(" --include=*.go .   # → aquí sí
  ```
- **Las tres APIs, demostradas en el demo desplegado:**
  1. **D1** — el formulario de contacto guarda y lista registros.
  2. **Router** — esas rutas están servidas por `tinywasm/router` sobre `goflare/edge`, y
     responden **200, no 403** (los `.Public()` están puestos).
  3. **Archivos** — se sube una **imagen real** y se recupera **byte a byte idéntica**. Que
     abra en el navegador sin corromperse es el criterio, no que el `PUT` devuelva 200.
- El despliegue a Cloudflare sigue funcionando: el modo inferido es `pages-functions` (lo
  dispara el import de `goflare/edge`), no `workers`.
