package cmd

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/XSAM/otelsql"
	"github.com/lib/pq"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/semconv/v1.17.0/netconv"
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

	// Setup Tracing: Uncomment this block
	//shutdown, err := setupTracing(ctx, logger)
	//if err != nil {
	//	return err
	//}
	//
	//defer shutdown()

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

func setupTracing(ctx context.Context, logger *slog.Logger) (shutdown func(), err error) {
	r, err := resource.New(ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceName("todo-list-api"),
			semconv.ServiceVersion("v0.1.0"),
			attribute.String("environment", "demo"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint())
	if err != nil {
		return nil, fmt.Errorf("creating Jaeger trace exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(r),
	)

	shutdown = func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("Failed to shutdown tracing provider", slog.String("error", err.Error()))
		}
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return shutdown, nil
}

func openDB(connURL string) (*sql.DB, error) {
	conn, err := pq.NewConnector(connURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse DB connection string: %w", err)
	}

	return sql.OpenDB(conn), nil
	// Instrument Database Calls: Replace the line above with the one below
	//return traceDB(connURL, conn)
}

func traceDB(connURL string, conn driver.Connector) (*sql.DB, error) {
	connURI, err := url.ParseRequestURI(connURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse DB connection string: %w", err)
	}

	// 🔐 Drop the password before inclusion in tracing attributes
	connURI.User = url.User(connURI.User.Username())

	db := otelsql.OpenDB(conn,
		otelsql.WithSpanOptions(otelsql.SpanOptions{OmitConnResetSession: true}),
		otelsql.WithAttributes(
			semconv.DBSystemPostgreSQL,
			semconv.DBConnectionString(connURI.String()),
			semconv.DBName(strings.TrimPrefix(connURI.Path, "/")),
			semconv.DBUser(connURI.User.Username()),
		),
		otelsql.WithAttributesGetter(func(ctx context.Context, method otelsql.Method, query string, args []driver.NamedValue) []attribute.KeyValue {
			attrs := []attribute.KeyValue{
				semconv.DBOperationKey.String(string(method)),
			}

			attrs = append(attrs, netconv.Client(connURI.Host, nil)...)

			for _, arg := range args {
				name := "db.statement.args."
				if arg.Name == "" {
					name += fmt.Sprintf("$%d", arg.Ordinal)
				} else {
					name += arg.Name
				}
				attrs = append(attrs, attribute.String(name, fmt.Sprintf("%v", arg.Value)))
			}

			return attrs
		}),
	)

	return db, nil
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
