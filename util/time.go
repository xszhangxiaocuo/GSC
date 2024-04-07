package util

import "time"

// GetTIme 获取当前时间并格式化为字符串
func GetTIme() string {
	t := time.Now()
	timestamp := t.Format("20060102_150405") // 格式化时间，例如20240308_125701
	return timestamp
}
