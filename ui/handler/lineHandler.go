package handler

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type LineHandler struct {
	Flag bool //标记当前的lineHandler要处理的文件是否已经存在行号
}

// GlobalLineHandler 全局只用一个LineHandler
var GlobalLineHandler = &LineHandler{false}

// AddLine 添加行号
func (l *LineHandler) AddLine(context []byte) string {
	l.Flag = true
	result := make([]byte, 0)
	line := 1
	result = append(result, []byte("1:\t")...)
	for _, ch := range context {
		if ch == '\n' {
			line++
			result = append(result, []byte(fmt.Sprintf("\n%d:\t", line))...)
		} else {
			result = append(result, ch)
		}
	}
	return string(result)
}

// DelLine 移除行号
func (l *LineHandler) DelLine(context []byte) string {
	l.Flag = false
	result := make([]byte, 0)
	for i := 0; i < len(context); i++ {
		ch := context[i]
		if i < 3 { // 前三个字节为第一行添加的行号
			continue
		}
		if ch == '\n' { // 换行跳过三个加入的行号字节
			i += 3
		}

		result = append(result, ch)
	}
	return string(result)
}

// SetAddLineText 添加行号并输出到entry显示
func (l *LineHandler) SetAddLineText(entry *widget.Entry, window fyne.Window) func() {
	return func() {
		context := entry.Text
		if len(context) == 0 {
			dialog.ShowInformation("添加行号", "文件内容不能为空！", window)
			return
		}
		if l.Flag {
			dialog.ShowInformation("添加行号", "行号已经添加！(要手动编辑左输入框源码时请先移除行号！！！)", window)
			return
		}
		entry.SetText(l.AddLine([]byte(context)))
	}
}

// SetDelLineText 移除行号并输出到entry显示
func (l *LineHandler) SetDelLineText(entry *widget.Entry, window fyne.Window) func() {
	return func() {
		context := entry.Text
		if len(context) == 0 {
			dialog.ShowInformation("移除行号", "文件内容不能为空！", window)
			return
		}
		if !l.Flag {
			dialog.ShowInformation("移除行号", "行号已经移除！", window)
			return
		}
		entry.SetText(l.DelLine([]byte(context)))
	}
}
