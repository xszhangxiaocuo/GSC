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

// lineFeed 换行操作
func (l *Lexer) lineFeed() {
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
			log.Println(err)
		}

		l.pos.Column++

		switch r {
		case '\n': //换行
			l.lineFeed()
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
			if unicode.IsSpace(r) || r == '\r' || r == '\t' { //如果当前字符是空格,\r,\t就跳过继续扫描下一个字符
				continue
			} else if unicode.IsDigit(r) { //数字
				startPos := l.pos
				l.backup()
				tokenid, token = l.lexNumber()
				return startPos, tokenid, token, nil
			} else if r == '_' || unicode.IsLetter(r) {
				startPos := l.pos
				l.backup()
				tokenid, token = l.lexIDKey()
				return startPos, tokenid, token, nil
			} else if r == '/' {
				startPos := l.pos
				l.backup()
				tokenid, token = l.lexDivision()
				return startPos, tokenid, token, nil
			} else {
				return l.pos, consts.TokenMap["ILLEGAL"], string(r), nil
			}
		}
	}
}

// lexNumber 扫描一串数字
func (l *Lexer) lexNumber() (consts.Token, string) {
	var tokenid consts.Token
	token := ""
	state := 0
	for state != -1 {
		r, _, err := l.reader.ReadRune() //读取一个字节
		if err != nil {
			if err == io.EOF { //文件末尾
				if len(token) == 0 {
					tokenid = consts.TokenMap["EOF"]
				}
			} else {
				tokenid = consts.TokenMap["ILLEGAL"]
				log.Println(err)
			}
			return tokenid, token
		}
		l.pos.Column++

		switch state {
		case 0:
			if r == '0' { //第一个数字为0,可能为二进制，八进制，十六进制以及0本身
				state = 6
				token += string(r)
			} else if unicode.IsDigit(r) { //十进制数
				state = 1
				token += string(r)
			} else {
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}

		case 1:
			if unicode.IsDigit(r) {
				token += string(r)
			} else if r == '.' { //浮点数
				state = 2
				token += string(r)
			} else if r == 'e' || r == 'E' { //指数形式
				state = 4
				token += string(r)
			} else { //一个整数读取完成
				state = -1
				l.backup()
				tokenid = consts.TokenMap["integer"]
			}

		case 2:
			if unicode.IsDigit(r) {
				state = 3
				token += string(r)
			} else {
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}

		case 3:
			if unicode.IsDigit(r) {
				token += string(r)
			} else if r == 'e' || r == 'E' { //指数形式
				state = 4
				token += string(r)
			} else { //非法格式
				state = -1
				l.backup()
				tokenid = consts.TokenMap["floatnumber"]
			}

		case 4:
			if r == '+' || r == '-' || unicode.IsDigit(r) {
				state = 5
				token += string(r)
			} else {
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}

		case 5:
			if unicode.IsDigit(r) {
				token += string(r)
			} else { //读取完一个指数形式的数，正常读取结束的情况下需要回退一个字符
				state = -1
				l.backup()
				tokenid = consts.TokenMap["exponent"]
			}

		case 6:
			if r >= '0' && r <= '7' { //八进制
				state = 7
				token += string(r)
			} else if r == 'x' || r == 'X' { //十六进制
				state = 8
				token += string(r)
			} else if r == 'b' || r == 'B' { //二进制
				state = 10
				token += string(r)
			} else if r == '.' { //小数0.xxx，转到状态2作为浮点数判断
				state = 2
				token += string(r)
			} else { //整数0
				state = -1
				l.backup()
				tokenid = consts.TokenMap["integer"]
			}

		case 7:
			if r >= '0' && r <= '7' {
				token += string(r)
			} else {
				state = -1
				l.backup()
				tokenid = consts.TokenMap["oct"]
			}

		case 8:
			if unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
				state = 9
				token += string(r)
			} else {
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}

		case 9:
			if unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
				token += string(r)
			} else {
				state = -1
				l.backup()
				tokenid = consts.TokenMap["hex"]
			}

		case 10:
			if r == '0' || r == '1' {
				state = 11
				token += string(r)
			} else {
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}

		case 11:
			if r == '0' || r == '1' {
				token += string(r)
			} else {
				state = -1
				l.backup()
				tokenid = consts.TokenMap["bin"]
			}
		}

	}
	return tokenid, token
}

// LexIDKey 识别标识符和关键字
func (l *Lexer) lexIDKey() (consts.Token, string) {
	var tokenid consts.Token
	token := ""
	state := 0 //初始状态
	for state != -1 {
		r, _, err := l.reader.ReadRune() //读取一个字节
		if err != nil {
			if err == io.EOF { //文件末尾
				if len(token) == 0 {
					tokenid = consts.TokenMap["EOF"]
				}
			} else {
				tokenid = consts.TokenMap["ILLEGAL"]
				log.Println(err)
			}
			return tokenid, token
		}
		l.pos.Column++

		switch state {
		case 0:
			if r == '_' || unicode.IsLetter(r) { //标识符必须以字母或'_'组成
				state = 1 //转换为状态1
				token += string(r)

			} else {
				state = -1
				l.backup() //回退一个字符
			}
		case 1:
			if !(r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)) { //非下划线，非字母，非数字转到状态2
				state = -1
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
func (l *Lexer) lexDivision() (consts.Token, string) {
	var tokenid consts.Token
	token := ""
	state := 0

	for state != -1 {
		r, _, err := l.reader.ReadRune() //读取一个字节
		l.pos.Column++
		if err != nil {
			if err == io.EOF { //文件末尾
				if len(token) == 0 {
					tokenid = consts.TokenMap["EOF"]
				}
			} else {
				tokenid = consts.TokenMap["ILLEGAL"]
				log.Println(err)
			}
			return tokenid, token
		}
		switch state {
		case 0:
			if r == '/' {
				state = 1
				token += string(r)
			} else {
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}
		case 1:
			if r == '/' {
				state = 2
				token += string(r)
			} else if r == '*' {
				state = 3
				token += string(r)
			} else { //识别为除号
				state = -1
				l.backup()
				tokenid = consts.TokenMap["/"]
			}
		case 2:
			if r == '\n' { //识别为单行注释
				state = -1
				l.backup()
				tokenid = consts.TokenMap["//"]
			} else {
				token += string(r)
			}
		case 3:
			if r == '*' {
				state = 4
				token += string(r)
			} else {
				if r == '\n' {
					l.lineFeed()
				}
				token += string(r)
			}
		case 4:
			if r == '/' { //识别为多行注释
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["/**/"]
			} else {
				state = 3
				token += string(r)
			}
		}
	}
	return tokenid, token
}
