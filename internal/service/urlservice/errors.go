package urlservice

import "errors"

// ErrInvalidURL means the caller passed a string that is not an absolute URL
// with a scheme and a host.
var ErrInvalidURL = errors.New("invalid url")
