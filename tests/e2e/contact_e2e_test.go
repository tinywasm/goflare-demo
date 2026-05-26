//go:build integration && !wasm

package e2e_test

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
)

type contactRow struct {
	ID      int    `json:"id"`
	Nombre  string `json:"nombre"`
	Email   string `json:"email"`
	Mensaje string `json:"mensaje"`
}

func TestE2E_ContactSubmission(t *testing.T) {
	demoURL := requireEnv(t, "DEMO_URL")

	resp, err := http.Get(demoURL + "/api/contacto")
	if err != nil {
		t.Fatalf("Failed to GET /api/contacto: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var submissions []contactRow
	if err := json.Unmarshal(body, &submissions); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v. Body was: %s", err, string(body))
	}

	found := false
	for _, row := range submissions {
		if row.Email == "ci@goflare-demo.test" && row.Nombre == "CI Test" {
			found = true
			t.Logf("Submission ID=%d persisted in D1 found via HTTP", row.ID)
			break
		}
	}

	if !found {
		t.Errorf("CI submission not found in D1 via GET /api/contacto. Submissions retrieved: %s", string(body))
	}
}

func requireEnv(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("env var %s not set", key)
	}
	return v
}
