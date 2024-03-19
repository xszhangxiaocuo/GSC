package compiler

import (
	"bufio"
	"complier/pkg/consts"
	"io"
	"log"
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

// Lex 一个字符一个字符扫描，识别出一个token后返回行列位置，token的值和编码
func (l *Lexer) Lex() (Position, consts.Token, string, error) {
	for {
		var tokenid consts.Token
		var token string
		//读取一个字节的utf8字符
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF { //文件末尾
				return l.pos, consts.EOF, "", nil
			}
			panic(err)
		}

		l.pos.Column++

		switch r {
		case '\n': //换行
			l.resetPosition()
		case ';':
			return l.pos, consts.TokenMap[";"], ";", nil
		case '+':
			return l.pos, consts.TokenMap["+"], "+", nil
		case '-':
			return l.pos, consts.TokenMap["-"], "-", nil
		case '*':
			return l.pos, consts.TokenMap["*"], "*", nil
		case '%':
			return l.pos, consts.TokenMap["%"], "%", nil
		case '=':
			return l.pos, consts.TokenMap["="], "=", nil
		default:
			if unicode.IsSpace(r) { //如果当前字符是空格就跳过继续扫描下一个字符
				continue
			} else if unicode.IsDigit(r) { //数字
				startPos := l.pos
				l.backup()
				tokenid, token = l.lexInt() //读取一个整数integer
				return startPos, tokenid, token, nil
			} else if r == '_' || unicode.IsLetter(r) {
				startPos := l.pos
				l.backup()
				tokenid, token = l.LexIDKey()
				return startPos, tokenid, token, nil
			} else if r == '/' {
				startPos := l.pos
				l.backup()
				tokenid, token = l.LexDivision()
				return startPos, tokenid, token, nil
			}
		}
	}
}

//// lexInt 扫描一串int数
//func (l *Lexer) lexInt() (consts.Token, string) {
//	tokenid := consts.TokenMap["integer"]
//	token := ""
//	state := 0
//	for state != 2 {
//		r, _, err := l.reader.ReadRune() //读取一个字节
//
//		if err != nil {
//			if err == io.EOF { //文件末尾
//				if len(token) == 0 {
//					tokenid = consts.TokenMap["EOF"]
//				}
//
//			} else {
//				tokenid = consts.TokenMap["ILLEGAL"]
//				token = ""
//			}
//			return tokenid, token
//		}
//		l.pos.Column++
//		switch state {
//		case 0:
//			if r == '0' { //第一个数字为0的只能是0
//				state = 2
//			} else if unicode.IsDigit(r) {
//				state = 1
//				token += string(r)
//			}
//		}
//	}
//	return tokenid, token
//}

// lexInt 扫描一串int数
func (l *Lexer) lexInt() (consts.Token, string) {
	var tokenid consts.Token
	token := ""
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF { //文件末尾
				if len(token) == 0 {
					tokenid = consts.TokenMap["EOF"]
				} else {
					tokenid = consts.TokenMap["integer"]
				}
			} else {
				tokenid = consts.TokenMap["ILLEGAL"]
				token = ""
			}
			return tokenid, token
		}
		l.pos.Column++

		if unicode.IsDigit(r) {
			token += string(r)
		} else { //当前字符不是数字
			l.backup() //回退一个字符
			tokenid = consts.TokenMap["integer"]
			return tokenid, token
		}
	}
}

// LexIDKey 识别标识符和关键字
func (l *Lexer) LexIDKey() (consts.Token, string) {
	var tokenid consts.Token
	token := ""
	state := 0 //初始状态
	for state != 2 {
		r, _, err := l.reader.ReadRune() //读取一个字节
		l.pos.Column++
		if err != nil {
			if err == io.EOF { //文件末尾
				if len(token) == 0 {
					tokenid = consts.TokenMap["EOF"]
				}

			} else {
				tokenid = consts.TokenMap["ILLEGAL"]
				token = ""
			}
			return tokenid, token
		}
		switch state {
		case 0:
			if r == '_' || unicode.IsLetter(r) { //标识符必须以字母或'_'组成
				token += string(r)
				state = 1 //转换为状态1
			} else {
				state = 2
				l.backup() //回退一个字符
			}
		case 1:
			if !(r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)) { //非下划线，非字母，非数字转到状态2
				state = 2
				l.backup() //回退一个字符
			} else {
				token += string(r)
			}
		}
	}
	if t, ok := consts.TokenMap[token]; ok {
		tokenid = t
	} else {
		tokenid = consts.TokenMap["identifier"]
	}

	return tokenid, token
}

// LexDivision 识别除号，单行注释和多行注释
func (l *Lexer) LexDivision() (consts.Token, string) {
	var tokenid consts.Token
	token := ""
	state := 0

	for state != 3 {
		r, _, err := l.reader.ReadRune() //读取一个字节
		l.pos.Column++
		if err != nil {
			if err == io.EOF { //文件末尾
				if len(token) == 0 {
					tokenid = consts.TokenMap["EOF"]
				}
			} else {
				log.Println(err)
				tokenid = consts.TokenMap["ILLEGAL"]
				token = ""
			}
			return tokenid, token
		}
		switch state {
		case 0:
			if r == '/' {
				state = 1
				tokenid = consts.TokenMap["/"]
				token += string(r)
			}
		case 1:
			if r == '/' {
				state = 2
				tokenid = consts.TokenMap["//"]
				token += string(r)
			}
		case 2:
			if r == '\n' {
				state = 3
				l.backup()
			} else {
				token += string(r)
			}
		}
	}
	return tokenid, token
}
