package urlservice_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"Ustasjs/yp-url-shortener/internal/audit"
	"Ustasjs/yp-url-shortener/internal/model"
	"Ustasjs/yp-url-shortener/internal/repository"
	"Ustasjs/yp-url-shortener/internal/service/urlservice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testShortURL = "http://localhost:8080/test-id"
	testUserID   = "user-1"
)

var errStorage = errors.New("storage is down")

type mockShortener struct {
	createErr error
	original  string
	getErr    error
	items     []model.UserURLItem
	listErr   error

	createCalls  int
	gotOriginal  string
	gotCreateUID string
}

func (m *mockShortener) CreateShortURL(_ context.Context, originalURL string, userID string) (string, error) {
	m.createCalls++
	m.gotOriginal = originalURL
	m.gotCreateUID = userID
	if m.createErr != nil && !errors.Is(m.createErr, repository.ErrConflict) {
		return "", m.createErr
	}
	return testShortURL, m.createErr
}

func (m *mockShortener) GetOriginalURL(_ context.Context, _ string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}
	return m.original, nil
}

func (m *mockShortener) GetUserURLs(_ context.Context, _ string) ([]model.UserURLItem, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.items, nil
}

type capturingAuditor struct {
	mu     sync.Mutex
	events []audit.Event
}

func (c *capturingAuditor) Publish(event audit.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event)
}

func (c *capturingAuditor) snapshot() []audit.Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]audit.Event, len(c.events))
	copy(out, c.events)
	return out
}

func TestService_Shorten(t *testing.T) {
	tests := []struct {
		name        string
		rawURL      string
		createErr   error
		wantErr     error
		wantURL     string
		wantStored  string
		wantCalls   int
		wantActions []string
	}{
		{
			name:        "success",
			rawURL:      "https://practicum.yandex.ru",
			wantURL:     testShortURL,
			wantStored:  "https://practicum.yandex.ru",
			wantCalls:   1,
			wantActions: []string{audit.ActionShorten},
		},
		{
			name:        "conflict still returns the short url",
			rawURL:      "https://practicum.yandex.ru",
			createErr:   repository.ErrConflict,
			wantErr:     repository.ErrConflict,
			wantURL:     testShortURL,
			wantStored:  "https://practicum.yandex.ru",
			wantCalls:   1,
			wantActions: []string{audit.ActionShorten},
		},
		{
			name:      "storage error",
			rawURL:    "https://practicum.yandex.ru",
			createErr: errStorage,
			wantErr:   errStorage,
			wantCalls: 1,
		},
		{
			name:    "not an absolute url",
			rawURL:  "practicum.yandex.ru",
			wantErr: urlservice.ErrInvalidURL,
		},
		{
			name:    "empty url",
			rawURL:  "",
			wantErr: urlservice.ErrInvalidURL,
		},
		{
			name:    "no host",
			rawURL:  "https://",
			wantErr: urlservice.ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortener := &mockShortener{createErr: tt.createErr}
			auditor := &capturingAuditor{}
			svc := urlservice.New(shortener, auditor)

			shortURL, err := svc.Shorten(context.Background(), tt.rawURL, testUserID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantURL, shortURL)
			assert.Equal(t, tt.wantCalls, shortener.createCalls)
			if tt.wantStored != "" {
				assert.Equal(t, tt.wantStored, shortener.gotOriginal)
				assert.Equal(t, testUserID, shortener.gotCreateUID)
			}

			events := auditor.snapshot()
			require.Len(t, events, len(tt.wantActions))
			for i, action := range tt.wantActions {
				assert.Equal(t, action, events[i].Action)
				assert.Equal(t, testUserID, events[i].UserID)
				assert.Equal(t, tt.wantStored, events[i].URL)
				assert.NotZero(t, events[i].Timestamp)
			}
		})
	}
}

// Shorten saves the normalized form, because later lookups and audit records
// use it.
func TestService_ShortenNormalizesURL(t *testing.T) {
	shortener := &mockShortener{}
	auditor := &capturingAuditor{}
	svc := urlservice.New(shortener, auditor)

	_, err := svc.Shorten(context.Background(), "https://practicum.yandex.ru/path?b=2&a=1", testUserID)

	require.NoError(t, err)
	assert.Equal(t, "https://practicum.yandex.ru/path?b=2&a=1", shortener.gotOriginal)
	assert.Equal(t, shortener.gotOriginal, auditor.snapshot()[0].URL)
}

func TestService_Expand(t *testing.T) {
	tests := []struct {
		name       string
		original   string
		getErr     error
		wantErr    error
		wantURL    string
		wantEvents int
	}{
		{
			name:       "success",
			original:   "https://practicum.yandex.ru",
			wantURL:    "https://practicum.yandex.ru",
			wantEvents: 1,
		},
		{
			name:    "not found",
			getErr:  repository.ErrRecordNotFound,
			wantErr: repository.ErrRecordNotFound,
		},
		{
			name:    "deleted",
			getErr:  repository.ErrDeleted,
			wantErr: repository.ErrDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortener := &mockShortener{original: tt.original, getErr: tt.getErr}
			auditor := &capturingAuditor{}
			svc := urlservice.New(shortener, auditor)

			originalURL, err := svc.Expand(context.Background(), "test-id", testUserID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantURL, originalURL)

			events := auditor.snapshot()
			require.Len(t, events, tt.wantEvents)
			if tt.wantEvents > 0 {
				assert.Equal(t, audit.ActionFollow, events[0].Action)
				assert.Equal(t, testUserID, events[0].UserID)
				assert.Equal(t, tt.original, events[0].URL)
				assert.NotZero(t, events[0].Timestamp)
			}
		})
	}
}

func TestService_ListUserURLs(t *testing.T) {
	items := []model.UserURLItem{
		{ShortURL: testShortURL, OriginalURL: "https://practicum.yandex.ru"},
	}

	t.Run("success", func(t *testing.T) {
		svc := urlservice.New(&mockShortener{items: items}, &capturingAuditor{})

		got, err := svc.ListUserURLs(context.Background(), testUserID)

		require.NoError(t, err)
		assert.Equal(t, items, got)
	})

	t.Run("error", func(t *testing.T) {
		svc := urlservice.New(&mockShortener{listErr: errStorage}, &capturingAuditor{})

		got, err := svc.ListUserURLs(context.Background(), testUserID)

		require.ErrorIs(t, err, errStorage)
		assert.Nil(t, got)
	})
}

func TestService_NilAuditor(t *testing.T) {
	svc := urlservice.New(&mockShortener{original: "https://practicum.yandex.ru"}, nil)

	shortURL, err := svc.Shorten(context.Background(), "https://practicum.yandex.ru", testUserID)
	require.NoError(t, err)
	assert.Equal(t, testShortURL, shortURL)

	originalURL, err := svc.Expand(context.Background(), "test-id", testUserID)
	require.NoError(t, err)
	assert.Equal(t, "https://practicum.yandex.ru", originalURL)
}
