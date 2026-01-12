package net

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
	status   int
	buffer   *bytes.Buffer
	bodySize int
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	// Limit the size of the logged body to 128KB
	if limit := 128 * 1024; len(b) > limit {
		io.Copy(rw.buffer, bytes.NewReader(b[:limit]))
		io.Copy(rw.buffer, bytes.NewReader([]byte("...")))
	} else {
		io.Copy(rw.buffer, bytes.NewReader(b))
	}
	// Write to the actual response writer
	n, err := rw.ResponseWriter.Write(b)
	// Track the size of the body written
	rw.bodySize += n
	return n, err
}

func (rw *ResponseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Hijack implements the http.Hijacker interface for WebSocket support
func (rw *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// --- Additional helper methods

func (rw *ResponseWriter) StatusCode() int {
	return rw.status
}

func (rw *ResponseWriter) Status() string {
	return http.StatusText(rw.status)
}

// Body returns the response a byte slice copy of the response body.
// Maximum size is 128KB plus ellipsis if truncated. This is useful for logging purposes.
func (rw *ResponseWriter) Body() []byte {
	return rw.buffer.Bytes()
}

// BodySize returns the size of the response body written.
// This is size of the actual body, not the logged body.
func (rw *ResponseWriter) BodySize() int {
	return rw.bodySize
}

func NewHttpWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{ResponseWriter: w, status: 200,
		buffer: bytes.NewBuffer(nil), bodySize: 0}
}
