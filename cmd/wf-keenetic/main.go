package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cordalace/waterflow/internal/resolver"
	"github.com/cordalace/waterflow/internal/rublacklist"
	"github.com/spf13/pflag"
)

type args struct {
	gateway          string
	numWorkers       int
	timeout          time.Duration
	bootstrapServers []string
	upstreamServers  []string
}

func parseArgs() *args {
	gateway := pflag.String("gateway", "192.168.0.1", "route gateway IP address")
	numWorkers := pflag.Int("workers", resolver.DefaultNumWorkers(), "number of workers")
	timeout := pflag.Float64("timeout", resolver.DefaultTimeout().Seconds(), "timeout for DNS requests (in seconds)")
	bootstrapServers := pflag.StringArray("bootstrap", []string{"1.1.1.1", "1.0.0.1"}, "bootstrap DNS servers")
	upstreamServers := pflag.StringArray("upstream", []string{"https://1.1.1.1/dns-query", "https://1.0.0.1/dns-query"}, "upstream DNS servers")

	pflag.Parse()

	return &args{
		gateway:          *gateway,
		numWorkers:       *numWorkers,
		timeout:          time.Duration(*timeout * 1e9),
		bootstrapServers: *bootstrapServers,
		upstreamServers:  *upstreamServers,
	}
}

func main() {
	args := parseArgs()

	domains := make(chan string)
	go func() {
		if err := rublacklist.GetDomains(domains); err != nil {
			panic(err)
		}
	}()

	resolveResults := make(chan resolver.Result)

	r := resolver.NewResolver().WithOptions(
		resolver.WithNumWorkers(args.numWorkers),
		resolver.WithTimeout(args.timeout),
		resolver.WithBootstrapServers(args.bootstrapServers),
		resolver.WithUpstreamServers(args.upstreamServers),
	)
	if err := r.Init(); err != nil {
		panic(err)
	}
	go r.Resolve(context.Background(), domains, resolveResults)

	for result := range resolveResults {
		for _, ip := range result.IPs {
			if ip.To4() != nil {
				fmt.Printf("route ADD %s MASK 255.255.255.255 %s\n", ip.String(), args.gateway)
			}
		}
	}
}
