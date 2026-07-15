---
message: "feat: the demo proves four APIs — router, D1, files and real Google authentication"
---

> Este plan se despacha vía el flujo CodeJob. Ver skill: agents-workflow.
> Orquestado por `tinywasm/docs/DEMO_FOUR_APIS_MASTER_PLAN.md` — **Fase D (aceptación)**.
> **Requiere publicadas las Fases E y F**: `user` sin `golang.org/x/oauth2` (o no entra en el
> Worker), y `goflare/files` con `PerOwner()`.

# Plan — el demo demuestra las CUATRO APIs

Autocontenido, en español. Es la **prueba de aceptación** del ecosistema: si las cuatro no
funcionan aquí, en Cloudflare de verdad, no funcionan.

## La lógica del demo

1. El visitante llega y ve el formulario de **registro**.
2. Para subir una imagen hay que **estar autenticado**. Si no lo está, se le pide **login real
   con Google** (OAuth).
3. Un usuario autenticado tiene **un** archivo. Si sube otro, **reemplaza** al anterior.

Cada paso ejercita una API, y por eso el demo se llama así:

| API | Dónde se demuestra |
|---|---|
| **Router** | las rutas responden 200 y no 403; los `.Public()` están puestos |
| **D1** | el registro guarda y lista usuarios |
| **Archivos** | la imagen sube a R2, vuelve byte a byte, y la segunda sustituye a la primera |
| **Autenticación** | login con Google; sin identidad, la subida es **403** |

## Estado de partida

El árbol trae ya, de la fase anterior: `goflare/edge` con `Config{Authn, Authorize}`, las tres
rutas explícitas de `/api/contacto` (que ya no hacen panic en el servidor nativo), el binding
R2 declarado en `workflow/spec.go`, y los logs `demo:`.

Trae también **`edge/access.go`, que este plan BORRA**. Era una política de juguete: un token
en una cabecera. La sustituye autenticación de verdad. **No la conserves "por si acaso"**: dos
formas de identificarse es una de más, y la débil es la que usará un atacante.

## Cambios

### 1. Subir dependencias y montar el módulo de `user`

`user` (Fase E) y `goflare` (Fase F) a sus versiones publicadas. Verifica **con TinyGo**, que
es el compilador que decide en el borde — `go build` da verde sobre código que TinyGo rechaza,
y así es como `user` llegó a declararse edge-ready sin serlo:

```bash
tinygo build -target=wasm -o /dev/null ./edge/
```

### 2. ⭐ El cableado: `user` ya trae los dos asientos que `edge` pide

No hay que escribir ningún puente. Encajan directamente, y que encajen sin adaptador es la
señal de que el contrato está bien puesto:

```go
import (
    "github.com/tinywasm/goflare/edge"
    userserver "github.com/tinywasm/user/server"
    "github.com/tinywasm/user"
)

auth, err := userserver.New(db, user.Config{
    OAuthProviders: []user.OAuthProvider{&userserver.GoogleProvider{}},
    // el resto de la config: cookie de sesión, TTL…
})
if err != nil { /* no lo silencies */ }

r := edge.NewRouter(edge.Config{
    Authn:     auth.Authenticate(),  // router.Middleware — QUIÉN llama
    Authorize: auth.Can,             // model.Authorizer  — si PUEDE
})

auth.MountAPI(r)   // /login, /logout, /oauth/google, /oauth/callback/google
routes.Register(r, db)
```

`auth.Can` **es** un `model.Authorizer` (misma firma). No lo envuelvas.

### 3. El permiso de subida: la política es del demo

`goflare/files` monta la subida como `Requires(files.Resource, files.Action)` — `files`/Create.
Un usuario recién registrado **no tiene ese permiso por defecto**: el RBAC de `user` es cerrado
por defecto, y eso está bien.

El demo debe **concederlo al registrarse**: al crear el usuario, dale el permiso
`files`/Create. Usa la API de RBAC de `user` (`server/rbac.go`) — **no escribas inserts
directos a la tabla de permisos**: saltarte la librería salta también sus invariantes y sus
eventos de seguridad.

Esa concesión ES la política del demo, y es lo que se está demostrando: la librería pregunta,
la app responde.

### 4. La subida: un archivo por usuario

```go
store, err := files.New(bucket, filesPrefix)
if err != nil { /* ... */ }
store.PerOwner().Mount(r)   // la clave es el id del usuario: subir otra vez REEMPLAZA
```

`PerOwner()` (Fase F) hace que la clave sea la identidad del que sube. No la construyas tú, y
**no mandes el nombre del archivo**: el servidor lo ignora a propósito.

### 5. Frontend

