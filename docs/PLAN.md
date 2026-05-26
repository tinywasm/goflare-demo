> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan: goflare-demo — E2E real: compilar + deploy + D1 persistente

## Objetivo

Convertir goflare-demo en la **prueba de integración pública y permanente de goflare**:

1. El CI compila `edge/main.go` con TinyGo → prueba que el compilador de goflare funciona.
2. Despliega en Cloudflare Pages → prueba el pipeline de deploy.
3. Hace un POST real al handler de contacto → el registro se guarda en D1.
4. Lee el registro de vuelta vía D1 REST → afirma el ciclo completo.
5. Los registros **persisten en la DB** — el demo vivo en `goflare-demo.tinywasm.app`
   muestra submissions reales, actuando como evidencia pública de que goflare funciona.

### Por qué en goflare-demo y no en goflare

goflare prueba que la API Go es correcta (tests unitarios); goflare-demo prueba que
*funciona en producción* con una app real. Separar los repos evita dependencias
circulares y que un error en el demo bloquee el CI del compilador.

### Por qué los registros son permanentes

Un demo que muestra datos reales generados por el propio CI es más valioso que un
test sintético. Cada corrida del pipeline agrega una submission visible para cualquier
visitante — ciclo de vida completo demostrado en producción.

---

## Contexto técnico

- **Modo**: Pages Functions (`edge/main.go` importa `goflare/pages`).
- **Handler actual**: `POST /api/contacto` parsea el form, valida, retorna 200 pero
  **no persiste nada en D1** (stub). El bloque de `sendEmail` está comentado.
- **`ContactForm`**: solo modelo de formulario (`ormc:formonly`). Sin ID ni persistencia.
- **`d1` package**: solo compila con `//go:build wasm`. El handler es build-agnostic →
  el acceso a D1 debe ir en un archivo `db_wasm.go` dentro del módulo `contact`.
- **Binding D1**: configurado en el dashboard de Cloudflare Pages → Settings →
  Functions → D1 database bindings → nombre del binding: `DB`.

---

## Stages

### Stage 1 — Modelo `ContactSubmission` con persistencia D1

Crear `modules/contact/submission.go` (sin build tag — build-agnostic):

```go
package contact

import "github.com/tinywasm/fmt"

// ContactSubmission es el registro persistido en D1.
type ContactSubmission struct {
	ID      int64
	Nombre  string
	Email   string
	Mensaje string
}

func (m *ContactSubmission) ModelName() string { return "contact_submissions" }

func (m *ContactSubmission) Schema() []fmt.Field {
	return []fmt.Field{
		{Name: "id", PK: true},
		{Name: "nombre"},
		{Name: "email"},
		{Name: "mensaje"},
	}
}

func (m *ContactSubmission) Pointers() []any {
	return []any{&m.ID, &m.Nombre, &m.Email, &m.Mensaje}
}
```

### Stage 2 — Acceso D1: `db_wasm.go` y stub `db_host.go`

**`modules/contact/db_wasm.go`** (`//go:build wasm`):

```go
//go:build wasm

package contact

import "github.com/tinywasm/goflare/d1"

func saveSubmission(sub *ContactSubmission) error {
	db, err := d1.New("DB")
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.CreateTable(sub); err != nil {
		return err
	}
	return db.Create(sub)
}

func listSubmissions() ([]*ContactSubmission, error) {
	db, err := d1.New("DB")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var rows []*ContactSubmission
	if err := db.All(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}
```

**`modules/contact/db_host.go`** (`//go:build !wasm`):

```go
//go:build !wasm

package contact

import "errors"

var errHostOnly = errors.New("d1 only available in wasm")

func saveSubmission(_ *ContactSubmission) error         { return errHostOnly }
func listSubmissions() ([]*ContactSubmission, error) { return nil, errHostOnly }
```

### Stage 3 — Handler POST: persistir en D1

Editar `modules/contact/handler.go`: tras `data.Validate`, llamar `saveSubmission`:

```go
sub := &ContactSubmission{Nombre: data.Nombre, Email: data.Email, Mensaje: data.Mensaje}
if err := saveSubmission(sub); err != nil {
	ctx.WriteStatus(502)
	ctx.Write([]byte(`{"error":"db error"}`))
	return
}
ctx.WriteStatus(200)
ctx.Write([]byte(`{"message":"¡Gracias! Hemos recibido tu solicitud."}`))
```

### Stage 4 — Ruta GET: listar submissions

**`modules/contact/list_handler.go`** (nuevo, build-agnostic):

```go
package contact

import (
	"github.com/tinywasm/goflare/router"
	"github.com/tinywasm/json"
)

func HandleList(ctx router.Context) {
	ctx.SetHeader("Content-Type", "application/json")
	ctx.SetHeader("Access-Control-Allow-Origin", "*")

	rows, err := listSubmissions()
	if err != nil {
		ctx.WriteStatus(502)
		ctx.Write([]byte(`{"error":"db error"}`))
		return
	}
	body, err := json.Encode(rows)
	if err != nil {
		ctx.WriteStatus(500)
		ctx.Write([]byte(`{"error":"encode error"}`))
		return
	}
	ctx.WriteStatus(200)
	ctx.Write(body)
}
```

