package grpcserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/model"
	shortenerv1 "Ustasjs/yp-url-shortener/pkg/proto/shortener/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// URLService is the business logic behind the RPC handlers.
// *urlservice.Service implements it.
type URLService interface {
	Shorten(ctx context.Context, rawURL string, userID string) (string, error)
	Expand(ctx context.Context, id string, userID string) (string, error)
	ListUserURLs(ctx context.Context, userID string) ([]model.UserURLItem, error)
}

// Server runs the ShortenerService on addr. Its API copies *http.Server, so the
// caller can start and stop both servers the same way.
type Server struct {
	grpc *grpc.Server
	addr string
}

// New builds a Server that serves urls and identifies callers through users.
// tlsConfig may be nil, and then the server uses plain HTTP/2.
func New(addr string, urls URLService, users middleware.UserRepository, tlsConfig *tls.Config) *Server {
	// Logging comes first so that requests rejected by auth are logged too. The
	// HTTP middleware chain has the same order.
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(LoggingInterceptor(), AuthInterceptor(users)),
	}
	if tlsConfig != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	srv := grpc.NewServer(opts...)
	shortenerv1.RegisterShortenerServiceServer(srv, &ShortenerServer{urls: urls})

	return &Server{grpc: srv, addr: addr}
}

// Serve serves on lis and blocks until the server stops. Tests use it to serve
// over an in-memory listener.
func (s *Server) Serve(lis net.Listener) error {
	return s.grpc.Serve(lis)
}

// ListenAndServe listens on the configured address and blocks until the server
// stops. It returns nil after Shutdown and grpc.ErrServerStopped after Stop.
func (s *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("grpc listen on %s: %w", s.addr, err)
	}
	return s.Serve(lis)
}

// Shutdown stops the server after the running calls finish. If ctx runs out
// first, it kills the remaining connections and returns the ctx error.
func (s *Server) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.grpc.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.grpc.Stop()
		return ctx.Err()
	}
}

// Stop stops the server at once, without waiting for running calls.
func (s *Server) Stop() {
	s.grpc.Stop()
}
