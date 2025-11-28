package messaging

import "context"

// Publisher defines the interface for message publishing
type Publisher interface {
	Publish(ctx context.Context, message []byte) error
}
