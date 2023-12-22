package discovery

import (
	"context"

	"github.com/go-chocolate/server/cluster/endpoint"
)

type Discovery interface {
	Discover(ctx context.Context, service string) ([]*endpoint.Endpoint, error)
	DiscoverWithMetadata(ctx context.Context, service string, metadata map[string]string) ([]*endpoint.Endpoint, error)
}
