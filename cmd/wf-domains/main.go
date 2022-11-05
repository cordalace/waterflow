package main

import (
	"fmt"

	"github.com/cordalace/waterflow/internal/rublacklist"
)

func main() {
	domains := make(chan string)
	go func() {
		if err := rublacklist.GetDomains(domains); err != nil {
			panic(err)
		}
	}()

	for domain := range domains {
		fmt.Println(domain)
	}
}
