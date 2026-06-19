package main

import (
	"os"

	"github.com/hrodrig/kiko/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
