package apihttp

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type metrics struct {
	requestsTotal atomic.Uint64
	jobsEnqueued  atomic.Uint64
}

func (m *metrics) requestCounterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.requestsTotal.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (m *metrics) markJobEnqueued() {
	m.jobsEnqueued.Add(1)
}

func (m *metrics) metricsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "statesight_requests_total %d\n", m.requestsTotal.Load())
	_, _ = fmt.Fprintf(w, "statesight_jobs_enqueued_total %d\n", m.jobsEnqueued.Load())
}
