package services

import (
	"context"
	"time"
)

type storagChecker interface {
	PingContext(context.Context) error
}

/*
CheckConnectionStorage check access to storage

Args:

	ctx context.Context
	l logger
	s storagChecker

Returns:

	error
*/
func CheckConnectionStorage(ctx context.Context, l logger, s storagChecker) error {
	start := time.Now()
	err := s.PingContext(ctx)
	if err != nil {
		l.Debug("checking connection to DB storage failed", "error", err.Error())
		return err
	}
	l.Debug("checking connection to DB success", "duration", time.Since(start))
	return nil
}
