//go:build integration

package tests

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// TestCIBuild_Docker replicates the GitHub Actions deploy workflow inside a
// fresh Docker container (golang:latest on Linux/amd64) so you can catch
// goflare build failures locally without waiting for a CI push.
//
// Run with:
//
//	go test -tags=integration -run TestCIBuild_Docker ./tests/ -v
//
// Skipped automatically when Docker is not available.
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

	// Resolve project root (two levels up from tests/)
	_, file, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(file), "..")
	projectRoot, _ = filepath.Abs(projectRoot)

	// Read the goflare version from go.mod so the test stays in sync automatically
	goflareVersion := readGoflareVersion(t, filepath.Join(projectRoot, "go.mod"))
	t.Logf("testing goflare %s in Docker", goflareVersion)

	goVersion := readGoVersion(t, filepath.Join(projectRoot, "go.mod"))

	image := "golang:" + goVersion + "-bookworm"

	script := strings.Join([]string{
		"set -e",
		"go install github.com/tinywasm/goflare/cmd/goflare@" + goflareVersion,
		"goflare build",
		"echo BUILD_OK",
	}, " && ")

	cmd := exec.Command("docker", "run", "--rm",
		"-v", projectRoot+":/workspace",
		"-w", "/workspace",
		image,
		"bash", "-c", script,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	t.Logf("running: %s", strings.Join(cmd.Args, " "))

	if err := cmd.Run(); err != nil {
		t.Fatalf("goflare build failed in Docker: %v", err)
	}

	// Verify artifacts were produced
	for _, artifact := range []string{
		"functions/edge.wasm",
		"functions/[[path]].mjs",
		"web/public/client.wasm",
	} {
		if _, err := os.Stat(filepath.Join(projectRoot, artifact)); err != nil {
			t.Errorf("expected artifact missing after build: %s", artifact)
		}
	}
}

// readGoflareVersion extracts the goflare version from go.mod.
func readGoflareVersion(t *testing.T, gomod string) string {
	t.Helper()
	return readModuleVersion(t, gomod, "github.com/tinywasm/goflare")
}

// readGoVersion extracts the minimum Go version from go.mod.
func readGoVersion(t *testing.T, gomod string) string {
	t.Helper()
	f, err := os.Open(gomod)
	if err != nil {
		t.Fatalf("cannot open go.mod: %v", err)
	}
	defer f.Close()
	re := regexp.MustCompile(`^go\s+(\d+\.\d+)`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := re.FindStringSubmatch(scanner.Text()); m != nil {
			return m[1]
		}
	}
	return "1.22" // safe fallback
}

func readModuleVersion(t *testing.T, gomod, module string) string {
	t.Helper()
	f, err := os.Open(gomod)
	if err != nil {
		t.Fatalf("cannot open go.mod: %v", err)
	}
	defer f.Close()
	re := regexp.MustCompile(`^\s*` + regexp.QuoteMeta(module) + `\s+(v\S+)`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := re.FindStringSubmatch(scanner.Text()); m != nil {
			return m[1]
		}
	}
	t.Fatalf("module %q not found in %s", module, gomod)
	return ""
}
