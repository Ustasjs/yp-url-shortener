package main

import (
	"net/http"
	_ "net/http/pprof"

	"Ustasjs/yp-url-shortener/internal/router"
)

func main() {
	// pprof debug endpoints (heap, profile, goroutine, ...) are served on a
	// separate port and kept off the main chi router, so the app middleware
	// (gzip compression / auth) can't corrupt the raw profile output.
	go func() {
		_ = http.ListenAndServe("localhost:6060", nil)
	}()

	router.StartServer()
}
