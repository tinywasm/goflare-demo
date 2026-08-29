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
      - uses: actions/checkout@v5

      - name: Setup Go
        uses: actions/setup-go@v6
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

      # Option C: goflare builds, wrangler deploys. wrangler owns the Cloudflare
      # Direct Upload protocol (assets hashing, _worker.bundle, _routes.json, bindings).
      - name: Deploy (wrangler)
        env:
          CLOUDFLARE_API_TOKEN: ${{ "{{" }} secrets.CLOUDFLARE_API_TOKEN {{ "}}" }}
          CLOUDFLARE_ACCOUNT_ID: ${{ "{{" }} secrets.CLOUDFLARE_ACCOUNT_ID {{ "}}" }}
        run: |
          # wrangler.toml is the source of truth for the Pages project config.
          # database_id is injected from a GitHub Variable (not committed).
          cat > wrangler.toml <<'EOF'
          name = "{{.ProjectName}}"
          pages_build_output_dir = "{{.PublicDir}}"
          compatibility_date = "{{.CompatibilityDate}}"

          [[d1_databases]]
          binding = "{{.D1Binding}}"
          database_name = "{{.D1DatabaseName}}"
          database_id = "${{ "{{" }} vars.D1_DATABASE_ID {{ "}}" }}"

          # A bucket has no id: the name is the reference. r2.NewEdge("{{.R2Binding}}")
          # fails loudly without this, and main() returns before serving anything.
          [[r2_buckets]]
          binding = "{{.R2Binding}}"
          bucket_name = "{{.R2BucketName}}"
          EOF
          # Scope the catch-all function to API routes; serve everything else statically.
          echo '{{.APIRoutes}}' > "{{.PublicDir}}/_routes.json"
          npx --yes wrangler@{{.WranglerVersion}} pages deploy --branch=main

  e2e:
    needs: deploy
    runs-on: ubuntu-latest
    env:
      FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true
      DEMO_URL: {{.DemoURL}}
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v6
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

      # The point of this repo: an image must survive the round trip byte for byte.
      # cmp is the criterion, not the 201 — a corrupted body still returns 201.
      - name: E2E — Files round trip (upload, fetch back, compare)
        run: |
          printf '\x89PNG\r\n\x1a\n' > foto.png
          head -c 4096 /dev/urandom >> foto.png

          KEY=$(curl -sf -X PUT "$DEMO_URL{{.FilesPrefix}}" \
            --data-binary @foto.png -H 'Content-Type: image/png')
          echo "key: $KEY"
          [ -n "$KEY" ] || (echo "upload returned no key" && exit 1)

          curl -sf -o vuelta.png "$DEMO_URL{{.FilesPrefix}}$KEY"
          cmp foto.png vuelta.png || (echo "the body got corrupted in transit" && exit 1)
          echo "identica"

      # The Content-Type lies on purpose. A 201 here means the magic-byte check is
      # not wired: we would be serving executable HTML from our own origin.
      - name: E2E — Files reject a disguised upload
        run: |
          echo '<html><script>alert(1)</script></html>' > malicioso.png
          STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "$DEMO_URL{{.FilesPrefix}}" \
            --data-binary @malicioso.png -H 'Content-Type: image/png')
          [ "$STATUS" = "415" ] || (echo "Expected 415, got $STATUS" && exit 1)
`

func main() {
	root := findRoot()

	version, err := workflow.ReadGoflareVersion(filepath.Join(root, "go.mod"))
	must(err)

	goVersion := readGoVersion(filepath.Join(root, "go.mod"))

	lines := workflow.InstallScript(version)

	data := map[string]any{
		"GoVersion":         goVersion,
		"InstallLines":      lines,
		"ProjectName":       workflow.ProjectName,
		"PublicDir":         workflow.PublicDir,
		"D1Binding":         workflow.D1Binding,
		"D1DatabaseName":    workflow.D1DatabaseName,
		"R2Binding":         workflow.R2Binding,
		"R2BucketName":      workflow.R2BucketName,
		"FilesPrefix":       workflow.FilesPrefix,
		"CompatibilityDate": workflow.CompatibilityDate,
		"WranglerVersion":   workflow.WranglerVersion,
		"DemoURL":           workflow.DemoURL,
		"APIRoutes":         workflow.APIRoutes,
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
