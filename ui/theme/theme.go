package theme

import (
	_ "embed"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

var (
	////go:embed 是一个编译指令，它告诉编译器在编译时将指定的文件或文件夹的内容嵌入到二进制文件中。这样做的好处是你可以将资源文件（如字体、图片等）直接包含在程序中，而不需要在运行时从外部文件系统中加载它们。
	//go:embed font/微软雅黑.ttf
	MicrosoftBlack []byte
)

type MyTheme struct{}

//var _ fyne.Theme = (*MyTheme)(nil)

// StaticName 为 font 目录下的 ttf 类型的字体文件名
func (m MyTheme) Font(fyne.TextStyle) fyne.Resource {
	return &fyne.StaticResource{
		StaticName:    "微软雅黑.ttf",
		StaticContent: MicrosoftBlack,
	}
}

func (*MyTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNameDisabled {
		return color.RGBA{255, 0, 0, 100} // 设置禁用属性下的字体颜色为红色
	}
	return theme.DefaultTheme().Color(n, v)
}

func (*MyTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (*MyTheme) Size(n fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(n)
}
