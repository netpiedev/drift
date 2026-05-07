package main

import (
	"os"

	"github.com/netpiedev/drift/core/internal/app"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
