package routes

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/exp/slog"
)

func Handler(lists *ListsAPI, logger *slog.Logger) http.Handler {
	m := mux{logger: logger, router: httprouter.New()}

	m.handlerFunc(http.MethodGet, "/api/v1/lists", lists.Lists)
	m.handlerFunc(http.MethodGet, "/api/v1/lists/:list_id", lists.List)
	m.handlerFunc(http.MethodGet, "/api/v1/lists/:list_id/items", lists.Items)
	m.handlerFunc(http.MethodGet, "/ping", Ping)

	return m.router
}

type mux struct {
	logger *slog.Logger
	router *httprouter.Router
}

func (m *mux) handler(method, route string, h http.Handler) {
	w := requestLog(h, m.logger, route)
	// Instrument HTTP Handlers: Uncomment the line below
	//w = otelhttp.NewHandler(w, method+" "+route)

	m.router.Handler(method, route, w)
}

func (m *mux) handlerFunc(method, route string, h http.HandlerFunc) {
	m.handler(method, route, h)
}

// HACK: forcing the import to be kept, so that enabling instrumentation only requires uncommenting in 'handler'
var _ = otelhttp.NewHandler
