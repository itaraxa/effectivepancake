package services

import "context"

type storagChecker interface {
	PingContext(context.Context) error
}

func CheckConnectionDB(ctx context.Context, s storagChecker) error {
	return nil
}
