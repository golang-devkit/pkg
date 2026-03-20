package net

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
	"google.golang.org/grpc/credentials"
)

// tokenCredentials wraps an oauth2.TokenSource into a gRPC PerRPCCredentials
type tokenCredentials struct {
	source oauth2.TokenSource
	ctx    context.Context
}

// GetRequestMetadata implements PerRPCCredentials.GetRequestMetadata
func (tc *tokenCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := tc.source.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// For Cloud Run service-to-service authentication, send the token (Access Token or ID Token)
	return map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", token.AccessToken),
	}, nil
}

// RequireTransportSecurity implements PerRPCCredentials.RequireTransportSecurity
func (tc *tokenCredentials) RequireTransportSecurity() bool {
	return true
}

// createApplicationCredentials creates PerRPCCredentials for application authentication
// It can use either service account JSON credentials or default credentials
func createApplicationCredentials(ctx context.Context, audience string, jsonCreds []byte, timeout time.Duration) (credentials.PerRPCCredentials, error) {

	// Strip port if present (must be target service URL not prefix with http:// or https://)
	if idx := strings.Index(audience, ":"); idx != -1 {
		audience = audience[:idx]
	}
	// Ensure https:// prefix
	if !strings.HasPrefix(audience, "https://") && !strings.HasPrefix(audience, "http://") {
		audience = "https://" + audience
	}

	// Make context for token source caller
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if len(jsonCreds) > 0 && audience != "" {
		// Use service account credentials from the specified key file (with audience)
		return createServiceAccountIDTokenCredentials(ctx, audience, jsonCreds)
	}
	if len(jsonCreds) > 0 && audience == "" {
		// Use service account credentials from the specified key file (without audience)
		return createServiceAccountCredentials(ctx, jsonCreds)
	}
	// Use default credentials
	// Create credentials with the token
	tokenSource, err := idtoken.NewTokenSource(ctx, audience)
	if err != nil {
		return nil, fmt.Errorf("failed to create token source: %v", err)
	}
	return &tokenCredentials{
		source: tokenSource,
		ctx:    ctx,
	}, nil
}

// createServiceAccountCredentials creates PerRPCCredentials from a service account key file
// This creates standard OAuth2 credentials (access token), typically used for Google Cloud APIs
func createServiceAccountCredentials(ctx context.Context, jsonCreds []byte) (credentials.PerRPCCredentials, error) {
	// Get credentials using a different approach - get access token first
	creds, err := google.CredentialsFromJSON(ctx, jsonCreds, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials from service account key: %w", err)
	}

	// Wrap the token source
	return &tokenCredentials{
		source: creds.TokenSource,
		ctx:    ctx,
	}, nil
}

// createServiceAccountIDTokenCredentials creates PerRPCCredentials from a service account key file for ID tokens
func createServiceAccountIDTokenCredentials(ctx context.Context, audience string, jsonCreds []byte) (credentials.PerRPCCredentials, error) {
	// Create an ID token source using the service account JSON
	ts, err := idtoken.NewTokenSource(ctx, audience, option.WithCredentialsJSON(jsonCreds))
	if err != nil {
		return nil, fmt.Errorf("failed to create ID token source: %w", err)
	}

	return &tokenCredentials{
		source: ts,
		ctx:    ctx,
	}, nil
}

type CredentialOption struct {
	// JSONCredentials is the service account JSON credentials to use for authentication.
	// If not provided, default application credentials will be used.
	JSONCredentials []byte

	// TransportCredentials specifies the transport credentials (e.g., TLS) to use for the gRPC connection.
	// If not provided, default TLS credentials will be used.
	TransportCredentials credentials.TransportCredentials

	// GrpcClientKeepAliveTime is the duration after which if the client doesn't see any activity it pings the server to see if the transport is still alive.
	// If zero, default value will be used. Else must be greater than 0.
	GrpcClientKeepAliveTime    time.Duration
	GrpcClientKeepAliveTimeout time.Duration
	PermitWithoutStream        bool

	//
	SkipPerRPCCredentials bool
}
