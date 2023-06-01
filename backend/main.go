package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/exp/slog"

	"github.com/dackroyd/todo-list/backend/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var logLevel slog.LevelVar

	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: &logLevel})
	logger := slog.New(h)
	slog.SetDefault(logger)

	root := cmd.Root(logger)

	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
