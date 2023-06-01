package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/dackroyd/todo-list/backend/todo/routes"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

func Root(logger *slog.Logger) *cobra.Command {
	var cfg Config

	root := &cobra.Command{
		Use:   "backend",
		Short: "Backend API for managing todo lists",
		Long:  `A simple todo list manager`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd.Context(), &cfg, logger, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}

	root.PersistentFlags().StringVarP(&cfg.Host, "host", "H", "127.0.0.1", "Host interface address for the server")
	root.PersistentFlags().IntVarP(&cfg.Port, "port", "P", 8080, "HTTP port which the server listens on")

	return root
}

type Config struct {
	Host string
	Port int
}

func Run(ctx context.Context, cfg *Config, logger *slog.Logger, stdout, stderr io.Writer) error {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s := newServer(logger)

	io.WriteString(stdout, fmt.Sprintf("Ready to accept requests on http://%s\n", addr))

	return runServer(ctx, s, lis)
}

func newServer(logger *slog.Logger) *http.Server {
	s := &http.Server{
		Handler: routes.Handler(logger),
	}

	return s
}

func runServer(ctx context.Context, s *http.Server, lis net.Listener) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := s.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		return s.Shutdown(context.Background())
	})

	return g.Wait()
}
