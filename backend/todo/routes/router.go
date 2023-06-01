package routes

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/exp/slog"
)

func Handler(logger *slog.Logger) http.Handler {
	m := mux{logger: logger, router: httprouter.New()}

	m.handlerFunc(http.MethodGet, "/ping", Ping)

	return m.router
}

type mux struct {
	logger *slog.Logger
	router *httprouter.Router
}

func (m *mux) handler(method, route string, h http.Handler) {
	w := requestLog(h, m.logger, route)

	m.router.Handler(method, route, w)
}

func (m *mux) handlerFunc(method, route string, h http.HandlerFunc) {
	m.handler(method, route, http.HandlerFunc(h))
}
