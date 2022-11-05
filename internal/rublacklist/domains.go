package rublacklist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetDomains(ctx context.Context, domains chan<- string) error {
	url := "https://reestr.rublacklist.net/api/v3/domains/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)

	// Read opening delimiter. `[` or `{`
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("decode opening delimiter: %w", err)
	}

	i := 0
	for decoder.More() {
		var domain string
		if err := decoder.Decode(&domain); err != nil {
			return fmt.Errorf("decode line %d: %w", i, err)
		}
		domains <- domain

		i++
	}

	// Read closing delimiter. `]` or `}`
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("decode closing delimiter: %w", err)
	}

	close(domains)

	return nil
}
