package grpcserver_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"Ustasjs/yp-url-shortener/internal/grpcserver"
	"Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/model"
	"Ustasjs/yp-url-shortener/internal/repository"
	"Ustasjs/yp-url-shortener/internal/service/urlservice"
	shortenerv1 "Ustasjs/yp-url-shortener/pkg/proto/shortener/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
)

const (
	bufSize      = 1024 * 1024
	testShortURL = "http://localhost:8080/test-id"
)

var errStorage = errors.New("storage is down")

// fakeURLService lets a test pick the exact result of every use case, including
// the errors a real MemStorage never produces.
type fakeURLService struct {
	shortURL   string
	shortenErr error
	original   string
	expandErr  error
	items      []model.UserURLItem
	listErr    error

	gotRawURL string
	gotID     string
	gotUserID string
}

func (f *fakeURLService) Shorten(_ context.Context, rawURL string, userID string) (string, error) {
	f.gotRawURL = rawURL
	f.gotUserID = userID
	return f.shortURL, f.shortenErr
}

func (f *fakeURLService) Expand(_ context.Context, id string, userID string) (string, error) {
	f.gotID = id
	f.gotUserID = userID
	return f.original, f.expandErr
}

func (f *fakeURLService) ListUserURLs(_ context.Context, userID string) ([]model.UserURLItem, error) {
	f.gotUserID = userID
	return f.items, f.listErr
}

// newTestClient starts the server on an in-memory listener and returns a client
// wired to it. Both are torn down when the test ends.
func newTestClient(t *testing.T, urls grpcserver.URLService, users middleware.UserRepository) shortenerv1.ShortenerServiceClient {
	t.Helper()

	lis := bufconn.Listen(bufSize)
	srv := grpcserver.New("bufnet", urls, users, nil)
	go func() { _ = srv.Serve(lis) }()

	// grpc.NewClient resolves the target through DNS unless the passthrough
	// scheme is spelled out, which would break the custom dialer below.
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = conn.Close()
		srv.Stop()
		_ = lis.Close()
	})

	return shortenerv1.NewShortenerServiceClient(conn)
}

