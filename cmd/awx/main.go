package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/salmonumbrella/airwallex-cli/internal/cmd"
	"github.com/salmonumbrella/airwallex-cli/internal/exitcode"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := cmd.ExecuteContext(ctx, os.Args[1:])
	os.Exit(exitcode.FromError(err))
}
