package grpcserver

import (
	"context"
	"errors"

	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/repository"
	"Ustasjs/yp-url-shortener/internal/service/urlservice"
	shortenerv1 "Ustasjs/yp-url-shortener/pkg/proto/shortener/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ShortURLHeader carries the existing short URL when ShortenURL fails with
// AlreadyExists. HTTP puts it into the body of the 409 response.
const ShortURLHeader = "x-short-url"

// ShortenerServer implements the ShortenerService RPCs on top of URLService.
type ShortenerServer struct {
	shortenerv1.UnimplementedShortenerServiceServer

	urls URLService
}

// ShortenURL saves the URL and returns its short form.
func (s *ShortenerServer) ShortenURL(ctx context.Context, req *shortenerv1.URLShortenRequest) (*shortenerv1.URLShortenResponse, error) {
	userID, _ := middleware.GetUserIDFromContext(ctx)

	shortURL, err := s.urls.Shorten(ctx, req.GetUrl(), userID)
	switch {
	case errors.Is(err, urlservice.ErrInvalidURL):
		return nil, status.Error(codes.InvalidArgument, "invalid url")
	case errors.Is(err, repository.ErrConflict):
		return nil, s.conflictError(ctx, shortURL)
	case err != nil:
		logger.Log.Error("grpc: failed to create short url", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot create short url")
	}

	return &shortenerv1.URLShortenResponse{Result: shortURL}, nil
}

// ExpandURL returns the original URL behind a short id.
func (s *ShortenerServer) ExpandURL(ctx context.Context, req *shortenerv1.URLExpandRequest) (*shortenerv1.URLExpandResponse, error) {
	userID, _ := middleware.GetUserIDFromContext(ctx)

	originalURL, err := s.urls.Expand(ctx, req.GetId(), userID)
	switch {
	case errors.Is(err, repository.ErrDeleted):
		return nil, status.Error(codes.FailedPrecondition, "url is deleted")
	case err != nil:
		return nil, status.Error(codes.NotFound, "url not found")
	}

	return &shortenerv1.URLExpandResponse{Result: originalURL}, nil
}

// ListUserURLs returns all URLs of the calling user.
func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *shortenerv1.UserURLsRequest) (*shortenerv1.UserURLsResponse, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	items, err := s.urls.ListUserURLs(ctx, userID)
	if err != nil {
		logger.Log.Error("grpc: failed to list user urls", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	// HTTP answers 204 No Content for a user without URLs. gRPC has nothing like
	// that, so an empty list with an OK status is the closest match.
	response := &shortenerv1.UserURLsResponse{Url: make([]*shortenerv1.URLData, 0, len(items))}
	for _, item := range items {
		response.Url = append(response.Url, &shortenerv1.URLData{
			ShortUrl:    item.ShortURL,
			OriginalUrl: item.OriginalURL,
		})
	}
	return response, nil
}

// conflictError builds the AlreadyExists error for a URL that was shortened
// before. A gRPC handler that returns an error makes grpc-go drop the response
// message, so the short URL cannot ride in URLShortenResponse the way it rides
// in the body of the HTTP 409. The x-short-url response header carries it
// instead.
func (s *ShortenerServer) conflictError(ctx context.Context, shortURL string) error {
	if err := grpc.SetHeader(ctx, metadata.Pairs(ShortURLHeader, shortURL)); err != nil {
		logger.Log.Error("grpc: failed to set short url header", zap.Error(err))
	}
	return status.Error(codes.AlreadyExists, "url already exists")
}
