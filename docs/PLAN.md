> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan: goflare-demo — Persist contact form submissions to Cloudflare D1

## Context

`tinywasm/goflare-demo` (module `github.com/tinywasm/goflare-demo`) is a demo Cloudflare Worker
built with `github.com/tinywasm/goflare`. It has a contact form (`/api/contacto`) that currently
validates input and returns a 200 response without persisting anything.

This plan wires the contact form handler to save each submission to a Cloudflare D1 database
using `github.com/tinywasm/goflare/d1` (ORM adapter) and `github.com/tinywasm/orm`.

Tests run via `gotest` — no TinyGo installation required.

## TinyWasm Constraints (mandatory)

- No `import "errors"`, `"fmt"`, `"strings"` from stdlib — use `github.com/tinywasm/fmt`.
- Files that use `syscall/js` or D1 (WASM-only APIs) must have `//go:build wasm` as first line.
- No `database/sql`. D1 is accessed exclusively via `github.com/tinywasm/goflare/d1`.

## Current State

| File | Role |
|---|---|
| `modules/contact/model.go` | `ContactForm` struct (form-only, `ormc:formonly`) — no DB fields |
| `modules/contact/model_orm.go` | Auto-generated — Schema + Pointers for ContactForm. **DO NOT EDIT** |
| `modules/contact/handler.go` | Decodes JSON, validates, returns 200. No persistence. |
| `go.mod` | Has `github.com/tinywasm/goflare` but NOT `github.com/tinywasm/orm` |

`ContactForm` is `ormc:formonly` — it has no `ID` or `ModelName()`. It must **not** be modified.
A separate `ContactSubmission` model is introduced for DB persistence.

## Goal

1. Add `ContactSubmission` model implementing `fmt.Model` (with `id` PK).
2. Add `store_wasm.go` — opens D1 and saves a `ContactSubmission` per form submission.
3. Update `handler.go` — call `saveSubmission` after validation, before returning 200.
4. Update `go.mod` — bump `github.com/tinywasm/goflare` to a version that includes `d1/`, add `github.com/tinywasm/orm`.

## D1 Binding Name

The Worker's D1 binding is named **`"DB"`** (configured in `wrangler.toml`). Use the constant
`d1Binding = "DB"` — never a string literal in logic.

## Stages

### Stage 1 — `modules/contact/model.go` (edit) + run `ormc`

`ormc` generates `ModelName()`, `Schema()`, `Pointers()`, and `Validate()` automatically.
**Do NOT write these methods by hand.**

Add `ContactSubmission` to `modules/contact/model.go` with the `// ormc:form` directive
(DB + Form — generates full CRUD boilerplate):

```go
// ormc:form
type ContactSubmission struct {
	ID      int64
	Nombre  string
	Email   string
	Mensaje string
}
```

Then run from the module root:

```bash
ormc
```

`ormc` will generate/update `model_orm.go` with `ModelName()`, `Schema()`, `Pointers()`,
`Validate()` for `ContactSubmission`. The existing `ContactForm` entries in `model_orm.go`
are preserved — `ormc` processes all structs in the file.

### Stage 2 — `modules/contact/store_wasm.go` (new file, `//go:build wasm`)

```go
//go:build wasm

package contact

import "github.com/tinywasm/goflare/d1"

const d1Binding = "DB"

func saveSubmission(form ContactForm) error {
	db, err := d1.New(d1Binding)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.CreateTable(&ContactSubmission{}); err != nil {
		return err
	}

	return db.Create(&ContactSubmission{
		Nombre:  form.Nombre,
		Email:   form.Email,
		Mensaje: form.Mensaje,
	})
}
```

### Stage 3 — `modules/contact/handler.go` (edit)

After the `data.Validate(0)` block and before `ctx.WriteStatus(200)`, add:

```go
if err := saveSubmission(data); err != nil {
    ctx.WriteStatus(502)
    ctx.Write([]byte(`{"error":"` + err.Error() + `"}`))
    return
}
```

The final handler flow must be:
1. Check OPTIONS → 204
2. Check non-POST → 405
3. Decode JSON → 400 on error
4. Validate → 422 on error
5. **saveSubmission → 502 on error** ← new
6. Return 200

Remove the commented-out `sendEmail` block — it is replaced by D1 persistence.

### Stage 4 — `go.mod` (edit)

Run:

```bash
go get github.com/tinywasm/goflare@latest
go get github.com/tinywasm/orm@latest
go mod tidy
```

`github.com/tinywasm/goflare/d1` is a subpackage of `goflare` — no separate module entry needed.
`github.com/tinywasm/orm` must be an explicit direct dependency (used in `db_model.go`).

## Stages Summary

| # | Archivo | Acción |
|---|---|---|
| 1 | `modules/contact/model.go` | Agregar struct `ContactSubmission` con `// ormc:form` + correr `ormc` |
| 2 | `modules/contact/store_wasm.go` | Crear — `saveSubmission` usando `d1.New` + `orm.DB` |
| 3 | `modules/contact/handler.go` | Editar — llamar `saveSubmission` + eliminar bloque comentado |
| 4 | `go.mod` | `go get goflare@latest` + `go get orm@latest` + `go mod tidy` |

## Verification

```bash
gotest
```

Sin regresiones. El módulo compila. Los tests de handler existentes deben seguir pasando.
