package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sedyh/ebitengine-mcp/internal/errs"
)

const (
	FlagURL = "emcp-url"
	FlagPub = "emcp-pub"
	FlagSub = "emcp-sub"
	FlagID  = "emcp-id"
	BinName = "emcp-bin"
)

func Run(target, url, pub, sub, id string) (e error) {
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("stat %s: %w", target, err)
	}

	pkg := target
	if !info.IsDir() {
		pkg = filepath.Dir(target)
	}

	tmp, err := os.CreateTemp("", fmt.Sprintf("%s-%s", BinName, id))
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	out := tmp.Name()
	defer os.Remove(out)
	defer errs.Closer(&e, tmp)

	bin, err := Compiler()
	if err != nil {
		return fmt.Errorf("find compiler: %w", err)
	}

	if err := Build(pkg, bin, out); err != nil {
		return err
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		return fmt.Errorf("output file not found: %w", err)
	}

	run := exec.Command(
		out,
		"-"+FlagURL, url,
		"-"+FlagPub, pub,
		"-"+FlagSub, sub,
		"-"+FlagID, id,
	)
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

func Build(dir, bin, out string) (e error) {
	old, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}
	defer errs.Closer(&e, Restore(old))

	build := exec.Command(bin, "build", "-o", out, ".")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("build project: %w", err)
	}

	return nil
}

func Restore(dir string) io.Closer {
	return errs.CloserFunc(func() error {
		return os.Chdir(dir)
	})
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
