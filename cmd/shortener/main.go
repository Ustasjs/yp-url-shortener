package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	"Ustasjs/yp-url-shortener/internal/router"
)

func main() {
	if pprofAddr := os.Getenv("PPROF_ADDRESS"); pprofAddr != "" {
		go func() {
			_ = http.ListenAndServe(pprofAddr, nil)
		}()
	}

	router.StartServer()
}
