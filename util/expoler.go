package util

import (
	"fmt"
	"github.com/sqweek/dialog"
	"os"
)

// OpenFIle 选取要打开的文件，返回文件的绝对路径
func OpenFIle() string {
	filePath, err := dialog.File().Load()
	if err != nil {
		fmt.Println("Error opening file dialog:", err)
		return ""
	}

	fmt.Println("Selected file:", filePath)
	return filePath
}

// ReadFile 读取文件返回文件内容
func ReadFile(path string) string {
	file, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return string(file)
}
