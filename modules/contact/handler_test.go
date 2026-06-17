//go:build !wasm

package contact

import (
	"testing"

	"github.com/tinywasm/sqlite"
)

// TestContactCreate_Local reproduces the POST /api/contacto path locally against
// in-memory SQLite: the exact NewContact + db.Create sequence the handler runs.
// Mirrors the cloud failure (HTTP 500 / code 1101, empty table) to localize the bug.
func TestContactCreate_Local(t *testing.T) {
	db, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatalf("sqlite open: %v", err)
	}
	defer db.Close()

	if err := db.CreateTable(&Contact{}); err != nil {
		t.Fatalf("create table: %v", err)
	}

	body := []byte(`{"nombre":"CI Test","email":"ci@goflare-demo.test","mensaje":"Automated e2e test submission from CI pipeline"}`)

	sub, err := NewContact(body)
	if err != nil {
		t.Fatalf("NewContact: %v", err)
	}

	if err := db.Create(sub); err != nil {
		t.Fatalf("db.Create: %v", err)
	}

	list, err := ReadAllContact(db.Query(&Contact{}).OrderBy("id").Desc())
	if err != nil {
		t.Fatalf("read all: %v", err)
	}
	if list.Len() != 1 {
		t.Fatalf("expected 1 row persisted, got %d", list.Len())
	}
	if got := (*list)[0]; got.Email != "ci@goflare-demo.test" || got.Nombre != "CI Test" {
		t.Fatalf("unexpected row: %+v", got)
	}
}
