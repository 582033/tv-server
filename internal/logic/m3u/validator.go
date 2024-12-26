package m3u

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
	"tv-server/utils"
)

var (
	processCount int
	totalCount   int
	processLock  sync.RWMutex
)

//验证及去重
/*
 @param allEntries []m3u.Entry // 所有的链接
 @param maxLatency time.Duration // 单位毫秒,最大延迟,超过此值的链接将被丢弃
 @param workerCount int // 工作协程数
 @return []m3u.Entry, []m3u.Entry, error
*/
func ValidateAndUnique(allEntries []Entry, maxLatency time.Duration, workerCount int) ([]Entry, []Entry, error) {

	// 使用带缓冲的通道进行并发控制
	tasks := make(chan Entry, len(allEntries))
	results := make(chan Entry, len(allEntries))
	process := make(chan int, 0)
	done := make(chan bool)

	validEntries := make([]Entry, 0, len(allEntries))

	//开始批量验证提醒,输出并发的协程数以及总耗时
	fmt.Println("maxLatency:", maxLatency)
	fmt.Printf("开始批量验证，总共链接数:%d，并发协程数: %d, 预计耗时:%s\n", len(allEntries), workerCount, utils.CalculateTotalTimeToString(maxLatency, workerCount, len(allEntries)))

	// 启动工作协程
	for i := 0; i < workerCount; i++ {
		go func(entryChan <-chan Entry, resultChan chan<- Entry, processChan chan<- int, maxLatency time.Duration) {
			for entry := range entryChan {
				valid, err := ValidateURL(entry.URL, maxLatency)
				if valid && err == nil {
					resultChan <- entry
				}
				processChan <- 1
			}
		}(tasks, results, process, maxLatency)
	}

	// 发送任务
	go func() {
		for _, entry := range allEntries {
			tasks <- entry
		}
		close(tasks)
	}()

	// 收集结果
	go func() {
		for entry := range results {
			validEntries = append(validEntries, entry)
		}
		done <- true
	}()

	// 收集进度
	go func() {
		for range process {
			processLock.Lock()
			processCount++
			processLock.Unlock()
		}
	}()

	// 设置超时(此处设置的是未验证完成也返回结果
	//timeout := time.After(maxLatency * time.Millisecond)

	var finalValidEntries []Entry
	select {
	case <-done:
		// 在验证完成后进行去重
		urlMap := make(map[string]Entry)
		for _, entry := range validEntries {
			urlMap[entry.URL] = entry
		}
		for _, entry := range urlMap {
			finalValidEntries = append(finalValidEntries, entry)
		}
		fmt.Printf("验证完成，原始链接 %d 个，验证通过 %d 个，去重后有效链接 %d 个\n",
			len(allEntries), len(validEntries), len(finalValidEntries))

		/*
			case <-timeout:
				// 超时时也进行去重
				urlMap := make(map[string]Entry)
				for _, entry := range validEntries {
					urlMap[entry.URL] = entry
				}
				for _, entry := range urlMap {
					finalValidEntries = append(finalValidEntries, entry)
				}
				fmt.Printf("验证超时，原始链接 %d 个，验证通过 %d 个，去重后有效链接 %d 个\n",
					len(allEntries), len(validEntries), len(finalValidEntries))
		*/
	}

	return validEntries, finalValidEntries, nil

}

// ValidateURLsWithLatency 使用指定的延迟阈值验证URLs
func ValidateURLsWithLatency(entries []Entry, maxLatency time.Duration) []Entry {
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
func validateM3U8Stream(url string, maxLatency time.Duration) (bool, int64) {
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

// ValidateURL 检查URL是否有效且延迟在允许范围内
func ValidateURL(url string, maxLatency time.Duration) (bool, error) {
	//fmt.Printf("开始验证链接: %s (最大延迟: %dms)\n,", url, maxLatency)

	client := &http.Client{
		Timeout: maxLatency * time.Millisecond,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	isValid := resp.StatusCode == http.StatusOK

	// Read a small portion of the body to ensure content is accessible
	if isValid {
		buffer := make([]byte, 1024)
		_, err := resp.Body.Read(buffer)
		if err != nil && err != io.EOF {
			isValid = false
			return isValid, fmt.Errorf("读取内容失败: %w", err)
		}
	}

	return isValid, nil
}
func checkWithFFprobe(url string, maxLatency time.Duration) (bool, error) {
	// Create a context with timeout based on the maxLatency
	ctx, cancel := context.WithTimeout(context.Background(), maxLatency*time.Millisecond)
	defer cancel()

	// Run the ffprobe command
	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=codec_name", "-of", "default=noprint_wrappers=1:nokey=1", url)

	// Capture command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return false, fmt.Errorf("ffprobe 超时: 超过最大延迟限制 %dms", maxLatency)
		}
		return false, fmt.Errorf("ffprobe 错误: %w", err)
	}

	// If output is empty, means no valid video stream found
	if len(output) == 0 {
		return false, fmt.Errorf("没有找到有效的视频流")
	}

	// Print stream type
	fmt.Printf("视频流类型: %s\n", string(output))
	return true, nil
}

func GetProcess() float64 {
	processLock.RLock()
	defer processLock.RUnlock()
	if totalCount == 0 {
		return 0
	}
	return float64(processCount) / float64(totalCount)
}
