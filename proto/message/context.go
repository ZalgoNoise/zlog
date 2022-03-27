package message

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

// default Unary RPC timeout in seconds
const TimeoutSeconds = 30

// default Stream RPC timeout in seconds
const StreamTimeoutSeconds = 3600

// default Unary RPC timeout
const DefaultTimeout = time.Second * TimeoutSeconds

// default Stream RPC timeout
const DefaultStreamTimeout = time.Second * StreamTimeoutSeconds

// metadata key for a request ID
const RequestIDKey = "id"

// default request ID when unset
const DefaultRequestID = "anonymous"

// CtxGet function will retrieve the metadata from an incoming context,
// as set with grpc/metadata.
//
// This function is currently used to retrieve request IDs from contexts,
// but it can be extended to retrieve other sorts of metadata
func CtxGet(ctx context.Context, key string) []string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	return md[key]

}

// NewContext function will create a new background context with metadata,
// containing a new, random UUID to represent this transaction or connection
func NewContext() context.Context {
	bgCtx := context.Background()
	return metadata.NewOutgoingContext(
		bgCtx,
		metadata.Pairs(RequestIDKey, uuid.New().String()))

}

// NewContextTimeout function will create a new context with timeout and metadata,
// containing a new, random UUID to represent this transaction or connection
func NewContextTimeout(t time.Duration) (context.Context, context.CancelFunc, string) {
	uuid := uuid.New().String()

	bgCtx := context.Background()

	ctx, cancel := context.WithTimeout(bgCtx, t)
	return metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs(RequestIDKey, uuid)), cancel, uuid
}
