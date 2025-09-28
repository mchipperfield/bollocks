package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chipperfieldm/bollocks/api.bollocks.social/api"
)

const (
	Service string = "api.bollocks.social"
)

func main() {
	logger := slog.Default().With("service", Service)
	logger.Info("initiating service")

	flags := flag.NewFlagSet("", flag.ContinueOnError)
	var (
		port = flags.Int("port", 8080, "port for API to listen on")
	)

	if err := flags.Parse(os.Args[1:]); err != nil {
		logger.Error("Failed to parse flags", "error", err)
		os.Exit(1)
	}

	mux := api.NewHandler()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", srv.Addr)
		errChan <- srv.ListenAndServe()
	}()

	select {
	case err := <-errChan:
		logger.Info("listen and serve", "error", err, "addr", srv.Addr)
	case sig := <-stopChan:
		logger.Info("shutdown signal received", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("gracefully shutting down server", "error", err, "addr", srv.Addr)
			os.Exit(1)
		}
		logger.Info("server gracefully shutdown", "addr", srv.Addr)
	}
}
