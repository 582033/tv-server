package utils

import (
	"fmt"
	"time"
)

// CalculateTotalTimeToString 计算处理给定数量链接在指定并发数和单个链接超时时长情况下的理论总耗时，并返回几分几秒格式的字符串表示
// 参数说明：
// - timeoutPerLink 单个链接的超时时长，以time.Duration类型传入, 单位ms
// - concurrentTasks 并发任务数量
// - totalLinks 总的链接数量
func CalculateTotalTimeToString(timeoutPerLink time.Duration, concurrentTasks int, totalLinks int) string {
	if concurrentTasks <= 0 {
		concurrentTasks = 1
	}

	batchCount := (totalLinks + concurrentTasks - 1) / concurrentTasks
	totalTime := time.Duration(batchCount) * timeoutPerLink
	fmt.Println("batchCount:", batchCount)
	fmt.Println("totalTime:", totalTime)
	return fmt.Sprintf("%v", totalTime.Round(time.Second))
}
