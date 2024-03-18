package ui

import (
	"complier/ui/handler"
	"complier/ui/theme"
	"complier/util"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"log"
)

var MyApp fyne.App
var MainWindow fyne.Window

func InitApp() {
	MyApp = app.New()
	MainWindow = MyApp.NewWindow("😋Go Sample Compiler")
	MainWindow.Resize(fyne.NewSize(800, 600)) // 设置窗口的固定大小

	// 设置默认字体
	MyApp.Settings().SetTheme(&theme.MyTheme{})

	// 创建两个输入框
	leftInput := widget.NewMultiLineEntry()
	leftInput.SetPlaceHolder("请输入源代码...")

	rightOutput := widget.NewMultiLineEntry()
	//设置换行模式，保证表示在单词边界处进行换行，而不是在任意字符处换行
	rightOutput.Wrapping = fyne.TextWrapWord

	// 创建一个用于输出的多行文本框，并放入滚动容器中
	bottomOutput := widget.NewMultiLineEntry()
	bottomOutput.Wrapping = fyne.TextWrapWord
	bottomScroll := container.NewScroll(bottomOutput)
	bottomScroll.SetMinSize(fyne.NewSize(0, 200)) // 设置底部滚动容器的最小高度为200

	// 创建一个网格容器，用于放置左侧和右侧输入框
	grid := container.NewGridWithColumns(2,
		container.NewScroll(leftInput),
		container.NewScroll(rightOutput),
	)

	// 创建一个边界容器，用于组织整个布局
	content := container.NewBorder(nil, bottomScroll, nil, nil, grid)

	MainWindow.SetContent(content)

	// 创建菜单项
	fileMenu := fyne.NewMenu("文件",
		fyne.NewMenuItem("打开", func() {
			leftInput.SetText(util.AddLine(util.ReadFile(util.OpenFIle())))
		}),
		fyne.NewMenuItem("保存源码文件", func() {
			file := leftInput.Text
			if len(file) == 0 {
				dialog.ShowInformation("保存失败", "文件内容不能为空！", MainWindow)
				return
			}
			path := fmt.Sprintf("pkg/saveFile/source/%s.txt", util.GetTIme())
			err := util.SaveFile(file, path)
			if err != nil {
				dialog.ShowInformation("保存失败", "文件保存失败！", MainWindow)
				log.Print(err.Error())
			} else {
				dialog.ShowInformation("保存成功", "文件保存成功！", MainWindow)
			}
		}),
		fyne.NewMenuItem("保存输出文件", func() {
			file := rightOutput.Text
			if len(file) == 0 {
				dialog.ShowInformation("保存失败", "文件内容不能为空！", MainWindow)
				return
			}
			path := fmt.Sprintf("pkg/saveFile/lex/%s.txt", util.GetTIme())
			err := util.SaveFile(file, path)
			if err != nil {
				dialog.ShowInformation("保存失败", "文件保存失败！", MainWindow)
				log.Print(err.Error())
			} else {
				dialog.ShowInformation("保存成功", "文件保存成功！", MainWindow)
			}
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("退出", func() { MyApp.Quit() }),
	)

	//TODO：完善编辑菜单选项函数
	editMenu := fyne.NewMenu("编辑",
		fyne.NewMenuItem("复制", func() { println("复制被点击了") }),
		fyne.NewMenuItem("粘贴", func() { println("粘贴被点击了") }),
	)

	//TODO：完善词法分析菜单选项函数
	lexerMenu := fyne.NewMenu("词法分析",
		fyne.NewMenuItem("词法分析器", handler.NewLexerMenuHandler().LexerHandler(leftInput, rightOutput, bottomOutput)),
	)

	//TODO：完善语法分析菜单选项函数
	parserMenu := fyne.NewMenu("语法分析",
		fyne.NewMenuItem("语法分析器", func() {
			println("语法分析器被点击了")
		}))

	//TODO：完善语义分析菜单选项函数
	analysierMenu := fyne.NewMenu("语义分析",
		fyne.NewMenuItem("语义分析器", func() {
			println("语义分析器被点击了")
		}))

	//TODO：完善中间代码菜单选项函数
	IRcodeMenu := fyne.NewMenu("中间代码",
		fyne.NewMenuItem("中间代码生成", func() {
			println("中间代码生成被点击了")
		}))

	//TODO：完善目标代码菜单选项函数
	targetcodeMenu := fyne.NewMenu("目标代码",
		fyne.NewMenuItem("目标代码生成", func() {
			println("目标代码生成被点击了")
		}))

	// 创建顶部菜单栏
	mainMenu := fyne.NewMainMenu(
		fileMenu,
		editMenu,
		lexerMenu,
		parserMenu,
		analysierMenu,
		IRcodeMenu,
		targetcodeMenu,
	)
	MainWindow.SetMainMenu(mainMenu)

	// 显示窗口
	MainWindow.ShowAndRun()
}
