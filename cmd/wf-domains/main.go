package main

import (
	"context"
	"fmt"

	"github.com/cordalace/waterflow/internal/rublacklist"
	"github.com/spf13/pflag"
)

type args struct {
	nolazy bool
}

func parseArgs() *args {
	nolazy := pflag.Bool("nolazy", false, "download full domain list and store it in memory before printing")

	pflag.Parse()

	return &args{nolazy: *nolazy}
}

func main() {
	args := parseArgs()

	domains := make(chan string)
	go func() {
		if err := rublacklist.GetDomains(context.Background(), domains); err != nil {
			panic(err)
		}
	}()

	if args.nolazy {
		bufferedDomainList := make([]string, 0)
		for domain := range domains {
			bufferedDomainList = append(bufferedDomainList, domain)
		}

		for _, domain := range bufferedDomainList {
			fmt.Println(domain)
		}
	} else {
		for domain := range domains {
			fmt.Println(domain)
		}
	}
}
