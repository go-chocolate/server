package endpoint

type Protocol string

const (
	HTTP Protocol = "http"
	GRPC Protocol = "grpc"
)

type Metadata map[string]string

type Endpoint struct {
	Protocol Protocol `json:"protocol"`
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Metadata Metadata `json:"metadata"`
	Health   *Health  `json:"health"`
}

type Option func(o *Endpoint)

func applyOptions(o *Endpoint, options ...Option) {
	for _, opt := range options {
		opt(o)
	}
}

func WithMetadata(metadata Metadata) Option {
	return func(o *Endpoint) {
		o.Metadata = metadata
	}
}

func WithHealthCheck(h *Health) Option {
	return func(o *Endpoint) {
		o.Health = h
	}
}

func New(protocol Protocol, name string, address string, options ...Option) *Endpoint {
	endpoint := &Endpoint{
		Protocol: protocol,
		Name:     name,
		Address:  address,
	}
	applyOptions(endpoint, options...)
	return endpoint
}
