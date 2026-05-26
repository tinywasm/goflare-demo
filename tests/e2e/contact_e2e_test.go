//go:build integration && !wasm

package e2e_test

import (
	"os"
	"testing"

	"github.com/tinywasm/fmt"
	"github.com/tinywasm/goflare/d1"
)

// contactRow refleja la tabla contact_submission. Schema usa fmt.Field (no orm.Field).
type contactRow struct {
	ID      int
	Nombre  string
	Email   string
	Mensaje string
}

func (m *contactRow) ModelName() string { return "contact_submission" } // = ormc ModelName
func (m *contactRow) Schema() []fmt.Field {
	return []fmt.Field{
		{Name: "id", DB: &fmt.FieldDB{PK: true, AutoInc: true}},
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

	// Query builder real: db.Query(m).Where(col).Eq(v).OrderBy(col).Desc().ReadOne()
	row := &contactRow{}
	err = db.Query(row).Where("email").Eq("ci@goflare-demo.test").OrderBy("id").Desc().ReadOne()
	if err != nil {
		t.Fatalf("CI submission not found in D1: %v", err) // orm.ErrNotFound si no existe
	}
	if row.Nombre != "CI Test" {
		t.Errorf("expected Nombre=CI Test, got %q", row.Nombre)
	}
	t.Logf("submission ID=%d persisted in D1", row.ID)
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
