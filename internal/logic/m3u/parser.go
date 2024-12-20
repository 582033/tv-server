package m3u

import (
	"bufio"
	"strings"
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
