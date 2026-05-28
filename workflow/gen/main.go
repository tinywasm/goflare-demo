// gen writes .github/workflows/deploy.yml from the spec in internal/workflow.
// Run via: go generate ./internal/workflow/
package main

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/tinywasm/goflare-demo/workflow"
)

const deployYML = `name: Deploy to Cloudflare Pages
on:
  push:
    branches: [main]
  workflow_dispatch:

concurrency:
  group: deploy-${{ "{{" }} github.ref {{ "}}" }}
  cancel-in-progress: true

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '{{.GoVersion}}'

      - name: Install goflare
        # Pre-built binary from GitHub Releases (~2-5s vs ~30-90s for go install).
        # Version read from go.mod — no hardcoding here.
        run: |
{{- range .InstallLines}}
          {{.}}
{{- end}}

      - name: Build
        run: goflare build

      - name: Deploy
        env:
          CLOUDFLARE_API_TOKEN: ${{ "{{" }} secrets.CLOUDFLARE_API_TOKEN {{ "}}" }}
          CLOUDFLARE_ACCOUNT_ID: ${{ "{{" }} secrets.CLOUDFLARE_ACCOUNT_ID {{ "}}" }}
          PROJECT_NAME: {{.ProjectName}}
        run: goflare deploy

  e2e:
    needs: deploy
    runs-on: ubuntu-latest
    env:
      FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true
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
            -d '{"nombre":"CI Test","email":"ci@goflare-demo.test","mensaje":"Automated e2e test submission from CI pipeline"}' || true)
          cat /tmp/resp.json
          [ "$STATUS" = "200" ] || (echo "Expected 200, got $STATUS" && exit 1)

      - name: E2E — Verify D1 record
        run: go test -tags=integration -run TestE2E_ContactSubmission ./tests/e2e/ -v
`

func main() {
	root := findRoot()

	version, err := workflow.ReadGoflareVersion(filepath.Join(root, "go.mod"))
	must(err)

	goVersion := readGoVersion(filepath.Join(root, "go.mod"))

	lines := workflow.InstallScript(version)

	data := map[string]any{
		"GoVersion":    goVersion,
		"InstallLines": lines,
		"ProjectName":  workflow.ProjectName,
	}

	tmpl := template.Must(template.New("").Parse(deployYML))

	out := filepath.Join(root, ".github", "workflows", "deploy.yml")
	f, err := os.Create(out)
	must(err)
	defer f.Close()
	must(tmpl.Execute(f, data))
}

func findRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("go.mod not found")
		}
		dir = parent
	}
}

func readGoVersion(gomod string) string {
	data, _ := os.ReadFile(gomod)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go ") {
			return strings.TrimPrefix(line, "go ")
		}
	}
	return "1.22"
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
