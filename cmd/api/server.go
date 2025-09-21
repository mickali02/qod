// Filename: cmd/api/server.go

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	//"os"
	"os/signal"
	"syscall"
	"net/http"
	"time"
)


// serve starts the HTTP server
func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// Buffer 1 so the goroutine can send without blocking.
	shutdownErr := make(chan error, 1)

	// Listen for SIGINT/SIGTERM in the background.
	go func() {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		<-ctx.Done() // wait for signal

		app.logger.Info("shutting down server...")

		// Give in-flight requests up to 30s to finish.
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Initiate graceful shutdown.
		shutdownErr <- srv.Shutdown(timeoutCtx)
	}()

	app.logger.Info("starting server",
		"addr", srv.Addr,
		"env", app.config.env,
		"version", version,
	)

	// Start serving (blocks until error or Shutdown).
	err := srv.ListenAndServe()
	// If it's not the expected "server closed" error, bubble up.
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Check if shutdown had any errors.
	if err := <-shutdownErr; err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)
	return nil
}

