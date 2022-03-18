package message

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

const defaultTimeout = time.Second * 30

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

func NewContextTimeout() (context.Context, context.CancelFunc) {
	bgCtx := context.Background()

	ctx, cancel := context.WithTimeout(bgCtx, defaultTimeout)
	return metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs("id", uuid.New().String())), cancel
}
