package main

import (
	"os"

	"github.com/tiagokriok/kanji/cmd/kanji/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
