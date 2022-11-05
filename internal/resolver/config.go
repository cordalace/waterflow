package resolver

import "time"

func (r *Resolver) WithOptions(opts ...Option) *Resolver {
	clone := *r
	for _, opt := range opts {
		opt(&clone)
	}
	return &clone
}

type Option func(r *Resolver)

func DefaultNumWorkers() int {
	return 100
}

func WithNumWorkers(numWorkers int) Option {
	return func(r *Resolver) {
		r.numWorkers = numWorkers
	}
}

func DefaultTimeout() time.Duration {
	return 5 * time.Second
}

func WithTimeout(timeout time.Duration) Option {
	return func(r *Resolver) {
		r.timeout = timeout
	}
}

func DefaultBootstrapServers() []string {
	return []string{"1.1.1.1", "1.0.0.1"}
}

func WithBootstrapServers(servers []string) Option {
	return func(r *Resolver) {
		r.bootstrapServers = servers
	}
}

func WithUpstreamServers(servers []string) Option {
	return func(r *Resolver) {
		r.upstreamServers = servers
	}
}

func DefaultUpstreamServers() []string {
	return []string{"https://1.1.1.1/dns-query", "https://1.0.0.1/dns-query"}
}
