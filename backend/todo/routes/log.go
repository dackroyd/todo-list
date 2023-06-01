package routes

import (
	"net/http"
	"time"

	"golang.org/x/exp/slog"
)

func requestLog(h http.Handler, logger *slog.Logger, route string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := &CaptureWriter{w: w}

		start := time.Now()
		h.ServeHTTP(cw, r)
		dur := time.Since(start)

		sc := cw.StatusCode()
		// TODO: more HTTP attributes...
		log := logger.With(slog.String("http.path", r.URL.Path), slog.String("http.route", route), slog.Int("http.status", sc), slog.Duration("http.request_duration", dur))

		lvl := codeToLevel(sc)

		if sc >= http.StatusBadRequest {
			log.Log(r.Context(), lvl, "HTTP Request Error")
			return
		}

		log.Log(r.Context(), lvl, "HTTP Request Success")
	})
}

func codeToLevel(c int) slog.Level {
	var l slog.Level

	switch {
	case c < http.StatusBadRequest || c == http.StatusNotFound:
		l = slog.LevelInfo
	case c < http.StatusInternalServerError:
		l = slog.LevelWarn
	default:
		l = slog.LevelError
	}

	return l
}

// TODO: 'unwrap' function for compat w/ ResponseController?
type CaptureWriter struct {
	w http.ResponseWriter

	bytes      int
	statusCode int
}

func (w *CaptureWriter) Header() http.Header {
	return w.w.Header()
}

func (w *CaptureWriter) Write(b []byte) (int, error) {
	n, err := w.w.Write(b)
	w.bytes += n
	return n, err
}

func (w *CaptureWriter) WriteHeader(statusCode int) {
	w.w.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func (w *CaptureWriter) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}

	return w.statusCode
}
