package shortener

import (
	"context"
	"strconv"
	"testing"

	"Ustasjs/yp-url-shortener/internal/config/settings"
	"Ustasjs/yp-url-shortener/internal/model"
)

// noopStorage is a do-nothing Storage so service-level allocation/format tests
// run without a database.
type noopStorage struct{}

func (noopStorage) Save(context.Context, string, string, string) error { return nil }
func (noopStorage) Get(context.Context, string) (string, error)        { return "", nil }
func (noopStorage) SaveListUrls(context.Context, []model.ShortURLRecord, string) error {
	return nil
}
func (noopStorage) CreateUser(context.Context) (string, error) { return "", nil }
func (noopStorage) GetUserURLs(context.Context, string) ([]model.UserURLItem, error) {
	return nil, nil
}
func (noopStorage) DeleteURLsBatch(context.Context, []model.DeleteItem) error { return nil }

func TestCreateShortURLFormat(t *testing.T) {
	// Trailing slash in baseURL must not produce a double slash.
	for _, base := range []settings.BaseURL{"http://localhost:8080", "http://localhost:8080/"} {
		s := NewShortener(noopStorage{}, base, nil)
		got, err := s.CreateShortURL(context.Background(), "https://example.com", "u1")
		if err != nil {
			t.Fatalf("base=%q: unexpected error: %v", base, err)
		}
		want := "http://localhost:8080/" + shortenURL("https://example.com")
		if got != want {
			t.Errorf("base=%q: got %q, want %q", base, got, want)
		}
	}
}

func TestCreateShortURLsBatchFormat(t *testing.T) {
	s := NewShortener(noopStorage{}, "http://localhost:8080/", nil)
	items := []model.BatchShortURLRequestItem{
		{CorrelationID: "1", OriginalURL: "https://a.example"},
		{CorrelationID: "2", OriginalURL: "https://b.example"},
	}

	resp, err := s.CreateShortURLsBatch(context.Background(), items, "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != len(items) {
		t.Fatalf("got %d items, want %d", len(resp), len(items))
	}
	for i, item := range items {
		want := "http://localhost:8080/" + shortenURL(item.OriginalURL)
		if resp[i].ShortURL != want {
			t.Errorf("item %d: ShortURL = %q, want %q", i, resp[i].ShortURL, want)
		}
		if resp[i].CorrelationID != item.CorrelationID {
			t.Errorf("item %d: CorrelationID = %q, want %q", i, resp[i].CorrelationID, item.CorrelationID)
		}
	}
}

func BenchmarkCreateShortURLsBatch(b *testing.B) {
	s := NewShortener(noopStorage{}, settings.BaseURL("http://localhost:8080"), nil)
	items := make([]model.BatchShortURLRequestItem, 100)
	for i := range items {
		items[i] = model.BatchShortURLRequestItem{
			CorrelationID: strconv.Itoa(i),
			OriginalURL:   "https://example.com/articles/" + strconv.Itoa(i),
		}
	}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.CreateShortURLsBatch(ctx, items, "user"); err != nil {
			b.Fatal(err)
		}
	}
}
