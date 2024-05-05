package logger

import (
	"complier/util"
	"fmt"
)

type Logger struct {
	Errs []string
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) AddErr(err string) {
	l.Errs = append(l.Errs, err)
}

func (l *Logger) AddParserErr(token util.TokenNode, nodeName string, msg ...string) {
	l.AddErr(fmt.Sprintf("%d:%d\t\t%d\t\t%s\t\t%s推断错误 %s\n", token.Pos.Line, token.Pos.Column, token.Type, token.Value, nodeName, msg))
}

func (l *Logger) AddAnalyseErr(token *util.TokenNode, msg ...string) {
	l.AddErr(fmt.Sprintf("%d:%d\t\t%d\t\t%s\t\t语义错误: %s\n", token.Pos.Line, token.Pos.Column, token.Type, token.Value, msg))
}
