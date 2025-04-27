package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	DefaultFlag = "mcp"
	debug       = true
)

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

func Debug(s string) {
	if !debug {
		return
	}
	f, err := os.OpenFile("mcp.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(s + "\n"); err != nil {
		log.Fatal(err)
	}
}
