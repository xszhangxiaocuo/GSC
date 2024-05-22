package handler

import (
	"complier/compiler"
	"complier/pkg/consts"
	"complier/pkg/logger"
	"complier/util"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
)

type MenuHandler struct {
	LexerFlag    bool               // 标记是否已经运行词法分析且没有错误
	ParserFlag   bool               // 标记是否已经运行语法分析且没有错误
	AnalyserFlag bool               // 标记是否已经运行语义分析且没有错误
	Parser       *compiler.Parser   // 语法分析器
	Analyser     *compiler.Analyser // 语义分析器
	QuaForm      *util.QuaFormList  // 四元式列表
	Target       *compiler.Target   // 目标代码生成器
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
				handler.Parser.Token = append(handler.Parser.Token, util.TokenNode{Pos: pos, Type: tokenid, Value: token})
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
		handler.ParserFlag = false
		if !handler.LexerFlag {
			dialog.ShowInformation("语法分析", "请先运行通过词法分析！", window)
			return
		}
		tree := handler.Parser.StartParse()
		output.SetText(tree)

		errs := len(handler.Parser.Logger.Errs)
		msg := fmt.Sprintf("---------语法分析完成---------\n%d error(s)\n\n", errs)

		if errs != 0 {
			msg += fmt.Sprintf("行:列\t\t种别码\ttoken值\t错误信息\n")
			for _, err := range handler.Parser.Logger.Errs {
				msg += err
			}
		}

		bottomOutput.SetText(msg)
		handler.LexerFlag = false
		if errs == 0 {
			handler.ParserFlag = true
		}
	}
}

func (handler *MenuHandler) AnalysierHandler(input *widget.Entry, output *widget.Entry, bottomOutput *widget.Entry, window fyne.Window) func() {
	return func() {
		handler.AnalyserFlag = false
		if !handler.ParserFlag {
			dialog.ShowInformation("语义分析", "请先运行通过语法分析！", window)
			return
		}
		handler.Analyser = compiler.NewAnalyser(handler.Parser.AST)
		handler.Analyser.StartAnalyse()
		handler.QuaForm = handler.Analyser.Qf
		result := handler.Analyser.SymbolTable.String() + "\n\n" + handler.Analyser.Qf.PrintQuaFormList()
		output.SetText(result)

		errs := len(handler.Analyser.Logger.Errs)
		msg := fmt.Sprintf("---------语义分析完成---------\n%d error(s)\n\n", errs)

		if errs != 0 {
			msg += fmt.Sprintf("行:列\t\t种别码\ttoken值\t错误信息\n")
			for _, err := range handler.Analyser.Logger.Errs {
				msg += err
			}
		}

		bottomOutput.SetText(msg)
		handler.LexerFlag = false
		handler.ParserFlag = false
		if errs == 0 {
			handler.AnalyserFlag = true
		}
	}
}

func (handler *MenuHandler) TargetHandler(input *widget.Entry, output *widget.Entry, bottomOutput *widget.Entry, window fyne.Window) func() {
	return func() {
		if !handler.AnalyserFlag {
			dialog.ShowInformation("目标代码生成", "请先运行通过语义分析！", window)
			return
		}
		handler.Target = compiler.NewTarget(handler.QuaForm, handler.Analyser.SymbolTable)
		handler.Target.GenerateAsmCode()
		result := handler.Target.Asm.String()
		output.SetText(result)

		errs := len(handler.Analyser.Logger.Errs)
		msg := fmt.Sprintf("---------目标代码生成完成---------\n%d error(s)\n\n", errs)

		if errs != 0 {
			msg += fmt.Sprintf("行:列\t\t种别码\ttoken值\t错误信息\n")
			for _, err := range handler.Analyser.Logger.Errs {
				msg += err
			}
		}
		msg += "\n\n" + handler.Analyser.Qf.PrintQuaFormList()
		bottomOutput.SetText(msg)
		handler.LexerFlag = false
		handler.ParserFlag = false
	}
}

func (handler *MenuHandler) AlgorithmHandler(myApp fyne.App, input *widget.Entry, output *widget.Entry, bottomOutput *widget.Entry, window fyne.Window) func() {
	return func() {
		mainWindow := myApp.NewWindow("DAG Window")

		// 创建左边的输入框
		leftEntry := widget.NewEntry()
		leftEntry.SetPlaceHolder("Left Entry")

		// 创建右边的两个输入框
		topRightEntry := widget.NewEntry()
		topRightEntry.SetPlaceHolder("Top Right Entry")
		bottomRightEntry := widget.NewEntry()
		bottomRightEntry.SetPlaceHolder("Bottom Right Entry")

		dagOptimizeButton := widget.NewButton("DAG优化", func() {
			//  TODO：调用 DAG 优化函数
			dag := compiler.NewDAG(nil)
			//  1. 从左边的输入框中获取输入的四元式代码
			inputQf := leftEntry.Text
			if inputQf == "" {
				dialog.ShowInformation("DAG优化", "请输入四元式代码", window)
				return
			}
			//  2. 调用 DAG 优化函数，传入四元式代码
			dag.StartDAG(inputQf)
			//  3. 将优化后的四元式代码分别显示在右边的两个输入框中
			topRightEntry.SetText(dag.PrintBasicBlocks())
			bottomRightEntry.SetText(dag.DAGQf.PrintQuaFormList())

			// 显示一个通知，表示优化已完成
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "DAG优化",
				Content: "DAG优化已完成",
			})
		})

		openFileButton := widget.NewButton("打开文件", func() {
			leftEntry.SetText(string(util.ReadFile(util.OpenFIle())))
		})

		// 将右边的两个输入框上下布局
		rightContainer := container.NewVSplit(topRightEntry, bottomRightEntry)
		rightContainer.Offset = 0.5 // 设置上下布局的初始分割比例

		// 将左边的输入框和右边的容器左右布局
		splitContainer := container.NewHSplit(leftEntry, rightContainer)
		splitContainer.Offset = 0.5 // 设置左右布局的初始分割比例

		// 创建包含输入框和按钮的主容器
		mainContainer := container.NewBorder(openFileButton, dagOptimizeButton, nil, nil, splitContainer)

		// 设置主窗口内容
		mainWindow.SetContent(mainContainer)

		// 设置主窗口尺寸并显示
		mainWindow.Resize(fyne.NewSize(600, 400))
		mainWindow.Show()
	}
}
