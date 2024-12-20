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
		fmt.Printf("创建文件失败: %v\n", err)
		return err
	}
	defer file.Close()

	fmt.Printf("开始写入文件 %s，共 %d 个条目\n", filename, len(entries))
	writer := bufio.NewWriter(file)

	for i, entry := range entries {
		if entry.Metadata != "" {
			fmt.Fprintln(writer, entry.Metadata)
		}
		if entry.URL != "" {
			fmt.Fprintln(writer, entry.URL)
		}

		if (i+1)%100 == 0 {
			fmt.Printf("已处理 %d/%d 个条目\n", i+1, len(entries))
		}
	}

	err = writer.Flush()
	if err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		return err
	}

	fmt.Printf("文件写入完成: %s\n", filename)
	return nil
}