Registrar en `routes/routes.go`:

```go
r.Get("/api/contacto", contact.HandleList)
```

### Stage 5 — Frontend: mostrar submissions

`web/client.go`: al cargar la página hacer `fetch("/api/contacto")` y renderizar la
lista debajo del formulario. Mostrar Nombre, email parcialmente oculto
(`ci***@test.com`), y los primeros 60 chars del Mensaje. Incluir contador:
"N solicitudes recibidas".

### Stage 6 — D1 binding en Cloudflare Pages (manual, una vez)

```
Cloudflare Pages → goflare-demo → Settings → Functions → D1 database bindings
  Variable name: DB
  D1 database:   [la DB ya creada]
```

Documentar en `docs/CI_D1_SETUP.md` con captura de pantalla.

### Stage 7 — E2E job en `deploy.yml`

Añadir job `e2e` con `needs: deploy`:

```yaml
  e2e:
    needs: deploy
    runs-on: ubuntu-latest
    env:
      FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true
      CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
      CLOUDFLARE_ACCOUNT_ID: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
      D1_DATABASE_ID: ${{ vars.D1_DATABASE_ID }}
      DEMO_URL: https://goflare-demo.tinywasm.app
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: 'stable'

      - name: Wait for Pages deployment
        run: sleep 30

      - name: E2E — POST contact form
        run: |
          STATUS=$(curl -s -o /tmp/resp.json -w "%{http_code}" \
            -X POST "$DEMO_URL/api/contacto" \
            -H "Content-Type: application/json" \
            -d '{"nombre":"CI Test","email":"ci@goflare-demo.test","mensaje":"Automated e2e test submission from CI pipeline"}')
          cat /tmp/resp.json
          [ "$STATUS" = "200" ] || (echo "Expected 200, got $STATUS" && exit 1)

      - name: E2E — Verify D1 record
        run: go test -tags=integration -run TestE2E_ContactSubmission ./tests/e2e/ -v
```

### Stage 8 — Test `tests/e2e/contact_e2e_test.go`

```go
//go:build integration && !wasm

package e2e_test

import (
	"os"
	"testing"

	"github.com/tinywasm/goflare/d1"
	"github.com/tinywasm/orm"
)

type contactRow struct {
	ID      int64
	Nombre  string
	Email   string
	Mensaje string
}

func (m *contactRow) ModelName() string { return "contact_submissions" }
func (m *contactRow) Schema() []orm.Field {
	return []orm.Field{
		{Name: "id", PK: true},
		{Name: "nombre"},
		{Name: "email"},
		{Name: "mensaje"},
	}
}
func (m *contactRow) Pointers() []any { return []any{&m.ID, &m.Nombre, &m.Email, &m.Mensaje} }

func TestE2E_ContactSubmission(t *testing.T) {
	token     := requireEnv(t, "CLOUDFLARE_API_TOKEN")
	accountID := requireEnv(t, "CLOUDFLARE_ACCOUNT_ID")
	dbID      := requireEnv(t, "D1_DATABASE_ID")

	db, err := d1.NewDirect(token, accountID, dbID)
	if err != nil {
		t.Fatalf("NewDirect: %v", err)
	}
	defer db.Close()

	row := &contactRow{}
	if err := db.First(row, orm.Where("email", "ci@goflare-demo.test"), orm.OrderBy("id", "DESC")); err != nil {
		t.Fatalf("CI submission not found in D1: %v", err)
	}
	if row.Nombre != "CI Test" {
		t.Errorf("expected Nombre=CI Test, got %q", row.Nombre)
	}
	t.Logf("✅ submission ID=%d persisted in D1", row.ID)
	// Sin cleanup — los registros persisten para el demo vivo
}

func requireEnv(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("env var %s not set", key)
	}
	return v
}
```

---

## Resumen de archivos

| Archivo | Acción |
|---|---|
| `modules/contact/submission.go` | Nuevo — `ContactSubmission` con ORM inline |
| `modules/contact/db_wasm.go` | Nuevo — `saveSubmission` + `listSubmissions` (d1.New) |
| `modules/contact/db_host.go` | Nuevo — stubs `!wasm` para compilación host |
| `modules/contact/handler.go` | Editar — llamar `saveSubmission` tras validar |
| `modules/contact/list_handler.go` | Nuevo — `HandleList` GET /api/contacto |
| `routes/routes.go` | Editar — añadir `r.Get("/api/contacto", contact.HandleList)` |
| `web/client.go` | Editar — fetch + render lista de submissions |
| `.github/workflows/deploy.yml` | Editar — añadir job `e2e` |
| `tests/e2e/contact_e2e_test.go` | Nuevo — `TestE2E_ContactSubmission` |
| `docs/CI_D1_SETUP.md` | Nuevo — instrucciones binding D1 en Pages dashboard |

---

## Verification

```bash
go build ./...
go vet ./...
```

- `go build` en host compila sin errores (stubs `!wasm` cubren el build-agnostic).
- CI deploy job produce `functions/edge.wasm` + `functions/[[path]].mjs`.
- E2E job: curl POST → 200; `TestE2E_ContactSubmission` → registro en D1 ✅.
- `goflare-demo.tinywasm.app` muestra la lista de submissions con los registros del CI.
