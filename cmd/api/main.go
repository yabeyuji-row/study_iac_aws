package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"study_iac_aws/internal/config"
	"study_iac_aws/internal/database"
	"study_iac_aws/internal/observability"
	"study_iac_aws/internal/todo"
	"study_iac_aws/internal/web"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	healthcheckURL := flag.String("healthcheck-url", "", "URL to check and exit")
	flag.Parse()

	if *healthcheckURL != "" {
		if err := checkHealth(*healthcheckURL); err != nil {
			slog.Error("healthcheck failed", "error", err)
			os.Exit(1)
		}
		return
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

type databasePinger interface {
	Ping(ctx context.Context) error
}

type buildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

func run(ctx context.Context) (err error) {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer pool.Close()

	shutdownTracing, err := observability.ConfigureTracing(ctx, cfg.OTELExporterEndpoint)
	if err != nil {
		return fmt.Errorf("configure tracing: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutdownErr := shutdownTracing(shutdownCtx); shutdownErr != nil {
			slog.Error("shutdown tracing", "error", shutdownErr)
		}
	}()

	repository := todo.NewPostgresRepository(pool)
	service := todo.NewService(repository, time.Now)
	handler := todo.NewHandler(service)

	echoServer := newEchoServer(handler, pool, buildInfo{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	}, cfg.FaultInjection)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           echoServer,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	slog.Info("starting api", "addr", cfg.HTTPAddr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- echoServer.StartServer(server)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		slog.Info("shutting down api")
		if err := echoServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
		err = <-errCh
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("listen and serve: %w", err)
		}
		return nil
	case err = <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("listen and serve: %w", err)
		}
		return nil
	}
}

func newEchoServer(
	todoHandler *todo.Handler,
	pinger databasePinger,
	info buildInfo,
	faultInjection bool,
) (echoServer *echo.Echo) {
	echoServer = echo.New()
	echoServer.HideBanner = true
	metrics := observability.NewHTTPMetrics()
	echoServer.Use(requestLogMiddleware())
	echoServer.Use(otelecho.Middleware("todo-api"))
	echoServer.Use(observability.HTTPMiddleware(metrics))

	healthHandler := func(echoContext echo.Context) error {
		return echoContext.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	}
	readyHandler := func(echoContext echo.Context) error {
		ctx, cancel := context.WithTimeout(echoContext.Request().Context(), 2*time.Second)
		defer cancel()

		if err := pinger.Ping(ctx); err != nil {
			return echoContext.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "unready",
			})
		}

		return echoContext.JSON(http.StatusOK, map[string]string{
			"status": "ready",
		})
	}

	echoServer.GET("/health", healthHandler)
	echoServer.GET("/healthz", healthHandler)
	echoServer.GET("/ready", readyHandler)
	echoServer.GET("/readyz", readyHandler)
	echoServer.GET("/version", func(echoContext echo.Context) error {
		return echoContext.JSON(http.StatusOK, info)
	})
	echoServer.GET("/metrics", echo.WrapHandler(metrics.Handler()))
	if faultInjection {
		registerFaultInjectionRoutes(echoServer)
	}

	todoHandler.Register(echoServer.Group("/v1"))
	echoServer.Any("/*", echo.WrapHandler(web.Handler()))

	return echoServer
}

func registerFaultInjectionRoutes(echoServer *echo.Echo) {
	const requiredHeader = "X-Fault-Injection"
	requireOptIn := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoContext echo.Context) error {
			if echoContext.Request().Header.Get(requiredHeader) != "enabled" {
				return echoContext.JSON(http.StatusForbidden, map[string]string{"error": "fault injection opt-in required"})
			}
			return next(echoContext)
		}
	}

	group := echoServer.Group("/debug/fault", requireOptIn)
	group.GET("/5xx", func(echoContext echo.Context) error {
		return echoContext.JSON(http.StatusInternalServerError, map[string]string{"error": "injected failure"})
	})
	group.GET("/delay", func(echoContext echo.Context) error {
		const maxDelay = 5 * time.Second
		delay, err := time.ParseDuration(echoContext.QueryParam("duration"))
		if err != nil || delay < 0 || delay > maxDelay {
			return echoContext.JSON(http.StatusBadRequest, map[string]string{"error": "duration must be between 0s and 5s"})
		}
		select {
		case <-time.After(delay):
			return echoContext.JSON(http.StatusOK, map[string]string{"status": "delayed"})
		case <-echoContext.Request().Context().Done():
			return echoContext.Request().Context().Err()
		}
	})
}

func requestLogMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoContext echo.Context) (err error) {
			started := time.Now()
			request := echoContext.Request()
			response := echoContext.Response()

			err = next(echoContext)
			if err != nil {
				echoContext.Error(err)
			}

			status := response.Status
			if status == 0 {
				status = http.StatusOK
			}

			duration := time.Since(started)
			slog.InfoContext(
				request.Context(),
				"http request completed",
				"event", "http_request_completed",
				"method", request.Method,
				"path", request.URL.Path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"remote_addr", request.RemoteAddr,
				"user_agent", request.UserAgent(),
				"request_id", request.Header.Get(echo.HeaderXRequestID),
			)

			return nil
		}
	}
}

func checkHealth(rawURL string) (err error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 1 * time.Second,
			}).DialContext,
		},
	}

	response, err := client.Get(rawURL)
	if err != nil {
		return fmt.Errorf("get health endpoint: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint status: %s", response.Status)
	}

	return nil
}
