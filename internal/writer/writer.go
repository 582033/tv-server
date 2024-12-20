package writer

import (
	"bufio"
	"fmt"
	"os"

	"tv-server/internal/model"
)

func WriteM3U(entries []model.M3UEntry, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, entry := range entries {
		if entry.Metadata != "" {
			fmt.Fprintln(writer, entry.Metadata)
		}
		if entry.URL != "" {
			fmt.Fprintln(writer, entry.URL)
		}
	}
	return writer.Flush()
}
