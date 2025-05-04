package main

import (
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/sedyh/ebitengine-mcp/internal/cli"
	"github.com/sedyh/ebitengine-mcp/internal/out"
)

func main() {
	url := flag.String(cli.FlagURL, "", "url to run the test")
	pub := flag.String(cli.FlagPub, "", "pub to run the test")
	sub := flag.String(cli.FlagSub, "", "sub to run the test")
	id := flag.String(cli.FlagID, "", "id to run the test")
	flag.Parse()

	out.Setup(out.DefaultLevel)

	slog.Info("bin started", "url", *url, "pub", *pub, "sub", *sub, "id", *id)
	<-time.After(1 * time.Second)
	os.Exit(1)
	slog.Info("bin stopped")
}
