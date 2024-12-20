package utils

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var client = &http.Client{
	Timeout: 3 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func CheckURL(urlStr string) (int, int64, error) {
	urlStr = normalizeURL(urlStr)

	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return 0, 0, err
	}

	req.Header.Set("User-Agent", "M3U8Validator/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	elapsed := time.Since(start).Milliseconds()

	return resp.StatusCode, elapsed, nil
}

func normalizeURL(urlStr string) string {
	if strings.Contains(urlStr, "[") && strings.Contains(urlStr, "]") {
		return urlStr
	}

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "http://" + urlStr
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	if strings.Count(u.Host, ":") > 1 {
		host := u.Host
		if !strings.HasPrefix(host, "[") {
			host = "[" + host + "]"
		}
		u.Host = host
	}

	return u.String()
}
