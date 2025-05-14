package proxy

import (
	"cloud-test/internal/lb"
	helper "cloud-test/internal/net"
	"cloud-test/internal/ratelimit"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
)

type Proxy struct {
	balancer lb.Balancer
	limiter  ratelimit.Limiter
}

func NewProxy(b lb.Balancer, r ratelimit.Limiter) http.Handler {
	return &Proxy{balancer: b, limiter: r}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.logRecievedRequest(r)

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, "Invalid client IP")
		return
	}

	if !p.limiter.Allow(r.Context(), host) {
		slog.Warn("Rate limit exceeded", "client_ip", host)
		helper.WriteError(w, http.StatusTooManyRequests, "Rate limit exceeded")
		return
	}
	backend, err := p.balancer.Next()
	if err != nil {
		helper.WriteError(w, http.StatusServiceUnavailable, "No available services. Please try again later")
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(&backend.Url)
	proxy.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, e error) {
		if e, ok := e.(net.Error); ok && e.Timeout() {
			slog.Warn("Timeout error", "error", e)
			helper.WriteError(rw, http.StatusGatewayTimeout, "Gateway Timeout")
			return
		}

		slog.Error("Bad Gateway", "error", e)
		helper.WriteError(rw, http.StatusBadGateway, "Bad Gateway")
	}
	proxy.ServeHTTP(w, r)
}

func (p *Proxy) logRecievedRequest(r *http.Request) {
	slog.Info("Received request", "method", r.Method, "url", r.URL.String(), "remote_addr", r.RemoteAddr, "headers", r.Header)
}
