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

	firebase "firebase.google.com/go"
	"github.com/chipperfieldm/bollocks/api.bollocks.social/api"
	"github.com/gorilla/handlers"
)

const (
	Service string = "api.bollocks.social"
)

func main() {
	logger := &Slogger{
		slog.Default().With(slog.Group("service", "name", Service)),
	}
	logger.Log("initiating service")

	flags := flag.NewFlagSet("", flag.ContinueOnError)
	var (
		port         = flags.Int("port", 8080, "port for API to listen on")
		geminiAPIKey = flags.String("gemini-api-key", "", "API key for the Google Gemini service")
	)

	if err := flags.Parse(os.Args[1:]); err != nil {
		logger.Log("Failed to parse flags", "error", err)
		os.Exit(1)
	}

	firebaseApp, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		logger.Log("failed to create firebase app", "error", err)
		os.Exit(1)
	}
	auth, err := firebaseApp.Auth(context.Background())
	if err != nil {
		logger.Log("failed to create firebase auth client", "error", err)
		os.Exit(1)
	}
	client, err := firebaseApp.Firestore(context.Background())
	if err != nil {
		logger.Log("failed to create firestore client", "error", err)
		os.Exit(1)
	}

	authMw := api.VerifyToken(auth)

	panicMw := api.PanicMw(logger)

	corsMw := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:5173"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
	)

	loggingMw := api.LoggingMiddleware(logger)

	mux := api.NewHandler(logger, client, *geminiAPIKey)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      panicMw(loggingMw(corsMw(authMw(mux)))),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		logger.Log("http server listening", "addr", srv.Addr)
		errChan <- srv.ListenAndServe()
	}()

	select {
	case err := <-errChan:
		logger.Log("listen and serve", "error", err, "addr", srv.Addr)
	case sig := <-stopChan:
		logger.Log("shutdown signal received", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Log("gracefully shutting down server", "error", err, "addr", srv.Addr)
			os.Exit(1)
		}
		logger.Log("server gracefully shutdown", "addr", srv.Addr)
	}
}
