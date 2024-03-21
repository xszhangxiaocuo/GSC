package test0

import (
	"bufio"
	"complier/pkg/consts"
	"io"
	"unicode"
)

// Position 当前读到的行列
type Position struct {
	Line   int
	Column int
}

// Lexer 词法分析器当前状态
type Lexer struct {
	pos    Position
	reader *bufio.Reader
}

// NewLexer 传入源文件reader创建一个Lexer
func NewLexer(reader io.Reader) *Lexer {
	return &Lexer{
		pos:    Position{1, 0},
		reader: bufio.NewReader(reader),
	}
}

// Lex 一个字符一个字符扫描，识别出一个token后返回行列位置，token的值和编码
func (l *Lexer) Lex() (Position, consts.Token, string) {
	for {
		//读取一个字节的utf8字符
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF { //文件末尾
				return l.pos, consts.EOF, ""
			}
			panic(err)
		}

		l.pos.Column++

		switch r {
		case '\n': //换行
			l.resetPosition()
		case ';':
			return l.pos, consts.TokenMap[";"], ";"
		case '+':
			return l.pos, consts.TokenMap["+"], "+"
		case '-':
			return l.pos, consts.TokenMap["-"], "-"
		case '*':
			return l.pos, consts.TokenMap["*"], "*"
		case '/':
			return l.pos, consts.TokenMap["/"], "/"
		case '=':
			return l.pos, consts.TokenMap["="], "="
		default:
			if unicode.IsSpace(r) { //如果当前字符是空格就跳过继续扫描下一个字符
				continue
			} else if unicode.IsDigit(r) { //数字
				startPos := l.pos
				l.backup()
				digit := l.lexInt() //读取一个整数integer
				return startPos, consts.TokenMap["integer"], digit
			} else if unicode.IsLetter(r) {
				startPos := l.pos
				l.backup()
				letters := l.lexLetters()
				return startPos, consts.TokenMap["identifier"], letters
			}
		}
	}
}

// resetPosition 换行操作
func (l *Lexer) resetPosition() {
	l.pos.Line++
	l.pos.Column = 0
}

// backup 将当前读取的位置回退到上个字符
func (l *Lexer) backup() {
	if err := l.reader.UnreadRune(); err != nil {
		panic(err)
	}
	l.pos.Column--
}

// lexInt 扫描一串int数
func (l *Lexer) lexInt() (digit string) {
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF { //文件末尾
				return
			}
			panic(err)
		}
		l.pos.Column++
		if unicode.IsDigit(r) {
			digit = digit + string(r)
		} else { //当前字符不是数字
			l.backup() //回退一个字符
			return
		}
	}
}

func (l *Lexer) lexLetters() (letters string) {
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF { //文件末尾
				return
			}
			panic(err)
		}
		l.pos.Column++
		if unicode.IsLetter(r) {
			letters = letters + string(r)
		} else { //当前字符不是字母
			l.backup() //回退一个字符
			return
		}
	}
}
