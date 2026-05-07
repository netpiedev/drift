package app

import "github.com/netpiedev/drift/core/internal/cli"

func Execute() error {
	return cli.NewRootCmd().Execute()
}
