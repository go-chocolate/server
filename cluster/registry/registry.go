package registry

import (
	"context"

	"github.com/go-chocolate/server/cluster/endpoint"
)

type Registry interface {
	Register(ctx context.Context, end *endpoint.Endpoint) error
}
