// Package repository provides storage backends for shortened URLs: a
// PostgreSQL-backed store and an in-memory store with file persistence.
package repository

import "errors"

// ErrRecordNotFound is returned when no URL is stored under the requested short
// ID.
var ErrRecordNotFound = errors.New("record not found")

// ErrDeleted is returned when the requested short URL exists but has been marked
// as deleted.
var ErrDeleted = errors.New("url is deleted")
