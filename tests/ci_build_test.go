//go:build integration

package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tinywasm/goflare-demo/workflow"
)

// TestCIBuild_Docker replicates the GitHub Actions deploy workflow inside a
// fresh Docker container so you can catch goflare build failures locally
// without waiting for a CI push.
//
// Run with:
//
//	go test -tags=integration -run TestCIBuild_Docker ./tests/ -v
//
// Skipped automatically when Docker is not available.
// The install method and image are defined in internal/workflow/spec.go —
// the same spec that generates .github/workflows/deploy.yml.
func TestCIBuild_Docker(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Docker Linux containers not supported natively on Windows")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not in PATH")
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("docker daemon not running")
	}

	_, file, _, _ := runtime.Caller(0)
	projectRoot, _ := filepath.Abs(filepath.Join(filepath.Dir(file), ".."))

	version, err := workflow.ReadGoflareVersion(filepath.Join(projectRoot, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("testing goflare %s in Docker (%s)", version, workflow.DockerImage)

	steps := append(
		[]string{"set -e"},
		append(workflow.InstallScript(version), "goflare build", "echo BUILD_OK")...,
	)
	script := strings.Join(steps, " && ")

	cmd := exec.Command("docker", "run", "--rm",
		"-v", projectRoot+":/workspace",
		"-w", "/workspace",
		workflow.DockerImage,
		"bash", "-c", script,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	t.Logf("script: %s", script)

	if err := cmd.Run(); err != nil {
		t.Fatalf("goflare build failed in Docker: %v", err)
	}

	for _, artifact := range []string{
		"functions/edge.wasm",
		"functions/[[path]].mjs",
		"web/public/client.wasm",
	} {
		if _, err := os.Stat(filepath.Join(projectRoot, artifact)); err != nil {
			t.Errorf("expected artifact missing: %s", artifact)
		}
	}
}
