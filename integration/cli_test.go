package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIWorkflow(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-independent smoke test currently targets Unix paths")
	}
	root := t.TempDir()
	templateDir := filepath.Join(root, "template")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatal(err)
	}
	manifest := "description: Go starter\ntags: [go, backend]\nvariables:\n  - name: project_name\n    required: true\ntasks:\n  check:\n    command: go\n    args: [version]\n"
	writeFile(t, filepath.Join(templateDir, ".stew.yaml"), manifest)
	writeFile(t, filepath.Join(templateDir, "README.md"), "# {{ .project_name }}\n")

	config := filepath.Join(root, "config.yaml")
	writeFile(t, config, fmt.Sprintf("stewsPath: %q\ntemplatesPath: %q\n", filepath.Join(root, "stews.json"), filepath.Join(root, "cache")))
	binary := buildBinary(t)

	result := runCLI(t, binary, root, config, "save", "go-starter", "--path", templateDir)
	if result.Err != nil {
		t.Fatalf("save failed: %v\n%s", result.Err, result.Output)
	}
	baseDir := filepath.Join(root, "base-template")
	childDir := filepath.Join(root, "child-template")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(baseDir, ".stew.yaml"), "description: Base\nvariables:\n  - name: project_name\n    required: true\n")
	writeFile(t, filepath.Join(baseDir, "base.txt"), "inherited\n")
	writeFile(t, filepath.Join(childDir, ".stew.yaml"), "extends: base-starter\ndescription: Child\n")
	writeFile(t, filepath.Join(childDir, "child.txt"), "child\n")
	if result := runCLI(t, binary, root, config, "save", "base-starter", "--path", baseDir); result.Err != nil {
		t.Fatalf("save base failed: %v\n%s", result.Err, result.Output)
	}
	if result := runCLI(t, binary, root, config, "save", "child-starter", "--path", childDir); result.Err != nil {
		t.Fatalf("save child failed: %v\n%s", result.Err, result.Output)
	}
	derivedDir := filepath.Join(root, "derived")
	if result := runCLI(t, binary, root, config, "create", "child-starter", derivedDir, "--var", "project_name=derived"); result.Err != nil {
		t.Fatalf("create inherited template failed: %v\n%s", result.Err, result.Output)
	}
	if inherited, err := os.ReadFile(filepath.Join(derivedDir, "base.txt")); err != nil || string(inherited) != "inherited\n" {
		t.Fatalf("inherited file = %q, err = %v", inherited, err)
	}

	view := runCLI(t, binary, root, config, "view", "go-starter", "--json")
	if view.Err != nil {
		t.Fatalf("view failed: %v\n%s", view.Err, view.Output)
	}
	var viewJSON struct {
		Manifest struct {
			Description string `json:"description"`
		} `json:"manifest"`
	}
	if err := json.Unmarshal([]byte(view.Output), &viewJSON); err != nil {
		t.Fatalf("view JSON: %v\n%s", err, view.Output)
	}
	if viewJSON.Manifest.Description != "Go starter" {
		t.Fatalf("view manifest = %+v", viewJSON.Manifest)
	}

	projectDir := filepath.Join(root, "project")
	create := runCLI(t, binary, root, config, "create", "go-starter", projectDir, "--var", "project_name=demo", "--json")
	if create.Err != nil {
		t.Fatalf("create failed: %v\n%s", create.Err, create.Output)
	}
	var changes []struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal([]byte(create.Output), &changes); err != nil {
		t.Fatalf("create JSON: %v\n%s", err, create.Output)
	}
	if len(changes) != 2 {
		t.Fatalf("changes = %+v", changes)
	}
	readme, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	if err != nil || string(readme) != "# demo\n" {
		t.Fatalf("README = %q, err = %v", readme, err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".stew.yaml")); err != nil {
		t.Fatalf("generated manifest: %v", err)
	}

	doctor := runCLI(t, binary, projectDir, config, "doctor", "--json")
	if doctor.Err != nil {
		t.Fatalf("doctor failed: %v\n%s", doctor.Err, doctor.Output)
	}
	var diagnostics []struct {
		Level string `json:"level"`
	}
	if err := json.Unmarshal([]byte(doctor.Output), &diagnostics); err != nil {
		t.Fatalf("doctor JSON: %v\n%s", err, doctor.Output)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Level == "ERROR" {
			t.Fatalf("doctor diagnostics = %s", doctor.Output)
		}
	}

	tasks := runCLI(t, binary, projectDir, config, "tasks", "--json")
	if tasks.Err != nil || !strings.Contains(tasks.Output, `"name":"check"`) {
		t.Fatalf("tasks failed: %v\n%s", tasks.Err, tasks.Output)
	}
	run := runCLI(t, binary, projectDir, config, "run", "check", "--yes")
	if run.Err != nil || !strings.Contains(run.Output, "go version") {
		t.Fatalf("run failed: %v\n%s", run.Err, run.Output)
	}

	archive := filepath.Join(root, "go-starter.tar.gz")
	export := runCLI(t, binary, root, config, "export", "go-starter", archive)
	if export.Err != nil {
		t.Fatalf("export failed: %v\n%s", export.Err, export.Output)
	}
	imported := runCLI(t, binary, root, config, "import", archive, "--name", "imported-starter")
	if imported.Err != nil {
		t.Fatalf("import failed: %v\n%s", imported.Err, imported.Output)
	}
}

type cliResult struct {
	Output string
	Err    error
}

func buildBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "stew")
	command := exec.Command("go", "build", "-o", binary, ".")
	command.Dir = repoRoot(t)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	return binary
}

func runCLI(t *testing.T, binary, dir, config string, args ...string) cliResult {
	t.Helper()
	command := exec.Command(binary, append([]string{"--config", config}, args...)...)
	command.Dir = dir
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	err := command.Run()
	return cliResult{Output: output.String(), Err: err}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine integration test location")
	}
	return filepath.Dir(filepath.Dir(file))
}
