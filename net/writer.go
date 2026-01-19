package net

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

const (
	maxLoggedBodySize = 256  // Maximum size of the logged body in bytes
	maxLoggedJsonSize = 1024 // Maximum size of the logged JSON in bytes
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
	// Limit the size of the logged body to 256 bytes
	if len(b) > maxLoggedBodySize {
		io.Copy(rw.buffer, bytes.NewReader(b[:maxLoggedBodySize]))
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
// Maximum size is 256 bytes plus ellipsis if truncated. This is useful for logging purposes.
func (rw *ResponseWriter) Body() []byte {
	// Check the Content-Type for special handling
	ctype := rw.Header().Get("Content-Type")
	// Handle JSON content type separately
	if strings.HasPrefix(ctype, "application/json") {
		return fmt.Appendf(rw.buffer.Next(maxLoggedJsonSize), "...")
	}
	if strings.HasPrefix(ctype, "image/jpeg") || strings.HasPrefix(ctype, "image/png") ||
		strings.HasPrefix(ctype, "image/gif") || strings.HasPrefix(ctype, "image/svg") {
		return fmt.Appendf(nil, "<<%s data>>", ctype)
	}
	if strings.HasPrefix(ctype, "application/pdf") || strings.HasPrefix(ctype, "application/vnd.") {
		return fmt.Appendf(nil, "<<%s data>>", ctype)
	}
	if strings.HasPrefix(ctype, "text/csv") {
		return fmt.Appendf(nil, "<<%s data>>", ctype)
	}
	if strings.HasPrefix(ctype, "application/octet-stream") || strings.HasPrefix(ctype, "application/pdf") ||
		strings.HasPrefix(ctype, "application/zip") {
		return fmt.Appendf(nil, "<<%s data>>", ctype)
	}
	return fmt.Appendf(rw.buffer.Next(maxLoggedBodySize), "...")
}

// BodySize returns the size of the response body written.
// This is size of the actual body, not the logged body.
func (rw *ResponseWriter) BodySize() string {
	if rw.bodySize < 1<<10 {
		return fmt.Sprintf("%d bytes", rw.bodySize) // in bytes
	}
	if rw.bodySize < 1<<20 {
		return fmt.Sprintf("%.3f KB", float64(rw.bodySize)/(1<<10)) // in KB
	}
	if rw.bodySize < 1<<30 {
		return fmt.Sprintf("%.3f MB", float64(rw.bodySize)/(1<<20)) // in MB
	}
	return fmt.Sprintf("%.3f GB", float64(rw.bodySize)/(1<<30)) // in GB
}

func NewHttpWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{ResponseWriter: w, status: 200,
		buffer: bytes.NewBuffer(nil), bodySize: 0}
}
