package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/router"

	"go.uber.org/zap"
)

func main() {
	if pprofAddr := os.Getenv("PPROF_ADDRESS"); pprofAddr != "" {
		if err := logger.Initialize(zap.NewAtomicLevel()); err != nil {
			panic(err)
		}

		go func() {
			if err := http.ListenAndServe(pprofAddr, nil); err != nil {
				logger.Log.Error("pprof server failed", zap.String("address", pprofAddr), zap.Error(err))
			}
		}()
	}

	router.StartServer()
}
