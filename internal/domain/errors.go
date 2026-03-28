package domain

import "errors"

var (
	ErrNotFound      = errors.New("qr code not found")
	ErrInvalidURL    = errors.New("invalid url")
	ErrURLTooLong    = errors.New("url exceeds maximum length of 2048 characters")
	ErrTokenConflict = errors.New("token collision, please retry")
)
