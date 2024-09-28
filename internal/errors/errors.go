package errors

import "errors"

var (
	ErrParseGauge   = errors.New("gauge value parsing error")
	ErrParseCounter = errors.New("counter value parsing error")
	ErrBadType      = errors.New("bad type in raw query string")
	ErrBadName      = errors.New("bad name in raw query string")
	ErrBadValue     = errors.New("bad value in raw query string")
	ErrBadRawQuery  = errors.New("bad raw query string")
)
