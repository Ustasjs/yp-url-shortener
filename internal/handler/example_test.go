package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"Ustasjs/yp-url-shortener/internal/handler"
	"Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/repository"
	"Ustasjs/yp-url-shortener/internal/service/shortener"
)

const exampleBaseURL = "http://localhost:8080"

func newExampleShortener() *shortener.Shortener {
	f, err := os.CreateTemp("", "example-urls-*.json")
	if err != nil {
		panic(err)
	}
	_, _ = f.WriteString("{}")
	_ = f.Close()

	store := repository.NewMemStorage(f.Name())
	return shortener.NewShortener(store, exampleBaseURL, nil)
}

// ExampleHandler_CreateShortURL shortens a URL through POST / with a plain-text
// body. The response is the short URL as text/plain with status 201 Created.
func ExampleHandler_CreateShortURL() {
	h := handler.NewHandler(newExampleShortener(), nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/"))
	rec := httptest.NewRecorder()

	h.CreateShortURL(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
	// Output:
	// 201
	// http://localhost:8080/sEjv7Ay3
}

// ExampleHandler_CreateShortURLJSONApi shortens a URL through POST /api/shorten
// with a JSON body. The response is a JSON object holding the short URL.
func ExampleHandler_CreateShortURLJSONApi() {
	h := handler.NewHandler(newExampleShortener(), nil, nil)

	body := strings.NewReader(`{"url":"https://practicum.yandex.ru/"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
	rec := httptest.NewRecorder()

	h.CreateShortURLJSONApi(rec, req)

	fmt.Println(rec.Code)
	fmt.Print(rec.Body.String())
	// Output:
	// 201
	// {"result":"http://localhost:8080/sEjv7Ay3"}
}

// ExampleHandler_CreateShortURLSByBatch shortens several URLs in one call through
// POST /api/shorten/batch. Each response item carries the caller's
// correlation_id so requests and responses can be matched.
func ExampleHandler_CreateShortURLSByBatch() {
	h := handler.NewHandler(newExampleShortener(), nil, nil)

	body := strings.NewReader(`[{"correlation_id":"1","original_url":"https://practicum.yandex.ru/"},` +
		`{"correlation_id":"2","original_url":"https://go.dev/"}]`)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", body)
	rec := httptest.NewRecorder()

	h.CreateShortURLSByBatch(rec, req)

	fmt.Println(rec.Code)
	fmt.Print(rec.Body.String())
	// Output:
	// 201
	// [{"correlation_id":"1","short_url":"http://localhost:8080/sEjv7Ay3"},{"correlation_id":"2","short_url":"http://localhost:8080/gAcFuKhm"}]
}

// ExampleHandler_GetShortURLByID resolves a short URL through GET /{id} and
// responds with a 307 Temporary Redirect to the original URL.
func ExampleHandler_GetShortURLByID() {
	svc := newExampleShortener()

	// Seed a short URL so there is something to resolve.
	shortURL, _ := svc.CreateShortURL(context.Background(), "https://practicum.yandex.ru/", "")
	id := strings.TrimPrefix(shortURL, exampleBaseURL+"/")

	h := handler.NewHandler(svc, nil, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("/{id}", h.GetShortURLByID)

	req := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Header().Get("Location"))
	// Output:
	// 307
	// https://practicum.yandex.ru/
}

// ExampleHandler_GetUserURLs lists every URL created by the authenticated user
// through GET /api/user/urls. The user identity is normally injected by the auth
// middleware; here it is placed in the request context directly.
func ExampleHandler_GetUserURLs() {
	const userID = "user-1"

	svc := newExampleShortener()
	_, _ = svc.CreateShortURL(context.Background(), "https://practicum.yandex.ru/", userID)

	h := handler.NewHandler(svc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDContextKey, userID))
	rec := httptest.NewRecorder()

	h.GetUserURLs(rec, req)

	fmt.Println(rec.Code)
	fmt.Print(rec.Body.String())
	// Output:
	// 200
	// [{"short_url":"http://localhost:8080/sEjv7Ay3","original_url":"https://practicum.yandex.ru/"}]
}

// ExampleHandler_DeleteUserURLs schedules asynchronous deletion of the user's
// short URLs through DELETE /api/user/urls and immediately returns 202 Accepted.
func ExampleHandler_DeleteUserURLs() {
	h := handler.NewHandler(newExampleShortener(), nil, nil)

	body := strings.NewReader(`["abc12345","def67890"]`)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", body)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDContextKey, "user-1"))
	rec := httptest.NewRecorder()

	h.DeleteUserURLs(rec, req)

	fmt.Println(rec.Code)
	// Output:
	// 202
}

// ExampleHandler_GetDBPing reports the availability of the configured database
// through GET /ping, responding with 200 OK when the database is reachable.
func ExampleHandler_GetDBPing() {
	h := handler.NewHandler(nil, newMockPingerOk(), nil)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	h.GetDBPing(rec, req)

	fmt.Println(rec.Code)
	// Output:
	// 200
}
