package grpcserver_test

import (
	"context"
	"errors"
	"testing"

	"Ustasjs/yp-url-shortener/internal/grpcserver"
	"Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const testUserID = "user-1"

var errCreateUser = errors.New("cannot create user")

type fakeUserRepo struct {
	userID string
	err    error
	calls  int
}

func (f *fakeUserRepo) CreateUser(_ context.Context) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.userID, nil
}

// fakeStream is the minimum grpc.ServerTransportStream needed by grpc.SetHeader,
// which otherwise refuses to work outside a running server.
type fakeStream struct {
	header metadata.MD
}

func (f *fakeStream) Method() string { return "/test/Method" }

func (f *fakeStream) SetHeader(md metadata.MD) error {
	f.header = metadata.Join(f.header, md)
	return nil
}

func (f *fakeStream) SendHeader(md metadata.MD) error { return f.SetHeader(md) }

func (f *fakeStream) SetTrailer(_ metadata.MD) error { return nil }

// interceptorCall runs the auth interceptor over a handler that only records the
// user ID it was given.
type interceptorCall struct {
	gotUserID string
	gotOK     bool
	called    bool
	stream    *fakeStream
	err       error
}

func runAuth(t *testing.T, users middleware.UserRepository, md metadata.MD) *interceptorCall {
	t.Helper()

	call := &interceptorCall{stream: &fakeStream{}}

	ctx := context.Background()
	if md != nil {
		ctx = metadata.NewIncomingContext(ctx, md)
	}
	ctx = grpc.NewContextWithServerTransportStream(ctx, call.stream)

	_, call.err = grpcserver.AuthInterceptor(users)(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/test/Method"},
		func(ctx context.Context, _ any) (any, error) {
			call.called = true
			call.gotUserID, call.gotOK = middleware.GetUserIDFromContext(ctx)
			return "ok", nil
		},
	)
	return call
}

func TestAuthInterceptorNoTokenCreatesUser(t *testing.T) {
	users := &fakeUserRepo{userID: testUserID}

	call := runAuth(t, users, nil)

	require.NoError(t, call.err)
	assert.True(t, call.called)
	assert.Equal(t, 1, users.calls)
	assert.True(t, call.gotOK)
	assert.Equal(t, testUserID, call.gotUserID)

	// The fresh token goes back in the response header, like Set-Cookie in HTTP.
	sent := call.stream.header.Get(grpcserver.AuthMetadataKey)
	require.Len(t, sent, 1)
	userID, err := service.GetUserID(sent[0])
	require.NoError(t, err)
	assert.Equal(t, testUserID, userID)
}

func TestAuthInterceptorReusesToken(t *testing.T) {
	token, err := service.BuildJWTString(testUserID)
	require.NoError(t, err)

	users := &fakeUserRepo{userID: "other-user"}

	call := runAuth(t, users, metadata.Pairs(grpcserver.AuthMetadataKey, token))

	require.NoError(t, call.err)
	assert.True(t, call.called)
	assert.Equal(t, 0, users.calls)
	assert.Equal(t, testUserID, call.gotUserID)
	assert.Empty(t, call.stream.header.Get(grpcserver.AuthMetadataKey))
}

// Only the bare token is supported, so a "Bearer <token>" header does not parse.
func TestAuthInterceptorRejectsBearerForm(t *testing.T) {
	token, err := service.BuildJWTString(testUserID)
	require.NoError(t, err)

	users := &fakeUserRepo{userID: testUserID}

	call := runAuth(t, users, metadata.Pairs(grpcserver.AuthMetadataKey, "Bearer "+token))

	require.Error(t, call.err)
	assert.Equal(t, codes.Unauthenticated, status.Code(call.err))
	assert.False(t, call.called)
}

func TestAuthInterceptorBrokenTokenIsUnauthenticated(t *testing.T) {
	users := &fakeUserRepo{userID: testUserID}

	call := runAuth(t, users, metadata.Pairs(grpcserver.AuthMetadataKey, "not-a-jwt"))

	require.Error(t, call.err)
	assert.Equal(t, codes.Unauthenticated, status.Code(call.err))
	assert.False(t, call.called)
	assert.Equal(t, 0, users.calls)
}

// An empty header value counts as no token, so the caller gets a new identity
// instead of an error.
func TestAuthInterceptorEmptyTokenCreatesUser(t *testing.T) {
	users := &fakeUserRepo{userID: testUserID}

	call := runAuth(t, users, metadata.Pairs(grpcserver.AuthMetadataKey, ""))

	require.NoError(t, call.err)
	assert.Equal(t, 1, users.calls)
	assert.Equal(t, testUserID, call.gotUserID)
}

func TestAuthInterceptorCreateUserFailure(t *testing.T) {
	users := &fakeUserRepo{err: errCreateUser}

	call := runAuth(t, users, nil)

	require.Error(t, call.err)
	assert.Equal(t, codes.Internal, status.Code(call.err))
	assert.False(t, call.called)
}

func TestLoggingInterceptorPassesResultThrough(t *testing.T) {
	wantErr := status.Error(codes.NotFound, "nope")

	tests := []struct {
		name     string
		resp     any
		err      error
		wantResp any
	}{
		{name: "success", resp: "ok", wantResp: "ok"},
		{name: "error", err: wantErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := grpcserver.LoggingInterceptor()(
				context.Background(),
				nil,
				&grpc.UnaryServerInfo{FullMethod: "/test/Method"},
				func(context.Context, any) (any, error) { return tt.resp, tt.err },
			)

			assert.Equal(t, tt.wantResp, resp)
			assert.Equal(t, tt.err, err)
		})
	}
}
