// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cmdutil

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
)

const (
	// MaxReaderLimitSize is the upper limit for reader capability.
	MaxReaderLimitSize = 250 * 1024 * 1024
	// ReaderTimeout is a time limit during the reader is waiting for data.
	ReaderTimeout = 1 * time.Minute
)

// Reader creates a reader instance according to the given name value
// Use "" or "-" for Stdin reader, else use a filename.
func Reader(name string) (io.Reader, error) {
	var (
		reader io.Reader
		err    error
	)

	// Create input reader
	switch name {
	case "", "-":
		// Check stdin
		info, errStat := os.Stdin.Stat()
		if errStat != nil {
			return nil, fmt.Errorf("unable to retrieve stdin information: %w", errStat)
		}
		if info.Mode()&os.ModeCharDevice != 0 {
			return nil, fmt.Errorf("the command expects stdin input but nothing seems readable")
		}

		// Stdin
		reader = bufio.NewReader(os.Stdin)
		reader = NewTimeoutReader(reader, ReaderTimeout)
	default:
		//nolint:gosec // G304: name is provided by caller for file reading
		reader, err = os.OpenFile(name, syscall.O_RDONLY, 0o400)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to build reader for read operations for %s: error: %w",
				name,
				err,
			)
		}
	}

	// Limit the reader
	limitedReader := &io.LimitedReader{R: reader, N: MaxReaderLimitSize}

	// No error
	return limitedReader, nil
}

// LineReader creates a reder and returns content read line by line.
func LineReader(name string) ([]string, error) {
	out := []string{}

	// Create input reader
	reader, err := Reader(name)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize a content reader: %w", err)
	}

	// Read line by line
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}

	// Check scanner error
	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("error occurs during scanner usage: %w", err)
	}

	// No error
	return out, nil
}

// Writer creates a writer according to the given name value
// Use "" or "-" for Stdout writer, else use a filename.
func Writer(name string) (io.Writer, error) {
	var (
		writer io.Writer
		err    error
	)

	// Create output writer
	switch name {
	case "", "-":
		// Stdout
		writer = os.Stdout
	default:
		// Open output file
		//nolint:gosec // G304: name is provided by caller for file writing
		writer, err = os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0o400)
		if err != nil {
			return nil, fmt.Errorf("unable to open '%s' for write: %w", name, err)
		}
	}

	// No error
	return writer, nil
}

// -----------------------------------------------------------------------------

// TimeoutReader implemnts a timelimited reader.
type TimeoutReader struct {
	reader  io.Reader
	timeout time.Duration
}

// NewTimeoutReader create a timed-out limited reader instance.
func NewTimeoutReader(reader io.Reader, timeout time.Duration) io.Reader {
	ret := new(TimeoutReader)
	ret.reader = reader
	ret.timeout = timeout
	return ret
}

// Read implements io.Reader interface.
func (r *TimeoutReader) Read(buf []byte) (n int, err error) {
	ch := make(chan bool, 1)
	n = 0
	err = nil
	go func() {
		n, err = r.reader.Read(buf)
		ch <- true
	}()
	select {
	case <-ch:
		return
	case <-time.After(r.timeout):
		return 0, errors.New("Reader timeout")
	}
}

// -----------------------------------------------------------------------------

// NewClosedWriter returns a io.WriteCloser instance which always fails when
// writing data. (Used for testing purpose).
func NewClosedWriter() io.WriteCloser {
	return &closedWriter{}
}

type closedWriter struct{}

func (c *closedWriter) Write(_ []byte) (int, error) {
	return 0, io.EOF
}

func (c *closedWriter) Close() error {
	return nil
}

// -----------------------------------------------------------------------------

// FileReader returns lazy evaluated reader.
func FileReader(filename string) func(context.Context) (io.Reader, error) {
	return func(_ context.Context) (io.Reader, error) {
		reader, err := Reader(filename)
		if err != nil {
			return nil, fmt.Errorf("unable to open file '%s' for reading: %w", filename, err)
		}

		// No error
		return reader, nil
	}
}

// FileWriter returns lazy evaluated writer.
func FileWriter(filename string) func(context.Context) (io.Writer, error) {
	return func(_ context.Context) (io.Writer, error) {
		writer, err := Writer(filename)
		if err != nil {
			return nil, fmt.Errorf("unable to open file '%s' for writing: %w", filename, err)
		}

		// No error
		return writer, nil
	}
}

// StdoutWriter returns lazy evaluated writer.
func StdoutWriter() func(context.Context) (io.Writer, error) {
	return func(_ context.Context) (io.Writer, error) {
		// No error
		return os.Stdout, nil
	}
}

// DiscardWriter returns discard writer.
func DiscardWriter() func(context.Context) (io.Writer, error) {
	return func(_ context.Context) (io.Writer, error) {
		// No error
		return io.Discard, nil
	}
}

// DirectWriter returns the given writer.
func DirectWriter(w io.Writer) func(context.Context) (io.Writer, error) {
	return func(_ context.Context) (io.Writer, error) {
		// No error
		return w, nil
	}
}
