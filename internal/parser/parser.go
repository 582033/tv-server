package parser

import (
	"bufio"
	"strings"

	"tv-server/internal/model"
)

func ParseM3U(content string) []model.M3UEntry {
	var entries []model.M3UEntry
	var currentMetadata string

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#EXTM3U") {
			entries = append(entries, model.M3UEntry{Metadata: line})
			continue
		}
		if strings.HasPrefix(line, "#EXTINF") {
			currentMetadata = line
			continue
		}
		if strings.HasPrefix(line, "http") {
			entries = append(entries, model.M3UEntry{
				Metadata: currentMetadata,
				URL:      line,
			})
		}
	}
	return entries
}
