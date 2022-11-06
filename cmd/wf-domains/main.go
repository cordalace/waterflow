package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cordalace/waterflow/internal/rublacklist"
	"github.com/spf13/pflag"
)

const (
	initialDomainsBufSize = 500000 // 4915200
)

type args struct {
	nolazy bool
}

func parseArgs() *args {
	nolazy := pflag.Bool("nolazy", false, "download full domain list and store it in memory before printing")

	pflag.Parse()

	return &args{nolazy: *nolazy}
}

func maybeBufferDomains(ctx context.Context, nolazy bool, lazyDomains <-chan string) <-chan string {
	if nolazy {
		output := make(chan string)

		bufferedDomainList := make([]string, 0, initialDomainsBufSize)
		for domain := range lazyDomains {
			bufferedDomainList = append(bufferedDomainList, domain)
		}

		go func() {
			for _, domain := range bufferedDomainList {
				select {
				case <-ctx.Done():
					close(output)
				default:
					output <- domain
				}
			}
			close(output)
		}()

		return output
	}

	return lazyDomains
}

func main() {
	args := parseArgs()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)

	lazyDomains := make(chan string)
	go func() {
		if err := rublacklist.GetDomains(ctx, lazyDomains); err != nil {
			log.Fatalf("error receiving domains: %v", err)
		}
	}()

	domains := maybeBufferDomains(ctx, args.nolazy, lazyDomains)

	go func() {
		for domain := range domains {
			fmt.Println(domain)
		}
		done <- syscall.SIGTERM
	}()

	<-done
	cancel()
}
