package handler

import (
	"complier/compiler"
	"complier/pkg/consts"
	"complier/util"
	"fmt"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
)

type LexerMenuHandler struct {
}

func NewLexerMenuHandler() *LexerMenuHandler {
	return &LexerMenuHandler{}
}

func (l *LexerMenuHandler) LexerHandler(input *widget.Entry, output *widget.Entry, bottomOutput *widget.Entry) func() {
	return func() {
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
			pos, tokn, lit := lexer.Lex()
			if tokn == consts.EOF {
				break
			}

			result = result + fmt.Sprintf("%d:%d\t%d\t%s\n", pos.Line, pos.Column, tokn, lit)
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
