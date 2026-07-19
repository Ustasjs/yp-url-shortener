package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/router"

	"go.uber.org/zap"
)

// Build information injected at link time via
// -ldflags "-X main.buildVersion=... -X main.buildDate=... -X main.buildCommit=...".
// The default "N/A" values are overwritten at compile time when the flags are provided.
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	printBuildInfo()

	if pprofAddr := os.Getenv("PPROF_ADDRESS"); pprofAddr != "" {
		if err := logger.Initialize(zap.NewAtomicLevel()); err != nil {
			log.Fatal(err)
		}

		go func() {
			if err := http.ListenAndServe(pprofAddr, nil); err != nil {
				logger.Log.Error("pprof server failed", zap.String("address", pprofAddr), zap.Error(err))
			}
		}()
	}

	app := router.Setup()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- app.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("Server failed", zap.Error(err))
		}
	case <-ctx.Done():
		stop()
		logger.Log.Info("Shutting down server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.Server.Shutdown(shutdownCtx); err != nil {
			logger.Log.Error("Server shutdown failed", zap.Error(err))
		}

		app.Deleter.Stop()
		logger.Log.Info("Pending deletions flushed")

		app.Notifier.Close()

		if app.DB != nil {
			if err := app.DB.Close(); err != nil {
				logger.Log.Error("close database failed", zap.Error(err))
			}
		}
	}
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
