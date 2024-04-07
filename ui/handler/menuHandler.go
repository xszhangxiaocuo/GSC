package handler

import (
	"complier/compiler"
	"complier/pkg/consts"
	"complier/pkg/logger"
	"complier/util"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
)

type MenuHandler struct {
	LexerFlag  bool //标记是否已经运行词法分析且没有错误
	ParserFlag bool //标记是否已经运行语法分析且没有错误
	Parser     *compiler.Parser
}

func NewMenuHandler() *MenuHandler {
	return &MenuHandler{}
}

func (handler *MenuHandler) LexerHandler(input *widget.Entry, output *widget.Entry, bottomOutput *widget.Entry, window fyne.Window) func() {
	return func() {
		handler.LexerFlag = false
		if GlobalLineHandler.Flag { //行号存在会影响词法分析
			dialog.ShowInformation("词法分析", "请先移除行号！", window)
			return
		}

		lexLogger := logger.NewLogger()
		result := "行:列\t\t种别码\t\ttoken值\n"
		tempPath := "pkg/temp/temp.txt"
		if len(input.Text) != 0 { //内容不为空，保存输入框的内容为临时文件
			tempText := input.Text
			if []byte(tempText)[len(tempText)-1] != '\n' { //保证最后一个字节是换行，避免出现当字符出现在最后时无法被识别的情况
				tempText += "\n"
			}
			if err := util.SaveFile(tempText, tempPath); err != nil {
				log.Println(err.Error())
				return
			}
		}

		file, err := os.Open(tempPath)
		if err != nil {
			log.Println(err.Error())
		}

		lexer := compiler.NewLexer(file)
		handler.Parser = compiler.NewParser()
		for {
			pos, tokenid, token, lexerr := lexer.Lex()

			if tokenid == consts.ILLEGAL { //当前识别结果不合法
				lexLogger.AddErr(fmt.Sprintf("%d:%d\t\t%d\t\t%s\n", pos.Line, pos.Column, tokenid, token))
			}

			if tokenid == consts.EOF || lexerr != nil {
				break
			}

			if tokenid != consts.TokenMap["//"] && tokenid != consts.TokenMap["/**/"] && tokenid != consts.ILLEGAL { //忽略注释和错误
				result = result + fmt.Sprintf("%d:%d\t\t%d\t\t\t%s\n", pos.Line, pos.Column, tokenid, token)
				handler.Parser.Token = append(handler.Parser.Token, compiler.TokenNode{Pos: pos, Type: tokenid, Value: token})
			}

		}
		output.SetText(result)
		msg := fmt.Sprintf("---------词法分析完成---------\n%d error(s)\n", len(lexLogger.Errs))
		for _, e := range lexLogger.Errs {
			msg += e
		}
		bottomOutput.SetText(msg)
		if len(lexLogger.Errs) == 0 { //词法分析结束且没有错误
			handler.LexerFlag = true
		}

		content := output.Text
		path := fmt.Sprintf("pkg/saveFile/lex/%s.txt", util.GetTIme())
		err = util.SaveFile(content, path)
		if err != nil {
			log.Print(err.Error())
		}
	}
}

func (handler *MenuHandler) ParserHandler(input *widget.Entry, output *widget.Entry, bottomOutput *widget.Entry, window fyne.Window) func() {
	return func() {
		if !handler.LexerFlag {
			dialog.ShowInformation("语法分析", "请先运行通过词法分析！", window)
			return
		}
		tree := handler.Parser.StartParse()
		output.SetText(tree)

		msg := fmt.Sprintf("---------语法分析完成---------\n%d error(s)\n\n行:列\t\t种别码\ttoken值\t错误信息\n", len(handler.Parser.Logger.Errs))
		for _, err := range handler.Parser.Logger.Errs {
			msg += err
		}
		bottomOutput.SetText(msg)
	}
}
