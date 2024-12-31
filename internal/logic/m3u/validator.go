package m3u

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
	"tv-server/utils"

	"github.com/panjf2000/ants/v2"
)

var (
	processCount int
	totalCount   int
	processLock  sync.RWMutex
)

type validateTask struct {
	entry      Entry
	maxLatency time.Duration
	results    chan<- Entry
	process    chan<- int
}

func validateWorker(task interface{}) {
	t := task.(*validateTask)
	valid, err := ValidateURL(t.entry.URL, t.maxLatency)
	if valid && err == nil {
		select {
		case t.results <- t.entry:
		default:
			fmt.Printf("警告: 无法发送结果: %s\n", t.entry.URL)
		}
	}
	select {
	case t.process <- 1:
	default:
		fmt.Printf("警告: 无法更新进度: %s\n", t.entry.URL)
	}
}

func ValidateAndUnique(allEntries []Entry, maxLatency time.Duration, workerCount int) ([]Entry, []Entry, error) {
	if workerCount > len(allEntries) {
		workerCount = len(allEntries)
	}

	pool, err := ants.NewPool(workerCount)
	if err != nil {
		return nil, nil, fmt.Errorf("创建协程池失败: %w", err)
	}
	defer pool.Release()

	results := make(chan Entry, len(allEntries))
	process := make(chan int, len(allEntries))
	validEntries := make([]Entry, 0, len(allEntries))
	var wg sync.WaitGroup

	fmt.Printf("开始批量验证，总共链接数:%d，并发协程数: %d, 预计耗时:%s\n",
		len(allEntries), workerCount,
		utils.CalculateTotalTimeToString(maxLatency, workerCount, len(allEntries)))

	processLock.Lock()
	processCount = 0
	totalCount = len(allEntries)
	processLock.Unlock()

	go func() {
		for range process {
			processLock.Lock()
			processCount++
			processLock.Unlock()
		}
	}()

	for _, entry := range allEntries {
		wg.Add(1)
		task := &validateTask{
			entry:      entry,
			maxLatency: maxLatency,
			results:    results,
			process:    process,
		}

		if err := pool.Submit(func() {
			defer wg.Done()
			validateWorker(task)
		}); err != nil {
			fmt.Printf("提交任务失败: %v\n", err)
			wg.Done()
			continue
		}
	}

	go func() {
		wg.Wait()
		close(results)
		close(process)
	}()

	for entry := range results {
		validEntries = append(validEntries, entry)
	}

	// 去重
	urlMap := make(map[string]Entry)
	for _, entry := range validEntries {
		urlMap[entry.URL] = entry
	}

	finalValidEntries := make([]Entry, 0, len(urlMap))
	for _, entry := range urlMap {
		finalValidEntries = append(finalValidEntries, entry)
	}

	fmt.Println("验证完成！")
	return validEntries, finalValidEntries, nil
}

func ValidateURL(url string, maxLatency time.Duration) (bool, error) {
	fmt.Printf("正在验证: %s\n", url)
	client := &http.Client{
		Timeout: maxLatency * 2,
		Transport: &http.Transport{
			DisableKeepAlives:     true,
			IdleConnTimeout:       maxLatency * 2,
			ResponseHeaderTimeout: maxLatency * 2,
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, nil
	}

	buffer := make([]byte, 1024)
	n, err := resp.Body.Read(buffer)
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("读取内容失败: %w", err)
	}

	return n > 0, nil
}

func GetProcess() float64 {
	processLock.RLock()
	defer processLock.RUnlock()
	if totalCount == 0 {
		return 0
	}
	return float64(processCount) / float64(totalCount) * 100
}
