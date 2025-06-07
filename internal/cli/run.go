package cli

import (
	"context"
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

func Run(ctx context.Context, target, url, pub, sub, id string) (l string, e error) {
	info, err := os.Stat(target)
	if err != nil {
		return "", fmt.Errorf("stat %s: %w", target, err)
	}

	pkg := target
	if !info.IsDir() {
		pkg = filepath.Dir(target)
	}

	tmp, err := os.CreateTemp("", fmt.Sprintf("%s-%s", BinName, id))
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}

	out := tmp.Name()
	defer os.Remove(out)
	defer errs.Closer(&e, tmp)

	bin, err := Compiler()
	if err != nil {
		return "", fmt.Errorf("find compiler: %w", err)
	}

	log, err := Build(ctx, pkg, bin, out)
	if err != nil {
		return log, fmt.Errorf("build: %w", err)
	}

	log, err = Launch(ctx, out, url, pub, sub, id)
	if err != nil {
		return log, fmt.Errorf("launch: %w", err)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		return "", fmt.Errorf("output file not found: %w", err)
	}

	return log, nil
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

func Build(ctx context.Context, dir, bin, out string) (log string, e error) {
	old, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working dir: %w", err)
	}
	if err := os.Chdir(dir); err != nil {
		return "", err
	}
	defer errs.Closer(&e, Restore(old))

	build := exec.CommandContext(
		ctx, bin,
		"build", "-o",
		out, ".",
	)
	data, err := build.CombinedOutput()

	return Trim(string(data)), WrapWithContext(err, ctx)
}

func Launch(ctx context.Context, out, url, pub, sub, id string) (log string, e error) {
	run := exec.CommandContext(
		ctx, out,
		"-"+FlagURL, url,
		"-"+FlagPub, pub,
		"-"+FlagSub, sub,
		"-"+FlagID, id,
	)
	data, err := run.CombinedOutput()

	return Trim(string(data)), WrapWithContext(err, ctx)
}

func WrapWithContext(err error, ctx context.Context) error {
	if e := ctx.Err(); e != nil {
		return fmt.Errorf("%w: %w", e, err)
	}
	return err
}

func Restore(dir string) io.Closer {
	return errs.CloserFunc(func() error {
		return os.Chdir(dir)
	})
}
