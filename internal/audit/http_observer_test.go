package audit_test

import (
	"Ustasjs/yp-url-shortener/internal/audit"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPObserver_PostsEventAsJSON(t *testing.T) {
	var (
		mu        sync.Mutex
		gotMethod string
		gotCT     string
		gotEvent  audit.Event
		received  = make(chan struct{}, 1)
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		mu.Lock()
		gotMethod = r.Method
		gotCT = r.Header.Get("Content-Type")
		_ = json.Unmarshal(body, &gotEvent)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		select {
		case received <- struct{}{}:
		default:
		}
	}))
	defer srv.Close()

	observer := audit.NewHTTPObserver(srv.URL)
	event := audit.Event{Timestamp: 42, Action: audit.ActionShorten, UserID: "u1", URL: "https://example.com"}
	observer.Notify(event)

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatal("http sink did not receive event")
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "application/json", gotCT)
	assert.Equal(t, event, gotEvent)
}

func TestHTTPObserver_BadStatusDoesNotPanic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	observer := audit.NewHTTPObserver(srv.URL)
	require.NotPanics(t, func() {
		observer.Notify(audit.Event{Action: audit.ActionShorten, URL: "https://example.com"})
	})
}

func TestHTTPObserver_UnreachableDoesNotPanic(t *testing.T) {
	observer := audit.NewHTTPObserver("http://127.0.0.1:1") // closed port
	require.NotPanics(t, func() {
		observer.Notify(audit.Event{Action: audit.ActionFollow, URL: "https://example.com"})
	})
}
