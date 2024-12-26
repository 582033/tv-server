package m3u

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Entry struct {
	Metadata string `json:"Metadata"`
	URL      string `json:"URL"`
}

type ParsedEntry struct {
	Channel string `json:"channel"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Logo    string `json:"logo"`
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

// ParseEntry 解析 Entry 数据并返回 ParsedEntry 列表
func ParseEntry(entries []Entry) []ParsedEntry {
	parsedEntries := make([]ParsedEntry, 0, len(entries))
	// 正则表达式匹配规则
	re := regexp.MustCompile(`#EXTINF:-1((?:\s+tvg-[^=]+="([^"]*)")*)(?:\s+group-title="([^"]+)"),([^,]+)`)

	for _, entry := range entries {
		// 使用正则表达式解析 Metadata 字段
		matches := re.FindStringSubmatch(entry.Metadata)
		// 仅当 matches 有足够的结果时才打印和处理
		if len(matches) > 0 {
			// 如果匹配结果大于 4，意味着我们可以获取到频道、标题、URL 和 Logo
			if len(matches) >= 5 {
				// 将解析结果存入 parsedEntries
				parsedEntries = append(parsedEntries, ParsedEntry{
					Channel: matches[3], // 频道名称
					Title:   matches[4], // 标题
					URL:     entry.URL,  // 原 URL
					Logo:    matches[2], // Logo URL
				})
			}
		}
	}
	return parsedEntries
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

	// 打印文件内容
	fmt.Printf("文件内容:\n%s\n", string(content))

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
