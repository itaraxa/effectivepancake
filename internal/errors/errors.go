package errors

import "errors"

var (
	// Server errors
	ErrParseGauge   = errors.New("gauge value parsing error")
	ErrParseCounter = errors.New("counter value parsing error")
	ErrBadType      = errors.New("bad type in raw query string")
	ErrBadName      = errors.New("bad name in raw query string")
	ErrBadValue     = errors.New("bad value in raw query string")
	ErrBadRawQuery  = errors.New("bad raw query string")

	// Agent errors
	ErrRequestCreating = errors.New("creating request error")
	ErrRequestSending  = errors.New("sending request error")
)
