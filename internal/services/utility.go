package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
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