func TestShortenURL(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		urls := &fakeURLService{shortURL: testShortURL}
		client := newTestClient(t, urls, &fakeUserRepo{userID: testUserID})

		resp, err := client.ShortenURL(context.Background(),
			shortenerv1.URLShortenRequest_builder{Url: proto.String("https://practicum.yandex.ru")}.Build())

		require.NoError(t, err)
		assert.Equal(t, testShortURL, resp.GetResult())
		assert.Equal(t, "https://practicum.yandex.ru", urls.gotRawURL)
		assert.Equal(t, testUserID, urls.gotUserID)
	})

	t.Run("invalid url", func(t *testing.T) {
		urls := &fakeURLService{shortenErr: urlservice.ErrInvalidURL}
		client := newTestClient(t, urls, &fakeUserRepo{userID: testUserID})

		_, err := client.ShortenURL(context.Background(), shortenerv1.URLShortenRequest_builder{Url: proto.String("nope")}.Build())

		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("storage failure", func(t *testing.T) {
		urls := &fakeURLService{shortenErr: errStorage}
		client := newTestClient(t, urls, &fakeUserRepo{userID: testUserID})

		_, err := client.ShortenURL(context.Background(),
			shortenerv1.URLShortenRequest_builder{Url: proto.String("https://practicum.yandex.ru")}.Build())

		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

// A conflict must still hand the caller the existing short URL, the way the
// HTTP API returns it in the body of the 409 response.
func TestShortenURLConflictCarriesShortURL(t *testing.T) {
	urls := &fakeURLService{shortURL: testShortURL, shortenErr: repository.ErrConflict}
	client := newTestClient(t, urls, &fakeUserRepo{userID: testUserID})

	var header metadata.MD
	_, err := client.ShortenURL(context.Background(),
		shortenerv1.URLShortenRequest_builder{Url: proto.String("https://practicum.yandex.ru")}.Build(),
		grpc.Header(&header))

	require.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
	assert.Equal(t, []string{testShortURL}, header.Get(grpcserver.ShortURLHeader))
}

func TestExpandURL(t *testing.T) {
	tests := []struct {
		name      string
		original  string
		expandErr error
		wantCode  codes.Code
		wantURL   string
	}{
		{
			name:     "success",
			original: "https://practicum.yandex.ru",
			wantCode: codes.OK,
			wantURL:  "https://practicum.yandex.ru",
		},
		{
			name:      "unknown id",
			expandErr: repository.ErrRecordNotFound,
			wantCode:  codes.NotFound,
		},
		{
			name:      "deleted url",
			expandErr: repository.ErrDeleted,
			wantCode:  codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls := &fakeURLService{original: tt.original, expandErr: tt.expandErr}
			client := newTestClient(t, urls, &fakeUserRepo{userID: testUserID})

			resp, err := client.ExpandURL(context.Background(), shortenerv1.URLExpandRequest_builder{Id: proto.String("test-id")}.Build())

			assert.Equal(t, tt.wantCode, status.Code(err))
			if tt.wantCode == codes.OK {
				require.NoError(t, err)
				assert.Equal(t, tt.wantURL, resp.GetResult())
				assert.Equal(t, "test-id", urls.gotID)
			}
		})
	}
}

func TestListUserURLs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		urls := &fakeURLService{items: []model.UserURLItem{
			{ShortURL: testShortURL, OriginalURL: "https://practicum.yandex.ru"},
			{ShortURL: "http://localhost:8080/other-id", OriginalURL: "https://ya.ru"},
		}}
		client := newTestClient(t, urls, &fakeUserRepo{userID: testUserID})

		resp, err := client.ListUserURLs(context.Background(), shortenerv1.UserURLsRequest_builder{}.Build())

		require.NoError(t, err)
		require.Len(t, resp.GetUrl(), 2)
		assert.Equal(t, testShortURL, resp.GetUrl()[0].GetShortUrl())
		assert.Equal(t, "https://practicum.yandex.ru", resp.GetUrl()[0].GetOriginalUrl())
		assert.Equal(t, "https://ya.ru", resp.GetUrl()[1].GetOriginalUrl())
		assert.Equal(t, testUserID, urls.gotUserID)
	})

	// The HTTP API answers 204 here, so gRPC must answer OK with nothing in it.
	t.Run("no urls", func(t *testing.T) {
		client := newTestClient(t, &fakeURLService{}, &fakeUserRepo{userID: testUserID})

		resp, err := client.ListUserURLs(context.Background(), shortenerv1.UserURLsRequest_builder{}.Build())

		require.NoError(t, err)
		assert.Empty(t, resp.GetUrl())
	})

	t.Run("storage failure", func(t *testing.T) {
		client := newTestClient(t, &fakeURLService{listErr: errStorage}, &fakeUserRepo{userID: testUserID})

		_, err := client.ListUserURLs(context.Background(), shortenerv1.UserURLsRequest_builder{}.Build())

		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

// The token issued on the first call must identify the same user on the next
// one, which is what makes ListUserURLs usable over gRPC.
func TestAuthTokenRoundTrip(t *testing.T) {
	urls := &fakeURLService{shortURL: testShortURL}
	client := newTestClient(t, urls, &fakeUserRepo{userID: testUserID})

	var header metadata.MD
	_, err := client.ShortenURL(context.Background(),
		shortenerv1.URLShortenRequest_builder{Url: proto.String("https://practicum.yandex.ru")}.Build(),
		grpc.Header(&header))
	require.NoError(t, err)

	tokens := header.Get(grpcserver.AuthMetadataKey)
	require.Len(t, tokens, 1)

	urls.gotUserID = ""
	ctx := metadata.AppendToOutgoingContext(context.Background(), grpcserver.AuthMetadataKey, tokens[0])
	_, err = client.ListUserURLs(ctx, shortenerv1.UserURLsRequest_builder{}.Build())

	require.NoError(t, err)
	assert.Equal(t, testUserID, urls.gotUserID)
}
