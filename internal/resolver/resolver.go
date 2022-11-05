package resolver

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/miekg/dns"
)

type DNSType int

type Resolver struct {
	numWorkers       int
	timeout          time.Duration
	bootstrapServers []string
	upstreamServers  []string
	wg               sync.WaitGroup

	upstreams []upstream.Upstream
}

func NewResolver() *Resolver {
	r := &Resolver{
		numWorkers:       DefaultNumWorkers(),
		timeout:          DefaultTimeout(),
		bootstrapServers: DefaultBootstrapServers(),
		upstreamServers:  DefaultUpstreamServers(),
		wg:               sync.WaitGroup{},
		upstreams:        nil,
	}
	return r
}

func (r *Resolver) Init() error {
	upstreamOpts := &upstream.Options{
		Bootstrap:                 r.bootstrapServers,
		Timeout:                   r.timeout,
		ServerIPAddrs:             nil,   // use default
		InsecureSkipVerify:        false, // use default
		HTTPVersions:              nil,   // use default
		VerifyServerCertificate:   nil,   // use default
		VerifyConnection:          nil,   // use default
		VerifyDNSCryptCertificate: nil,   // use default
	}
	upstreams := make([]upstream.Upstream, len(r.upstreamServers))
	for i, upstreamServer := range r.upstreamServers {
		u, err := upstream.AddressToUpstream(upstreamServer, upstreamOpts)
		if err != nil {
			return err
		}
		upstreams[i] = u
	}
	r.upstreams = upstreams

	return nil
}

type Result struct {
	domain string
	IPs    []net.IP
	Err    error
}

func (r *Resolver) Resolve(ctx context.Context, domains <-chan string, results chan<- Result) {
	for w := 1; w <= r.numWorkers; w++ {
		r.wg.Add(1)
		go r.worker(ctx, domains, results)
	}

	r.wg.Wait()
	close(results)
}

func (r *Resolver) worker(ctx context.Context, domains <-chan string, results chan<- Result) {
	for domain := range domains {
		ips, err := r.resolve(ctx, domain)
		if err != nil {
			results <- Result{domain: domain, IPs: nil, Err: err}
		}

		results <- Result{domain: domain, IPs: ips}
	}
	r.wg.Done()
}

func (r *Resolver) resolve(ctx context.Context, domain string) ([]net.IP, error) {
	req := new(dns.Msg)
	req.SetQuestion(domain+".", dns.TypeA)
	responses, err := upstream.ExchangeAll(r.upstreams, req)
	if err != nil {
		return nil, err
	}

	ipSet := make(map[string]struct{})
	for _, resp := range responses {
		for _, answer := range resp.Resp.Answer {
			if answer.Header().Rrtype == dns.TypeA {
				answerTypeA, ok := answer.(*dns.A)
				if !ok {
					panic("error dns answer type conversion")
				}
				ipSet[answerTypeA.A.String()] = struct{}{}
			}
		}
	}

	ips := make([]net.IP, 0, len(ipSet))
	for ip := range ipSet {
		ips = append(ips, net.ParseIP(ip))
	}

	return ips, nil
}
