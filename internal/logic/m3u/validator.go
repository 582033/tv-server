package m3u

import (
	"bufio"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ValidateURLsWithLatency 使用指定的延迟阈值验证URLs
func ValidateURLsWithLatency(entries []Entry, maxLatency int) []Entry {
	var validEntries []Entry
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, entry := range entries {
		wg.Add(1)
		go func(e Entry) {
			defer wg.Done()

			// 验证 m3u8 文件
			isValid, _ := validateM3U8Stream(e.URL, maxLatency)
			if isValid {
				mu.Lock()
				validEntries = append(validEntries, e)
				mu.Unlock()
			}
		}(entry)
	}

	wg.Wait()
	return validEntries
}

// validateM3U8Stream 验证 m3u8 流
func validateM3U8Stream(url string, maxLatency int) (bool, int64) {
	client := &http.Client{
		Timeout: time.Duration(maxLatency*2) * time.Millisecond,
	}

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, 0
	}

	// 添加 User-Agent 头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// 1. 首先获取 m3u8 文件
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return false, 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, 0
	}

	// 读取并解析 m3u8 内容
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, 0
	}

	// 检查是否是有效的 m3u8 文件
	if !isValidM3U8Content(string(content)) {
		return false, 0
	}

	// 如果是主 m3u8，尝试获取第一个分片
	tsURL := getFirstSegmentURL(string(content), url)
	if tsURL != "" {
		// 为分片请求创建新的请求对象
		tsReq, err := http.NewRequest("GET", tsURL, nil)
		if err != nil {
			return false, 0
		}
		// 同样添加 User-Agent 头
		tsReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

		tsResp, err := client.Do(tsReq)
		if err != nil {
			return false, 0
		}
		defer tsResp.Body.Close()

		if tsResp.StatusCode != http.StatusOK {
			return false, 0
		}

		// 读取一小部分数据以验证分片可访问性
		buffer := make([]byte, 1024)
		_, err = tsResp.Body.Read(buffer)
		if err != nil && err != io.EOF {
			return false, 0
		}
	}

	// 计算总延迟
	latency := time.Since(start).Milliseconds()
	return latency <= int64(maxLatency), latency
}

// isValidM3U8Content 检查内容是否是有效的 m3u8 文件
func isValidM3U8Content(content string) bool {
	return strings.Contains(content, "#EXTM3U")
}

// getFirstSegmentURL 获取第一个分片的 URL
func getFirstSegmentURL(content, baseURL string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") && (strings.HasSuffix(line, ".ts") || strings.HasSuffix(line, ".m3u8")) {
			// 处理相对路径
			if !strings.HasPrefix(line, "http") {
				if strings.HasPrefix(line, "/") {
					// 绝对路径
					baseURL := getBaseURL(baseURL)
					return baseURL + line
				}
				// 相对路径
				return getDirectoryURL(baseURL) + line
			}
			return line
		}
	}
	return ""
}

// getBaseURL 获取基础 URL
func getBaseURL(url string) string {
	if idx := strings.Index(url[8:], "/"); idx != -1 {
		return url[:idx+8]
	}
	return url
}

// getDirectoryURL 获取目录 URL
func getDirectoryURL(url string) string {
	if idx := strings.LastIndex(url, "/"); idx != -1 {
		return url[:idx+1]
	}
	return url + "/"
}

// ValidateURLs 使用默认的1000ms延迟阈值验证URLs
func ValidateURLs(entries []Entry) []Entry {
	return ValidateURLsWithLatency(entries, 1000)
}
