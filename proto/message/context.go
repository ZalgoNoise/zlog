package message

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

const TimeoutSeconds = 30
const StreamTimeoutSeconds = 3600

const DefaultTimeout = time.Second * TimeoutSeconds
const DefaultStreamTimeout = time.Second * StreamTimeoutSeconds

func CtxGet(ctx context.Context) []string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	return md["id"]

}

func NewContext() context.Context {
	bgCtx := context.Background()
	return metadata.NewOutgoingContext(
		bgCtx,
		metadata.Pairs("id", uuid.New().String()))

}

func NewContextTimeout(t time.Duration) (context.Context, context.CancelFunc, string) {
	uuid := uuid.New().String()

	bgCtx := context.Background()

	ctx, cancel := context.WithTimeout(bgCtx, t)
	return metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs("id", uuid)), cancel, uuid
}
