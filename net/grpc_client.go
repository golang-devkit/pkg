package net

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"testing"
	"time"

	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

var (
	grpcClientKeepAliveTime    = 120 * time.Second // Gửi keepalive ping mỗi 120s
	grpcClientKeepAliveTimeout = 30 * time.Second  // Timeout cho keepalive ping
	permitWithoutStream        = true              // Cho phép keepalive khi không có stream
)

// GetAuthenticatedContext creates a new context with authentication metadata
// for the given target audience.
// It retrieves an ID token and adds it to the context metadata.
// Returns the new context, the metadata, and any error encountered.
//
// Parameters:
//   - ctx: The base context to use for authentication.
//   - targetAudience: The audience for which the ID token is requested. It must match the service's expected audience.
//
// Returns:
//   - context.Context: A new context containing the authentication metadata.
//   - metadata.MD: The metadata containing the authorization header.
//   - error: An error if the token retrieval or context creation fails.
//
// Deprecated: This function is deprecated.
// Use alternative authentication methods as recommended by the gRPC and Google Cloud documentation.
//
// Note: In this package, prefer using createApplicationCredentials for creating PerRPCCredentials.
func GetAuthenticatedContext(ctx context.Context, targetAudience string) (context.Context, metadata.MD, error) {
	if testing.Testing() {
		return ctx, metadata.Pairs("Authorization", "Bearer "+"<idToken>"), nil
	}
	// Get the ID token
	idToken, err := GetIDTokenFromClient(ctx, targetAudience)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ID token: %v", err)
	}

	// Create a new context with the authorization header
	// Prepare metadata with API key for authentication
	md := metadata.Pairs("Authorization", "Bearer "+idToken)
	// Create a new context with the metadata
	outgoingContext := metadata.NewOutgoingContext(ctx, md)
	// Call the provided function with the new context
	return outgoingContext, md, nil
}

// GetIDTokenFromClient retrieves an ID token for the specified target audience.
// It uses the Google ID token package to obtain the token.
//
// Parameters:
//   - ctx: The context to use for the token retrieval.
//   - targetAudience: The audience for which the ID token is requested. It must match the service's expected audience.
//
// Returns:
//   - string: The retrieved ID token.
//   - error: An error if the token retrieval fails.
//
// Deprecated: This function is deprecated.
// Use alternative authentication methods as recommended by the gRPC and Google Cloud documentation.
//
// Note: In this package, prefer using createApplicationCredentials for creating PerRPCCredentials.
func GetIDTokenFromClient(ctx context.Context, targetAudience string) (string, error) {
	// Create a token source
	tokenSource, err := idtoken.NewTokenSource(ctx, targetAudience)
	if err != nil {
		return "", fmt.Errorf("failed to create token source: %v", err)
	}

	// Get the token
	token, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}

	return token.AccessToken, nil
}

func NewCloudRunGRPCClient(target string, with ...CredentialOption) (*grpc.ClientConn, error) {

	// Validate target format (make sure it doesn't contain http:// or https:// prefix)
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return nil, fmt.Errorf("target should be in the format 'host:port', without 'http://' or 'https://'")
	}
	// Prepare dial options
	opts, err := prepareDialOptions(target, with...)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare dial options: %w", err)
	}
	// Create the gRPC client connection
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return conn, nil
}

func prepareDialOptions(target string, with ...CredentialOption) ([]grpc.DialOption, error) {

	var (
		certificateJson  []byte
		transport        credentials.TransportCredentials = credentials.NewTLS(&tls.Config{})
		clientParameters keepalive.ClientParameters       = keepalive.ClientParameters{
			Time:                grpcClientKeepAliveTime,    // Gửi keepalive ping mỗi 120s
			Timeout:             grpcClientKeepAliveTimeout, // Timeout cho keepalive ping
			PermitWithoutStream: permitWithoutStream,        // Cho phép keepalive khi không có stream
		}

		// Process options
		flagCredentialSet bool
		flagTransportSet  bool
		flagKeepaliveSet  bool

		skipPerRPCCredentials bool
	)

	// Establish options
	for _, opt := range with {
		// Use JSON credentials from options if provided
		if !flagCredentialSet && len(opt.JSONCredentials) > 0 {
			certificateJson = opt.JSONCredentials
			flagCredentialSet = true
		}
		// Use transport credentials from options if provided
		if !flagTransportSet && opt.TransportCredentials != nil {
			transport = opt.TransportCredentials
			flagTransportSet = true
		}
		// Use keepalive parameters from options if provided
		if !flagKeepaliveSet &&
			(opt.GrpcClientKeepAliveTime > 0 || opt.GrpcClientKeepAliveTimeout > 0) {
			clientParameters.Time = opt.GrpcClientKeepAliveTime
			clientParameters.Timeout = opt.GrpcClientKeepAliveTimeout
			clientParameters.PermitWithoutStream = opt.PermitWithoutStream
			flagKeepaliveSet = true
		}
		// Ignore PerRPCCredentials if specified by options
		if opt.SkipPerRPCCredentials {
			skipPerRPCCredentials = true
		}
	}
	//
	opts := []grpc.DialOption{
		// TLS credentials (always required for Cloud Run)
		grpc.WithTransportCredentials(transport),
		// Keepalive để phát hiện connection chết sớm
		grpc.WithKeepaliveParams(clientParameters),
	}

	// Skip PerRPCCredentials return basic dial options
	if skipPerRPCCredentials {
		return opts, nil
	}

	// Create PerRPCCredentials for authentication
	appCreds, err := createApplicationCredentials(context.TODO(), target, certificateJson, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create application credentials: %w", err)
	}
	opts = append(opts, grpc.WithPerRPCCredentials(appCreds))

	return opts, nil
}
