package cmd_test

import (
	"context"
	"os"
	"testing"

	"github.com/dackroyd/todo-list/backend/cmd"
	"golang.org/x/exp/slog"
)

func TestRoot(t *testing.T) {
	root := cmd.Root(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{})))

	root.SetArgs([]string{"backend", "--port=9090"})

	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
}
