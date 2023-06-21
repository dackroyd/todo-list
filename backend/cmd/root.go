package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/lib/pq"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"

	"github.com/dackroyd/todo-list/backend/todo/database"
	"github.com/dackroyd/todo-list/backend/todo/routes"
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

	root.PersistentFlags().StringVar(&cfg.DBConn, "dburl", "postgres://todo:password@127.0.0.1/todo?sslmode=disable", "DB connection string")
	root.PersistentFlags().StringVarP(&cfg.Host, "host", "H", "127.0.0.1", "Host interface address for the server")
	root.PersistentFlags().IntVarP(&cfg.Port, "port", "P", 8080, "HTTP port which the server listens on")

	return root
}

type Config struct {
	DBConn string
	Host   string
	Port   int
}

func Run(ctx context.Context, cfg *Config, logger *slog.Logger, stdout, stderr io.Writer) error {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	db, err := openDB(cfg.DBConn)
	if err != nil {
		return fmt.Errorf("unable open DB: %w", err)
	}

	listRepo := database.NewListRepository(db)
	listsAPI := routes.NewListAPI(listRepo)

	s := newServer(logger, listsAPI)

	io.WriteString(stdout, fmt.Sprintf("Ready to accept requests on http://%s\n", addr))

	return runServer(ctx, s, lis)
}

func openDB(connURL string) (*sql.DB, error) {
	conn, err := pq.NewConnector(connURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse DB connection string: %w", err)
	}

	return sql.OpenDB(conn), nil
}

func newServer(logger *slog.Logger, lists *routes.ListsAPI) *http.Server {
	s := &http.Server{
		Handler: routes.Handler(lists, logger),
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
