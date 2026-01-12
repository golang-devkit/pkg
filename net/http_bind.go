package net

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

var (
	firstOf = func(args []string) string {
		if len(args) > 0 {
			return args[0]
		}
		return ""
	}
)

func GetHeader(r *http.Request, key string) string {
	return r.Header.Get(key)
}

func QueryParams(r *http.Request, key string, def ...string) string {

	if val := r.URL.Query().Get(key); val != "" {
		return val
	}
	return firstOf(def)
}

func GetQueryParams(r *http.Request, key string) []string {
	return r.URL.Query()[key]
}

func GetRawData(r *http.Request) ([]byte, error) {

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			println(err.Error())
		}
	}(r.Body)

	return io.ReadAll(r.Body)
}

func CloneBodyWithLimitReader(r *http.Request, limit int64) (httpStatusCode int, raw []byte, err error) {
	// Check Content-Length header first
	contentLength, err := strconv.ParseInt(r.Header.Get(headerContentLength), 10, 64)
	if err != nil {
		// Silently ignore error and continue to read body
		// fmt.Printf("can not get content-length header of %s\n", r.URL.String())
		// return http.StatusLengthRequired, nil, fmt.Errorf("can not get content-length header")
	} else if contentLength > limit {
		return http.StatusRequestURITooLong, nil, fmt.Errorf("request body too large: exceeds %d bytes, got %d bytes", limit, contentLength)
	}
	// Continue to read body if content-length is acceptable
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			println(err.Error())
		}
		// Restore the original body for further reading
		if len(raw) > 0 {
			r.Body = io.NopCloser(bytes.NewReader(raw))
		}
	}(r.Body)
	// Read body with limit reader
	raw, err = io.ReadAll(io.LimitReader(r.Body, limit))
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	return http.StatusOK, raw, nil
}

func ShouldBindJSON(r *http.Request, v interface{}) (raw []byte, err error) {

	raw, err = GetRawData(r)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(raw, v); err != nil {
		return raw, err
	}

	return raw, nil
}
