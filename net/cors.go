package net

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-devkit/pkg/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// CORS enabled by default
	enableCORS = true
)

func setCORSEnabled(enabled bool) {
	enableCORS = enabled
}

func writeHeadersCORS(w http.ResponseWriter) {
	// Set CORS headers
	w.Header().Set(corsAllowOriginHeader, "*")
	w.Header().Set(corsAllowMethodsHeader, "GET, POST, OPTIONS")
	w.Header().Set(corsAllowHeadersHeader, "*")
	w.Header().Set(corsMaxAgeHeader, "3600")
	w.Header().Set(corsAllowCredentialsHeader, "true")
}

// Define CORS options
func corsHandler(h http.Handler) http.Handler {
	if !enableCORS {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			// Set CORS headers
			writeHeadersCORS(w)
			// Respond with 204 No Content
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Set CORS headers
		writeHeadersCORS(w)
		// Call the next handler
		h.ServeHTTP(w, r)
	})
}

func loggerIntercepter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create logger with request context
		reqLogger := getLogEntry().With(
			zap.String(logger.KeyNetHttpMethod, r.Method),
			zap.String(logger.KeyNetHttpPath, r.URL.Path),
			zap.String(logger.KeyNetHttpQuery, r.URL.RawQuery),
			zap.String(logger.KeyNetRequestID, r.Header.Get(xApiRequestId)),
			zap.Object(logger.KeyNetRequestHeaders,
				zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
					for k, v := range r.Header {
						enc.AddString(k, strings.Join(v, "; "))
					}
					return nil
				})),
		)
		// Use the context with the logger
		rc := r.WithContext(setLoggerToContext(r.Context(), reqLogger))

		// Call the next handler with the new context
		h.ServeHTTP(w, rc)
	})
}

func customizeMethodNotAllowedHandler() http.Handler {
	//
	fmt.Printf("[DEBUG] Customize MethodNotAllowedHandler default to allow method OPTIONS\n")
	// Return 403 for Method Not Allowed
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//  for OPTIONS, we have handled it in CORS middleware
		if r.Method == http.MethodOptions {
			// Set CORS headers
			writeHeadersCORS(w)
			// Respond with 204 No Content
			w.WriteHeader(http.StatusNoContent)
			return
		}
		fmt.Printf("[DEBUG] 403 Not Found: %s %s\n", r.Method, r.URL.Path)
		http.Error(w, "403 method not allowed", http.StatusMethodNotAllowed)
	})
}

func Middleware(ro *mux.Router, enableCORS bool, middlewareFunc ...http.HandlerFunc) http.Handler {

	// Apply the middleware functions
	middlewareHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					getLogEntry().Error("Panic recovered in middleware",
						zap.Any("recover", rec),
					)
					http.Error(w, "500 internal server error", http.StatusInternalServerError)
				}
			}()
			// Set the request-Id
			if requestId := r.Header.Get(xApiRequestId); requestId == "" {
				r.Header.Set(xApiRequestId, uuid.NewString())
			}
			for _, Func := range middlewareFunc {
				Func(w, r)
			}
			h.ServeHTTP(w, r)
		})
	}

	// Set CORS global flag
	if setCORSEnabled(enableCORS); enableCORS {
		fmt.Printf("[DEBUG] CORS is enabled\n")
	} else {
		fmt.Printf("[DEBUG] CORS is disabled\n")
	}

	// Replace the default MethodNotAllowedHandler
	ro.MethodNotAllowedHandler = customizeMethodNotAllowedHandler()

	// Apply the middleware in order
	ro.Use(corsHandler, middlewareHandler, loggerIntercepter, apiLoggerHandler)

	// Walk through all the registered routes
	if err := ro.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		entry := getLogEntry()
		methods, err := route.GetMethods()
		if err != nil {
			// fmt.Printf("[%-8s] HTTP/1.1 %s\n", "", pathTemplate)
			entry.Debug("HTTP/1.1",
				zap.String("http_method", "ANY"),
				zap.String("path", pathTemplate),
			)
			return nil
		}
		for _, method := range methods {
			// fmt.Printf("[%-8s] HTTP/1.1 %s\n", method, pathTemplate)
			entry.Debug("HTTP/1.1",
				zap.String("http_method", method),
				zap.String("path", pathTemplate),
			)
		}
		return nil
	}); err != nil {
		fmt.Println("Error walking routes: ", err)
	}

	return ro
}
