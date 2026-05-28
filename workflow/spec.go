//go:generate go run gen/main.go

// Package workflow is the single source of truth for the CI/CD pipeline.
// Edit this file to change how the project is built and deployed.
// Then run: go generate ./internal/workflow/
// to regenerate .github/workflows/deploy.yml.
package workflow

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

const (
	// GoflareModule is the module path used to find the version in go.mod.
	GoflareModule = "github.com/tinywasm/goflare"

	// BinaryURLTemplate is the GitHub Releases download URL.
	// {version} is replaced with the version read from go.mod at generate/test time.
	BinaryURLTemplate = "https://github.com/tinywasm/goflare/releases/download/{version}/goflare-linux-amd64"

	// DockerImage is the container image used for local CI simulation.
	// Must have Go so TinyGo can invoke 'go' internally.
	DockerImage = "golang:1.25-bookworm"

	// ProjectName is the Cloudflare Pages project name — used by goflare deploy
	// to construct the API path /accounts/{id}/pages/projects/{name}/deployments.
	// Not a secret; safe to hardcode here.
	ProjectName = "goflare-demo"
)

// InstallScript returns the shell commands to install goflare from a
// pre-built binary. version is e.g. "v0.2.22".
func InstallScript(version string) []string {
	url := fmt.Sprintf(
		"https://github.com/tinywasm/goflare/releases/download/%s/goflare-linux-amd64",
		version,
	)
	return []string{
		"curl -fsSL " + url + " -o /usr/local/bin/goflare",
		"chmod +x /usr/local/bin/goflare",
	}
}

// ReadGoflareVersion reads the goflare version from the given go.mod file.
func ReadGoflareVersion(gomodPath string) (string, error) {
	f, err := os.Open(gomodPath)
	if err != nil {
		return "", fmt.Errorf("open go.mod: %w", err)
	}
	defer f.Close()
	re := regexp.MustCompile(`^\s*` + regexp.QuoteMeta(GoflareModule) + `\s+(v\S+)`)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if m := re.FindStringSubmatch(sc.Text()); m != nil {
			return m[1], nil
		}
	}
	return "", fmt.Errorf("%s not found in go.mod", GoflareModule)
}
