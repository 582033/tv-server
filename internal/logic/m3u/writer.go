package m3u

import (
	"bufio"
	"fmt"
	"os"
)

func WriteToFile(entries []Entry, filename string) error {
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