- El formulario de contacto pasa a ser el de **registro**.
- Botón **"Entrar con Google"** → `/oauth/google`.
- La sección de subida solo se muestra a un usuario autenticado. Si un anónimo intenta subir,
  la respuesta es **403** — y el demo debe **mostrarlo**, no esconderlo: es la demostración.
- La imagen de vuelta se pinta con `<img src="/api/files/{clave}">`. La clave es el id del
  usuario; sin extensión, y no hace falta: el `Content-Type` viaja en la metadata de R2.
- La subida manda **los bytes crudos** con `tinywasm/fetch` (`PUT`). Nada de `multipart`.

### 6. Los logs de seguimiento se quedan

`goflare` ya emite todo 4xx/5xx con su causa. Lo que añade el demo son los de la petición que
**va bien** (`demo: ...`). Añade uno para la autenticación:

```go
fmt.Println("demo: usuario autenticado", ctx.UserID())
fmt.Println("demo: upload recibido", ctx.GetHeader("Content-Length"), "bytes")
```

**Nunca imprimas cuerpos, cabeceras, cookies ni tokens.** Tamaños, claves e ids.

## ✅ Verificación REAL

Los tests demuestran que nuestro Go es correcto. **No demuestran que Cloudflare se comporte
como creemos.** Primero Nivel A (`wrangler dev --local`: D1 y R2 emulados de verdad); solo con
A en verde, Nivel B (desplegado).

**1. Router — 200, no 403.** Si sale 403 en una ruta pública, faltó un `.Public()`.

**2. D1 — el registro guarda y lista.**

**3. Autenticación — sin identidad no se sube.**
```bash
curl -X PUT $BASE/api/files/ --data-binary @foto.jpg    # → 403
```
**Este 403 es un criterio de aceptación, no un fallo.** Si devuelve 201, la subida está abierta
a internet y el bucket es de escritura.

**4. Archivos — ida y vuelta byte a byte, con sesión real.**
```bash
# con la cookie de sesión de un usuario logueado:
curl -X PUT $BASE/api/files/ -b "$COOKIE" --data-binary @foto.jpg
# → 201 + la clave (el id del usuario)

curl -o vuelta.jpg -b "$COOKIE" $BASE/api/files/<clave>
cmp foto.jpg vuelta.jpg && echo "✅ idéntica"
```
**`cmp` es el criterio, no que el PUT devuelva 201**: un cuerpo corrupto también da 201.

**5. Reemplazo — un archivo por usuario.**
```bash
curl -X PUT $BASE/api/files/ -b "$COOKIE" --data-binary @otra.jpg
curl -o vuelta2.jpg $BASE/api/files/<misma-clave>
cmp otra.jpg vuelta2.jpg && echo "✅ reemplazada"   # y NO hay un segundo objeto
```

**6. Validación — un archivo ilegítimo se rechaza.**
```bash
echo '<html><script>alert(1)</script></html>' > malicioso.png
curl -X PUT $BASE/api/files/ -b "$COOKIE" --data-binary @malicioso.png
# → 415, y NADA escrito en el bucket
```

Estas comprobaciones van al job `e2e` de CI, que se genera desde
[workflow/spec.go](../workflow/spec.go) — **edita el spec y regenera, nunca el `deploy.yml` a
mano**.

## Configuración manual en Cloudflare (no la hace el agente)

- Secretos del Worker: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`.
- En Google Cloud, registrar el callback: `https://goflare-demo.pages.dev/oauth/callback/google`.

## Criterios de aceptación

- **`tinygo build -target=wasm -o /dev/null ./edge/` pasa.** Es el único que prueba algo sobre
  el borde.
- `go build ./...` y `gotest` pasan.
- La política de juguete ha desaparecido:
  ```bash
  ls edge/access.go                              # → no existe
  grep -rn "X-Demo-Token\|DEMO_UPLOAD_TOKEN" .   # → vacío
  ```
- La subida usa `goflare/files`, no una copia:
  ```bash
  grep -rn "filetype\|unixid" --include=*.go .        # → vacío
  grep -rn "PerOwner()\|store.Mount" --include=*.go . # → aquí sí
  ```
- No hay inserts directos a las tablas de `user`:
  ```bash
  grep -rn "permissions\|roles" --include=*.go . | grep -i "insert\|Create(" # → vacío
  ```
- **Las CUATRO APIs, demostradas en el demo DESPLEGADO:**
  1. **Router** — 200, no 403.
  2. **D1** — el registro guarda y lista.
  3. **Autenticación** — login con Google funciona; **sin identidad, la subida es 403**.
  4. **Archivos** — una imagen real sube, vuelve **byte a byte idéntica**, la segunda
     **reemplaza** a la primera, y un HTML disfrazado es **415**.
