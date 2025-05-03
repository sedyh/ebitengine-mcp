package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sedyh/ebitengine-mcp/internal/errs"
)

const (
	DefaultFlag = "mcp"
	DefaultBin  = "ebitengine-mcp-run"
)

func Run(id, target string) (e error) {
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("stat %s: %w", target, err)
	}

	pkg := filepath.Dir(target)
	if info.IsDir() {
		pkg = target
	}

	tmp, err := os.CreateTemp("", fmt.Sprintf("%s-%s-*", DefaultBin, id))
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer errs.Closer(&e, tmp)

	out := tmp.Name()
	defer os.Remove(out)

	gobin, err := Compiler()
	if err != nil {
		return fmt.Errorf("find compiler: %w", err)
	}

	old, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}

	if err := os.Chdir(pkg); err != nil {
		return fmt.Errorf("chdir to %s: %w", pkg, err)
	}

	build := exec.Command(gobin, "build", "-o", out, ".")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr

	if err := build.Run(); err != nil {
		os.Chdir(old)
		return fmt.Errorf("build project: %w", err)
	}

	if err := os.Chdir(old); err != nil {
		return fmt.Errorf("restore dir to %s: %w", old, err)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		return fmt.Errorf("output file not found: %w", err)
	}

	run := exec.Command(out, "-"+DefaultFlag, id)
	run.Stdin = os.Stdin
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	run.Env = os.Environ()

	if err := run.Run(); err != nil {
		var ex *exec.ExitError
		if errors.As(err, &ex) {
			os.Exit(ex.ExitCode())
		}
		return fmt.Errorf("run program: %w", err)
	}

	return nil
}

func Compiler() (string, error) {
	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		return "", fmt.Errorf("env variable GOROOT is not set")
	}
	gobin := filepath.Join(goroot, "bin", "go")
	if _, err := os.Stat(gobin); os.IsNotExist(err) {
		return "", fmt.Errorf("go binary not found at %q", gobin)
	}
	return gobin, nil
}
