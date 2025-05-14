package internal

import (
	"cloud-test/internal/backends"
	"cloud-test/internal/clients"
	"cloud-test/internal/configuration"
	"cloud-test/internal/datebase"
	"cloud-test/internal/handler"
	"cloud-test/internal/lb"
	"cloud-test/internal/proxy"
	"cloud-test/internal/ratelimit"
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

func Run(cfg configuration.Config, ctx context.Context) error {
	err := datebase.MigrateUp(cfg.PgDb)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %v", err)
	}

	repo := datebase.NewRepo(datebase.GetPostgresConnectionString(cfg.PgDb))
	clientManager, err := clients.NewClientManager(ctx, repo)
	if err != nil {
		return fmt.Errorf("failed to create client manager: %v", err)
	}

	backendsManager, err := backends.NewRegistry(ctx, cfg.Loadbalancer, &http.Client{})
	if err != nil {
		return fmt.Errorf("failed to create backend manager: %v", err)
	}
	balancer, err := lb.NewRoundRobin(backendsManager, ctx)
	if err != nil {
		return fmt.Errorf("failed to create balancer: %v", err)
	}

	redisClient := datebase.NewRedisRepo(cfg.Redis)
	limiter, err := ratelimit.NewLimiterManager(ctx, redisClient, clientManager, cfg.RateLimit)
	if err != nil {
		return fmt.Errorf("failed to create rate limiter: %v", err)
	}

	clientHandler := handler.NewClientHandler(clientManager, limiter)
	proxyHandler := proxy.NewProxy(balancer, limiter)

	h1 := http.NewServeMux()
	h1.Handle("/clients", clientHandler)
	clientServer := &http.Server{
		Addr:    cfg.AdminListen,
		Handler: h1,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	h2 := http.NewServeMux()
	h2.Handle("/", proxyHandler)
	proxyServer := &http.Server{
		Addr:    cfg.ProxyListen,
		Handler: h2,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	log.Printf("starting client server on %s", cfg.AdminListen)
	go func() {
		if err := clientServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start client server", "error", err)
		}
	}()
	log.Printf("starting proxy server on %s", cfg.ProxyListen)
	go func() {
		if err := proxyServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start proxy server", "error", err)
		}
	}()

	<-ctx.Done()

	gracefulCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := clientServer.Shutdown(gracefulCtx); err != nil {
			slog.Error("failed to shutdown client server", "error", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := proxyServer.Shutdown(gracefulCtx); err != nil {
			slog.Error("failed to shutdown proxy server", "error", err)
		}
	}()
	wg.Wait()
	return nil
}
