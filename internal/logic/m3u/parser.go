package m3u

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Entry struct {
	Metadata string
	URL      string
}

func Parse(content string) []Entry {
	var entries []Entry
	var currentMetadata string

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#EXTM3U") {
			entries = append(entries, Entry{Metadata: line})
			continue
		}
		if strings.HasPrefix(line, "#EXTINF") {
			currentMetadata = line
			continue
		}
		if strings.HasPrefix(line, "http") {
			entries = append(entries, Entry{
				Metadata: currentMetadata,
				URL:      line,
			})
		}
	}
	return entries
}

// ParseFile 从文件解析M3U
func ParseFile(filename string) ([]Entry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return Parse(string(content)), nil
}

// ParseURL 从URL解析M3U
func ParseURL(url string) ([]Entry, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return Parse(string(content)), nil
}

// ValidateURL 检查URL是否有效且延迟在允许范围内
func ValidateURL(url string, maxLatency int) bool {
	fmt.Printf("开始验证链接: %s (最大延迟: %dms)\n", url, maxLatency)

	client := &http.Client{
		Timeout: time.Duration(maxLatency) * time.Millisecond,
		Transport: &http.Transport{
			DisableKeepAlives:     true,
			IdleConnTimeout:       time.Duration(maxLatency) * time.Millisecond,
			TLSHandshakeTimeout:   time.Duration(maxLatency) * time.Millisecond,
			ResponseHeaderTimeout: time.Duration(maxLatency) * time.Millisecond,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100,
		},
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		fmt.Printf("创建请求失败: %s, 错误: %v\n", url, err)
		return false
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %s, 错误: %v\n", url, err)
		return false
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()
	isValid := resp.StatusCode == http.StatusOK && latency <= int64(maxLatency)

	if isValid {
		fmt.Printf("链接有效: %s (延迟: %dms)\n", url, latency)
	} else {
		fmt.Printf("链接无效: %s (状态码: %d, 延迟: %dms)\n", url, resp.StatusCode, latency)
	}

	return isValid
}
