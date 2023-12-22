package traceutil

type Config struct {
	ServiceName    string
	ServiceVersion string
	Attributes     map[string]string
	Sampler        float64
}
