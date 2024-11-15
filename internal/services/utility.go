package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func ShowQuery(q Querier) {
	fmt.Println(q.String())
}

func ShowStorage(s MetricPrinter) {
	fmt.Println(s.String(context.TODO()))
}

/*
compress takes a byte slice as input and returns the original byte slice compressed with the gzip algorithm and an error

Args:

	data []byte: input byte slice

Returns:

	[]byte: compressed byte slyce
	error: nil or error, if occured
*/
func compress(data []byte) ([]byte, error) {
	var b bytes.Buffer

	w := gzip.NewWriter(&b)

	_, err := w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}

	return b.Bytes(), nil

}

/*
decompress takes a compressed byte slice as input and returns the original byte slice uncompressed with the gzip algorithm and an error

Args:

	data []byte: input compressed byte slice

Returns:

	[]byte: decompressed byte slyce
	error: nil or error, if occured
*/
func decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed read data from compressed buffer: %v", err)
	}
	defer r.Close()

	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return b.Bytes(), nil
}

/*
retryableError checks the error type and decides whether to retry the request

Args:

	err error: a checkable error

Returns:

	bool: true - do retry, false - don't retry
*/
func retryableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.DeadlockDetected, pgerrcode.LockNotAvailable, pgerrcode.ConnectionException, pgerrcode.ConnectionFailure, pgerrcode.SQLClientUnableToEstablishSQLConnection:
			return true
		}
	}
	return false
}

/*
retry a wrapper function for retrying functions of the form `func() error`. Performs 3 retries with delays of 1, 3, and 5 seconds

Args:

	operation func() error: function for retrying

Returns:

	error
*/
func retry(operation func() error) error {
	for i := 0; i < 3; i++ {
		err := operation()
		if err == nil {
			return nil
		}

		if retryableError(err) {
			time.Sleep(time.Second * time.Duration(2*i+1))
		} else {
			return err
		}
	}
	return fmt.Errorf("operation failed after 3 attempts")
}
