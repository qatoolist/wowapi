package compat

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCandidateArchitectureSmokeRequiresDigestAndTargetsPlatform(t *testing.T) {
	dir := t.TempDir()
	dockerLog := filepath.Join(dir, "docker.log")
	fakeDocker := filepath.Join(dir, "docker")
	body := "#!/bin/sh\nprintf '%s\\n' \"$*\" > \"$DOCKER_LOG\"\necho 'wowapi v1.2.3'\n"
	if err := os.WriteFile(fakeDocker, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join("..", "..", "scripts", "smoke_candidate_arch.sh")
	digest := "ghcr.io/qatoolist/wowapi@sha256:" + strings.Repeat("a", 64)

	cmd := exec.Command("sh", script, digest, "arm64")
	cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "DOCKER_LOG="+dockerLog)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("smoke failed: %v: %s", err, output)
	}
	logged, err := os.ReadFile(dockerLog)
	if err != nil {
		t.Fatal(err)
	}
	want := "run --rm --platform linux/arm64 " + digest + " version"
	if strings.TrimSpace(string(logged)) != want {
		t.Fatalf("docker invocation = %q, want %q", strings.TrimSpace(string(logged)), want)
	}

	cmd = exec.Command("sh", script, "ghcr.io/qatoolist/wowapi:latest", "amd64")
	cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "DOCKER_LOG="+dockerLog)
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "immutable digest") {
		t.Fatalf("mutable tag must fail loudly, err=%v output=%s", err, output)
	}
}

func TestCandidateOCISmokeCopiesExactLayoutAndRunsEveryPlatform(t *testing.T) {
	dir := t.TempDir()
	toolLog := filepath.Join(dir, "tools.log")
	tools := map[string]string{
		"docker": `#!/bin/sh
printf 'docker %s\n' "$*" >> "$TOOL_LOG"
case "$*" in
  "run --detach"*) echo registry-container ;;
  "run --rm --platform"*) echo "wowapi v1.2.3" ;;
esac
`,
		"oras": `#!/bin/sh
printf 'oras %s\n' "$*" >> "$TOOL_LOG"
`,
		"curl": "#!/bin/sh\nexit 0\n",
		"tar":  "#!/bin/sh\nexit 0\n",
	}
	for name, body := range tools {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	archive := filepath.Join(dir, "candidate.oci.tar")
	if err := os.WriteFile(archive, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	digest := "sha256:" + strings.Repeat("b", 64)
	script := filepath.Join("..", "..", "scripts", "smoke_candidate_oci.sh")
	cmd := exec.Command("sh", script, archive, "v1.2.3", digest)
	cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "TOOL_LOG="+toolLog)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("OCI candidate smoke failed: %v: %s", err, output)
	}
	loggedBytes, err := os.ReadFile(toolLog)
	if err != nil {
		t.Fatal(err)
	}
	logged := string(loggedBytes)
	for _, want := range []string{
		"registry:2@sha256:a3d8aaa63ed8681a604f1dea0aa03f100d5895b6a58ace528858a7b332415373",
		"oras cp --to-plain-http --from-oci-layout ",
		":v1.2.3 127.0.0.1:5000/wowapi:v1.2.3",
		"docker run --rm --platform linux/amd64 127.0.0.1:5000/wowapi@" + digest + " version",
		"docker run --rm --platform linux/arm64 127.0.0.1:5000/wowapi@" + digest + " version",
		"docker rm -f wowapi-candidate-registry-",
	} {
		if !strings.Contains(logged, want) {
			t.Fatalf("tool log missing %q:\n%s", want, logged)
		}
	}
}
