package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

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

	router.StartServer()
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
