package handler

import (
	"complier/compiler"
	"complier/pkg/consts"
	"complier/util"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
)

type LexerMenuHandler struct {
}

func NewLexerMenuHandler() *LexerMenuHandler {
	return &LexerMenuHandler{}
}

func (l *LexerMenuHandler) LexerHandler(input *widget.Entry, output *widget.Entry, bottomOutput *widget.Entry, window fyne.Window) func() {
	return func() {
		if GlobalLineHandler.Flag { //行号存在会影响词法分析
			dialog.ShowInformation("词法分析", "请先移除行号！", window)
			return
		}
		result := ""
		tempPath := "pkg/temp/temp.txt"
		if len(input.Text) != 0 { //内容不为空
			if err := util.SaveFile(input.Text, tempPath); err != nil {
				log.Println(err.Error())
				return
			}
		}

		file, err := os.Open(tempPath)
		if err != nil {
			log.Println(err.Error())
		}

		lexer := compiler.NewLexer(file)
		for {
			pos, toknid, token, lexerr := lexer.Lex()

			if toknid == consts.EOF || lexerr != nil {
				break
			}

			result = result + fmt.Sprintf("%d:%d\t%d\t%s\n", pos.Line, pos.Column, toknid, token)
		}
		output.SetText(result)
		bottomOutput.SetText("---------词法分析完成---------\n0 error(s)")

		content := output.Text
		path := fmt.Sprintf("pkg/saveFile/lex/%s.txt", util.GetTIme())
		err = util.SaveFile(content, path)
		if err != nil {
			log.Print(err.Error())
		}
	}
}
