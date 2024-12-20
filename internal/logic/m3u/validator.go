package m3u

import (
	"context"
	"net/http"
	"sync"
	"time"
)

func ValidateURLs(entries []Entry) []Entry {
	var validEntries []Entry
	var mutex sync.Mutex
	var wg sync.WaitGroup

	for _, entry := range entries {
		if entry.URL == "" {
			mutex.Lock()
			validEntries = append(validEntries, entry)
			mutex.Unlock()
			continue
		}

		wg.Add(1)
		go func(entry Entry) {
			defer wg.Done()
			if isValidURL(entry.URL) {
				mutex.Lock()
				validEntries = append(validEntries, entry)
				mutex.Unlock()
			}
		}(entry)
	}

	wg.Wait()
	return validEntries
}

func isValidURL(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return false
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()
	return (resp.StatusCode >= 200 && resp.StatusCode < 400) && latency <= 1000
}
