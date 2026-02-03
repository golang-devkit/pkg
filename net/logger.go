package net

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-devkit/pkg/logger"

	"go.uber.org/zap"
)

var (
	// Header constants
	hostname = func() string {
		name, err := os.Hostname()
		if err != nil {
			name = "app"
		}
		// Append the process ID to the hostname for uniqueness
		return fmt.Sprintf("%s#%d", name, os.Getpid())
	}()
)

func valueDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func getLogEntry() *zap.Logger {
	return logger.NewEntry()
}

// getLoggerFromContext retrieves the logger from the context.
func getLoggerFromContext(ctx context.Context) *zap.Logger {
	return logger.GetLoggerFromContext(ctx)
}

func setLoggerToContext(ctx context.Context, entry *zap.Logger) context.Context {
	return logger.SetLoggerToContext(ctx, entry)
}

func apiLoggerHandler(h http.Handler) http.Handler {

	// Define the API logger
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get the logger from the context
		reqLogger := getLoggerFromContext(r.Context())

		// make a new response writer with the original writer
		wc := NewHttpWriter(w)
		defer func() {
			// Log the response body after the handler has processed the request
			reqLogger.Debug("API Logger",
				zap.ByteString(logger.KeyNetResponsePayload, wc.Body()),
				zap.String(logger.KeyNetResponseSize, wc.BodySize()))

			// Log the API request details
			printLogApi(wc, r, time.Now())
		}()

		// Clone the body with limit reader (10 MB)
		httpStatus, payload, err := CloneBodyWithLimitReader(r, 10*1024*1024)
		if err != nil {
			WriteError(wc, httpStatus, err)
			return
		}
		// Truncate the payload if it's too large
		if len(payload) > maxLoggedBodySize {
			payload = payload[:maxLoggedBodySize]
			payload = append(payload, []byte("...")...)
		}
		// Log the request body
		reqLogger.Debug("API Logger", zap.ByteString(logger.KeyNetRequestPayload, payload))

		// Call the next handler
		h.ServeHTTP(wc, r)
	})
}

func printLogApi(wc *ResponseWriter, r *http.Request, t time.Time) {

	// Collect response headers
	cH := wc.Header().Clone()

	// Check if there is an error message
	//if msg := cH.Get(xApiMoreError); msg != "" {
	//	more = fmt.Sprintf("%s > %s", more, msg)
	//}

	logger.NewEntry().Info("API request completed",
		zap.Time(logger.KeyTimestamp, time.Now()),
		zap.String(logger.KeyNetHostname, hostname),
		zap.String(logger.KeyNetRemoteAddr, r.RemoteAddr),
		zap.String(logger.KeyNetHttpMethod, r.Method),
		zap.String(logger.KeyNetHttpPath, r.URL.String()),
		zap.String(logger.KeyNetStatus, wc.Status()),
		zap.Int(logger.KeyNetStatusCode, wc.StatusCode()),
		zap.String(logger.KeyNetDuration, time.Since(t).String()),
		zap.String(logger.KeyNetClientID, r.Header.Get(xApiClientId)),
		zap.String(logger.KeyNetRequestID, r.Header.Get(xApiRequestId)),
		zap.String(logger.KeyNetOrigin, r.Header.Get(headerOrigin)),
		zap.String(logger.KeyNetUserAgent, r.Header.Get(headerUserAgent)),
		zap.String(logger.KeyNetDescription, cH.Get(xDescription)),
		zap.String(logger.KeyNetDescriptionError, cH.Get(xDescriptionError)),
	)
}
