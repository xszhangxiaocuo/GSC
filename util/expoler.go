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

// SaveFile 保存文件
func SaveFile(content string, path string) error {
	// 创建或打开文件
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("无法创建文件:", err)
		return err
	}
	defer file.Close()

	// 写入数据到文件
	_, err = file.WriteString(content)
	if err != nil {
		fmt.Println("无法写入文件:", err)
		return err
	}

	fmt.Println("文件保存成功!")
	return nil
}
