package logger

import "net/http"

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write writes b to the underlying ResponseWriter and accumulates the response
// size.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader writes statusCode to the underlying ResponseWriter and records it
// for logging.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
