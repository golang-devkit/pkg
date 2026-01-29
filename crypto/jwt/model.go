package jwt

import (
	"context"
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/golang-jwt/jwt/v5"
)

var (
	SigningMethodEdDSA = jwt.SigningMethodEdDSA
)

// MapClaims defines the custom JWT claims structure.
//
// Example:
//
//	{
//		  "iss": "Issuer,omitempty",
//		  "sub": "Subject,omitempty",
//		  "jti": "ID,omitempty",
//		  "aud": ["Audience,omitempty"],
//		  "nbf": 1753416843,
//		  "iat": 1753416843,
//		  "exp": 1753416963,
//		  "userId":"6883664f484674420f55c16b",
//		  "sessionId": "0cf835de-5c39-481d-a371-94884ba91fcd",
//		  "protobufDataHash":"37a8ddae362e58fbaa3c75f0f201cdd6aa209280cbeb682db29f6d5909595971"
//		}
//
// Fields:
//   - ID: JWT ID, a unique identifier for the token.
//   - Issuer: The issuer of the token.
//   - Subject: The subject of the token.
//   - Audience: The audience for which the token is intended.
//   - NotBefore: The time before which the token is not valid.
//   - IssuedAt: The time at which the token was issued.
//   - ExpiresAt: The expiration time of the token.
//   - SessionId: A custom field representing the session ID.
//   - UserId: A custom field representing the user ID.
//   - ProtobufDataHash: A custom field representing the hash of the protobuf data.
//   - Payload: A custom field that can hold any additional payload data.
type MapClaims struct {
	jwt.RegisteredClaims
	SessionId        string `json:"sessionId"`
	UserId           string `json:"userId"`
	ProtobufDataHash string `json:"protobufDataHash,omitempty"`
	Payload          any    `json:"payload,omitempty"`
}

// ParsePayload parses the JWT claims payload into the given struct.
// If v is proto.Message, please use MapClaims.ParseMessage(v proto.Message) instead.
func (claims *MapClaims) ParsePayload(v any) error {
	if claims.Payload == nil {
		return nil
	}
	b, err := json.Marshal(claims.Payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// ParseMessage parses the JWT claims payload into the given proto.Message.
func (claims *MapClaims) ParseMessage(v proto.Message) error {
	if claims.Payload == nil {
		return nil
	}
	b, err := json.Marshal(claims.Payload)
	if err != nil {
		return err
	}
	return protojson.Unmarshal(b, v)
}

// ApplyContext applies the claims to the given context.
//
// Conficts will be overwritten:
// You should not set these values in the context before calling this method.
//
//	func (claims *MapClaims) ApplyContext(ctx context.Context, reqId string) context.Context {
//		ctx = context.WithValue(ctx, ApiRequestIdKey, reqId)
//		ctx = context.WithValue(ctx, SessionIdKey, claims.SessionId)
//		ctx = context.WithValue(ctx, UserIdKey, claims.UserId)
//		ctx = context.WithValue(ctx, RoleKey, claims.Role)
//		return ctx
//	}
func (claims *MapClaims) ApplyContext(ctx context.Context) context.Context {
	ctx = setSessionIdToContext(ctx, claims.SessionId)
	ctx = setUserIdToContext(ctx, claims.UserId)
	return ctx
}
