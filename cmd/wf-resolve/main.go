package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cordalace/waterflow/internal/resolver"
	"github.com/cordalace/waterflow/internal/rublacklist"
	"github.com/spf13/pflag"
)

type args struct {
	numWorkers       int
	timeout          time.Duration
	bootstrapServers []string
	upstreamServers  []string
}

func parseArgs() *args {
	numWorkers := pflag.Int("workers", resolver.DefaultNumWorkers(), "number of workers")
	timeout := pflag.Float64("timeout", resolver.DefaultTimeout().Seconds(), "timeout for DNS requests (in seconds)")
	bootstrapServers := pflag.StringArray("bootstrap", resolver.DefaultBootstrapServers(), "bootstrap DNS servers")
	upstreamServers := pflag.StringArray("upstream", resolver.DefaultUpstreamServers(), "upstream DNS servers")

	pflag.Parse()

	return &args{
		numWorkers:       *numWorkers,
		timeout:          time.Duration(*timeout * 1e09),
		bootstrapServers: *bootstrapServers,
		upstreamServers:  *upstreamServers,
	}
}

func mainWithExitCode() int {
	args := parseArgs()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	domains := make(chan string)
	go func() {
		if err := rublacklist.GetDomains(ctx, domains); err != nil {
			log.Printf("error receiving domains: %v", err)
			return
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
		log.Printf("error initializing resolver: %v", err)
		return 1
	}
	go r.Resolve(ctx, domains, resolveResults)

	uniqueIPs := make(map[string]struct{})
	for result := range resolveResults {
		for _, ip := range result.IPs {
			if ip.To4() != nil {
				ipString := ip.String()
				if _, ok := uniqueIPs[ipString]; !ok {
					uniqueIPs[ipString] = struct{}{}
					fmt.Println(ipString)
				}
			}
		}
	}

	return 0
}

func main() {
	os.Exit(mainWithExitCode())
}
