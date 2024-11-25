package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func ShowQuery(q Querier) string {
	return q.String()
}

func ShowStorage(s MetricPrinter) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.String(ctx)
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
retryablePgError checks the error type and decides whether to retry the request

Args:

	err error: a checkable error

Returns:

	bool: true - do retry, false - don't retry
*/
func retryablePgError(err error) bool {
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
retryQueryToDB a wrapper function for retrying functions of the form `func() error`. Performs 3 retries with delays of 1, 3, and 5 seconds

Args:

	operation func() error: function for retrying

Returns:

	error
*/
func retryQueryToDB(operation func() error) error {
	for i := 0; i < 3; i++ {
		err := operation()
		if err == nil {
			return nil
		}

		if retryablePgError(err) {
			time.Sleep(time.Second * time.Duration(2*i+1))
		} else {
			return err
		}
	}
	return fmt.Errorf("operation failed after 3 attempts")
}

func retryRequest(operation func() (*http.Response, error)) (*http.Response, error) {
	for i := 0; i < 3; i++ {
		resp, err := operation()
		if err == nil {
			return resp, nil
		}

		time.Sleep(time.Second * time.Duration(2*i+1))
	}
	return nil, fmt.Errorf("operation failed after 3 attempts")
}

func createURL(serverURL string, p ...string) string {

	if strings.Contains(serverURL, `http://`) {
		return serverURL + `/` + strings.Join(p, `/`)
	} else {
		return `http://` + serverURL + `/` + strings.Join(p, `/`)
	}
}

/*
CheckSign compares two digital signature values in string format

Args:

	l logger
	expectedSign string: expectd digital signature
	actualSign string: actual digital signature

Returns:

	bool: result of an exact comparison
*/
func CheckSign(l logger, expectedSign string, actualSign string) bool {
	l.Debug("check digit signatures", "expected", expectedSign, "actual", actualSign)
	if expectedSign != actualSign {
		l.Error("check signatures", "error", fmt.Sprintf("expected signature %s not equal actual signature %s", expectedSign, actualSign))
		return false
	}
	l.Info("check signatures - success")
	return true
}

/*
SignSHA256 calculates an SHA-256 digital signature

Args:

	l logger
	body io.Reader: the data to be signed
	key string: the key used for signing

Returns:

	string: SHA-256 digital signature
	error
*/
func SignSHA256(l logger, body io.Reader, key string) (string, error) {
	if key == `` {
		l.Debug(`empty key value is set`)
	}

	h := hmac.New(sha256.New, []byte(key))
	var data []byte
	d := 0
	buffer := make([]byte, 1024)
	for {
		n, err := body.Read(buffer)
		if err != nil && err != io.EOF {
			return ``, fmt.Errorf("sha256 sign error: %w", err)
		}
		d += n
		data = append(data, buffer[:n]...)
		if err == io.EOF {
			break
		}
	}

	if d == 0 {
		return ``, fmt.Errorf("sha256 sign error: empty body input")
	}
	_, err := h.Write(data)
	if err != nil {
		return ``, fmt.Errorf("sha256 sign error: %w", err)
	}
	l.Debug(`sha256 sign`, `body`, fmt.Sprintf("%x", data), `key`, key, `sha256sum`, fmt.Sprintf("%x", h.Sum(nil)))

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
