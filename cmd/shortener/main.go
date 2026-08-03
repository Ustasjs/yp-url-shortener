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

	"Ustasjs/yp-url-shortener/internal/grpcserver"
	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/router"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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

	g, gCtx := errgroup.WithContext(ctx)

	// Run the servers. A non-graceful failure returns an error, which cancels
	// gCtx and triggers the shutdown goroutine below.
	g.Go(func() error {
		if err := app.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		if err := app.GRPCServer.ListenAndServe(); err != nil && !errors.Is(err, grpcserver.ErrServerStopped) {
			return err
		}
		return nil
	})

	// Wait for a shutdown signal (or a server failure) via gCtx, then tear the
	// components down in order: the servers first, then the dependencies they
	// use.
	g.Go(func() error {
		<-gCtx.Done()
		logger.Log.Info("Shutting down servers")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		httpErr := app.Server.Shutdown(shutdownCtx)
		grpcErr := app.GRPCServer.Shutdown(shutdownCtx)
		if err := errors.Join(httpErr, grpcErr); err != nil {
			return err
		}

		app.Deleter.Stop()
		logger.Log.Info("Pending deletions flushed")

		app.Notifier.Close()

		if app.DB != nil {
			return app.DB.Close()
		}
		return nil
	})

	// The root goroutine blocks until every child goroutine has stopped and
	// logs the first critical error, if any.
	if err := g.Wait(); err != nil {
		logger.Log.Fatal("Server terminated with error", zap.Error(err))
	}
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
