package net

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		fmt.Printf("Stream started: %v\n", info.FullMethod)
		if err := handler(srv, ss); err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return err
		}
		fmt.Printf("Stream completed successfully\n")
		return nil
	}
}

// UnaryServerAuthInterceptor creates a server interceptor for attack middleware function to gRPC requests
func UnaryServerAuthInterceptor(authFunc func(fullMethod string, bodyHash string, jwtStr string) error) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		var (
			startTime = time.Now()
			bodyHash  string
		)
		// Parse metadata, jwt token and request ID
		md, jwtAuthStr, reqID := metadataFromContext(ctx, req)

		// Create logger with request context
		reqLogger := getLogEntry().With(
			zap.Bool("proto_marshaled", false),
			zap.String("method", info.FullMethod),
			zap.String("req_id", reqID),
			zap.Any("metadata", md),
		)
		// Use the context with the logger
		ctx = setLoggerToContext(ctx, reqLogger)

		msg, ok := req.(proto.Message)
		if ok {
			// Marshal the proto message to log its SHA256 hash
			b, err := proto.Marshal(msg)
			if err != nil {
				reqLogger = reqLogger.With(
					zap.Any("proto_message", msg), // If the request is a proto message, log it
					zap.Errors("marshal_error", []error{err}),
				)
			} else {
				sum := sha256.Sum256(b)
				bodyHash = hex.EncodeToString(sum[:])
				reqLogger = reqLogger.With(
					zap.Bool("proto_marshaled", true),
					zap.String("sum", hex.EncodeToString(sum[:])),
					zap.String("proto_message", string(b)), // If the request is a proto message, log it
				)
			}
		} else {
			// Otherwise, log the request as a generic interface
			reqLogger = reqLogger.With(
				zap.Any("request", req),
				zap.Errors("proto_marshal_error", []error{status.Errorf(codes.Internal, "request is not a proto message")}),
			)
		}

		if err := authFunc(info.FullMethod, bodyHash, jwtAuthStr); err != nil {
			reqLogger.Error("Authorization failed",
				zap.String("body_hash", bodyHash),
				zap.String("jwt", jwtAuthStr),
				zap.Error(err))
			return nil, status.Errorf(codes.Unauthenticated, "Authorization failed: %v", err)
		}

		// Log the request start
		reqLogger.Info("gRPC request started",
			zap.Any("request", req),
			zap.Time("start_time", startTime))

		// Process the request
		resp, err := handler(ctx, req)

		// Get status code
		statusCode := codes.OK
		if err != nil {
			statusCode = status.Code(err)
		}

		// Log completion
		reqLogger.Info("gRPC request completed",
			zap.Any("response", resp),
			zap.String("status", statusCode.String()),
			zap.Duration("duration", time.Since(startTime)),
			zap.Error(err),
		)

		return resp, err
	}
}

// UnaryServerLoggingInterceptor creates a server interceptor for logging gRPC requests
func UnaryServerLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		// Parse metadata and request ID
		md, _, reqID := metadataFromContext(ctx, req)
		startTime := time.Now()

		// Create logger with request context
		reqLogger := getLogEntry().With(
			zap.Bool("proto_marshaled", false),
			zap.String("method", info.FullMethod),
			zap.String("req_id", reqID),
			zap.Any("metadata", md),
		)
		// Use the context with the logger
		ctx = setLoggerToContext(ctx, reqLogger)

		// Log the request start
		reqLogger.Info("gRPC request started",
			zap.Any("request", req),
			zap.Time("start_time", startTime),
		)

		// Process the request
		resp, err := handler(ctx, req)

		// Get status code
		statusCode := codes.OK
		if err != nil {
			statusCode = status.Code(err)
		}

		// Log completion
		reqLogger.Info("gRPC request completed",
			zap.Any("response", resp),
			zap.String("status", statusCode.String()),
			zap.Duration("duration", time.Since(startTime)),
			zap.Error(err),
		)

		return resp, err
	}
}

func metadataFromContext(ctx context.Context, req any) (metadata.MD, string, string) {
	var (
		md                metadata.MD
		reqID, jwtAuthStr string
	)
	// Extract Metadata from context
	if fromInc, ok := metadata.FromIncomingContext(ctx); ok && fromInc != nil {
		md = fromInc.Copy()
	}
	if md != nil {
		// Use x-request-id from metadata if available
		if vals := md.Get(xApiRequestId); len(vals) > 0 {
			reqID = vals[0]
		}
		// Extract authorization from metadata
		if auth := md.Get(headerAuthorization); len(auth) > 0 {
			if strings.HasPrefix(auth[0], "Bearer ") {
				jwtAuthStr = strings.TrimPrefix(auth[0], "Bearer ")
			} else {
				jwtAuthStr = auth[0]
			}
		}
	}
	// Extract request ID if available with req_id
	if r, ok := req.(interface{ GetReqId() string }); ok && r != nil {
		reqID = r.GetReqId()
	}
	// Extract request ID if available with request_id
	if r, ok := req.(interface{ GetRequestId() string }); ok && r != nil {
		reqID = r.GetRequestId()
	}
	// Fallback to generated request ID
	if reqID == "" {
		reqID = fmt.Sprintf("SERVER-GEN-%d", time.Now().UnixNano())
	}
	// Ensure the request ID is in the metadata
	if md == nil {
		md = metadata.Pairs(xApiRequestId, reqID)
	}
	return md, jwtAuthStr, reqID
}
