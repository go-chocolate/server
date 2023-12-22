package balancer

import "github.com/go-chocolate/server/cluster/endpoint"

type Balancer interface {
	LB(endpoints []*endpoint.Endpoint) (*endpoint.Endpoint, error)
}
