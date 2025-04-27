package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sedyh/ebitengine-mcp/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("please specify the path to a Go program")
	}

	target := os.Args[1]
	args := os.Args[2:]

	fi, err := os.Stat(target)
	if err != nil {
		log.Fatalf("error accessing %s: %v", target, err)
	}

	var pkg string
	if fi.IsDir() {
		pkg = target
	} else {
		pkg = filepath.Dir(target)
	}

	tmp, err := os.CreateTemp("", "gorun-*.exe")
	if err != nil {
		log.Fatalf("failed to create temp file: %v", err)
	}
	tmp.Close()
	out := tmp.Name()
	defer os.Remove(out)

	gobin, err := cli.Compiler()
	if err != nil {
		log.Fatalf("failed to find Go binary: %v", err)
	}

	old, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(pkg); err != nil {
		log.Fatalf("failed to change to directory %s: %v", pkg, err)
	}

	build := exec.Command(gobin, "build", "-o", out, ".")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr

	if err := build.Run(); err != nil {
		os.Chdir(old)
		log.Fatalf("build error: %v", err)
	}

	if err := os.Chdir(old); err != nil {
		log.Fatalf("failed to return to original directory: %v", err)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		log.Fatalf("compiled file not found: %v", err)
	}

	run := exec.Command(out, args...)
	run.Stdin = os.Stdin
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	run.Env = os.Environ()

	if err := run.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		log.Fatalf("run error: %v", err)
	}
}
