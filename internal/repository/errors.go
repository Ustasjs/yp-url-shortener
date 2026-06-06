package repository

import "errors"

var ErrRecordNotFound = errors.New("record not found")
var ErrDeleted = errors.New("url is deleted")
