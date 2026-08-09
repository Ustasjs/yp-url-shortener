package grpcserver_test

import (
	"context"
	"net"
	"testing"
	"time"

	"Ustasjs/yp-url-shortener/internal/grpcserver"

	"github.com/stretchr/testify/require"
)

// freePort binds an ephemeral port, closes it and returns the address, so the
// server under test gets one that is very likely still free.
func freePort(t *testing.T) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := lis.Addr().String()
	require.NoError(t, lis.Close())

	return addr
}

func TestServerListenAndShutdown(t *testing.T) {
	addr := freePort(t)
	srv := grpcserver.New(addr, &fakeURLService{}, &fakeUserRepo{userID: testUserID}, nil)

	served := make(chan error, 1)
	go func() { served <- srv.ListenAndServe() }()

	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err != nil {
			return false
		}
		return conn.Close() == nil
	}, 3*time.Second, 20*time.Millisecond, "server never started listening")

	require.NoError(t, srv.Shutdown(context.Background()))
	// A graceful stop is not a failure, so ListenAndServe must report no error.
	require.NoError(t, <-served)
}

func TestServerListenOnBadAddress(t *testing.T) {
	srv := grpcserver.New("not-an-address", &fakeURLService{}, &fakeUserRepo{userID: testUserID}, nil)

	require.Error(t, srv.ListenAndServe())
}
