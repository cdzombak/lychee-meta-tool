package ai

import "context"

type Client interface {
	GenerateTitle(ctx context.Context, imageURL string) (string, error)
}
