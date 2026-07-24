package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ulibaysya/caseit/internal/app"
	"github.com/ulibaysya/caseit/internal/auth/key"
	"github.com/ulibaysya/caseit/internal/auth/key/keygen"
	"github.com/ulibaysya/caseit/internal/user"
)

func main() {
	if err := run(context.Background()); err != nil {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
		slog.Error("running app", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// TODO config
	pool, err := pgxpool.New(ctx, "postgresql://caselya@127.0.0.1:5432/caselya")
	if err != nil {
		return fmt.Errorf("initializing db connection: %w", err)
	}
	defer pool.Close()
	if err = pool.Ping(ctx); err != nil {
		return fmt.Errorf("pinging db: %w", err)
	}

	// TODO config
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	userStorage := user.NewStorage(pool)
	authKeyStorage := key.NewStorage(pool)

	authKeyService := key.NewService(authKeyStorage, keygen.NewUsingB64url(32))

	userHandler := user.NewHandler(logger, userStorage)
	authKeyHandler := key.NewHandler(logger, authKeyService)

	handler := app.NewHandler(logger, userHandler, authKeyHandler)

	httpServ := http.Server{
		Addr:    ":8080",
		Handler: handler,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	// TODO config
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return fmt.Errorf("initializing http listener: %w", err)
	}
	defer listener.Close() //nolint:errcheck

	errCh := make(chan error, 1)
	go func() {
		if err := httpServ.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	var errs []error
	select {
	case <-ctx.Done():
	case err = <-errCh:
		logger.Error("listening on http server", "error", err)
	}
	cancel()

	// TODO config
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	if err = httpServ.Shutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		logger.Error("shutting down http server", "error", err)
	}

	logger.Info("http server shut down, closing app")

	return errors.Join(errs...)
}
