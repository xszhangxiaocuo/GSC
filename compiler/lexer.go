package compiler

import (
	"bufio"
	"complier/pkg/consts"
	"complier/util"
	"io"
	"log"
	"unicode"
)

// Lexer 词法分析器当前状态
type Lexer struct {
	pos        util.Position
	reader     *bufio.Reader
	numberFlag int //识别数字时标识当前是几进制
}

// NewLexer 传入源文件reader创建一个Lexer
func NewLexer(reader io.Reader) *Lexer {
	return &Lexer{
		pos:    util.Position{1, 0},
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

// isOperator 判断是否是运算符
func (l *Lexer) isOperator(r rune) bool {
	if r == '+' || r == '-' || r == '*' || r == '/' || r == '%' || r == '>' || r == '<' || r == '=' || r == '&' || r == '|' || r == '!' || r == '(' || r == ')' || r == '[' || r == ']' {
		return true
	}
	return false
}

// isDelimiters 判断是否是界符
func (l *Lexer) isDelimiters(r rune) bool {
	if r == '{' || r == '}' || r == ';' || r == ',' {
		return true
	}
	return false
}

// isSpace 判断是否是空符号
func (l *Lexer) isSpace(r rune) bool {
	if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
		return true
	}
	return false
}

// isLetter 识别是否是字母
func (l *Lexer) isLetter(r rune) bool {
	if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
		return true
	}
	return false
}

// isFinish 判断当前token是否识别完
func (l *Lexer) isFinish(peek rune) bool {
	if l.isSpace(peek) || l.isOperator(peek) || l.isDelimiters(peek) {
		return true
	}
	return false
}

func (l *Lexer) isNumber(n rune) bool {
	switch l.numberFlag {
	case 2:
		return n == '0' || n == '1'
	case 8:
		return n >= '0' && n <= '7'
	case 10:
		return n >= '0' && n <= '9'
	case 16:
		return (n >= '0' && n <= '9') || (n >= 'a' && n <= 'f') || (n >= 'A' && n <= 'F')
	}
	return false
}

// peek 查看下一个字符，仅查看不改变指针位置
func (l *Lexer) peek(n int) ([]byte, error) {
	r, err := l.reader.Peek(n)
	if err != nil {
		if err != io.EOF { //文件末尾
			log.Println(err)
		}
		return r, err
	}
	return r, nil
}

// Lex 一个字符一个字符扫描，识别出一个token后返回行列位置，token的值和编码
func (l *Lexer) Lex() (util.Position, consts.Token, string, error) {
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
		startPos := l.pos
		switch r {
		case '\n': //换行
			l.lineFeed()
		case '{':
			return l.pos, consts.TokenMap["{"], "{", nil
		case '}':
			return l.pos, consts.TokenMap["}"], "}", nil
		case ';':
			return l.pos, consts.TokenMap[";"], ";", nil
		case ',':
			return l.pos, consts.TokenMap[","], ",", nil
		case '(':
			return l.pos, consts.TokenMap["("], "(", nil
		case ')':
			return l.pos, consts.TokenMap[")"], ")", nil
		case '[':
			return l.pos, consts.TokenMap["["], "[", nil
		case ']':
			return l.pos, consts.TokenMap["]"], "]", nil
		default:
			if unicode.IsSpace(r) || r == '\r' || r == '\t' { //如果当前字符是空格,\r,\t就跳过继续扫描下一个字符
				continue
			} else if unicode.IsDigit(r) { //数字
				startPos = l.pos
				l.backup()
				tokenid, token = l.lexNumber()
				return startPos, tokenid, token, nil
			} else if r == '_' || l.isLetter(r) {
				startPos = l.pos
				l.backup()
				tokenid, token = l.lexIDKey()
				return startPos, tokenid, token, nil
			} else if r == '/' {
				startPos = l.pos
				l.backup()
				tokenid, token = l.lexDivision()
				return startPos, tokenid, token, nil
			} else if r == '\'' {
				startPos = l.pos
				l.backup()
				tokenid, token = l.lexChar()
				return startPos, tokenid, token, nil
			} else if r == '"' {
				startPos = l.pos
				l.backup()
				tokenid, token = l.lexString()
				return startPos, tokenid, token, nil
			} else if l.isOperator(r) {
				startPos = l.pos
				_, tid, t := l.lexOpe(r)
				return startPos, tid, t, nil
			} else {
				startPos = l.pos
				l.backup()
				tokenid, token = l.lexIllegal()
				return startPos, consts.TokenMap["ILLEGAL"], token, nil
			}
		}
	}
}

// lexOpe 识别运算符
func (l *Lexer) lexOpe(r rune) (bool, consts.Token, string) {
	var tokenid consts.Token
	token := ""
	peeks, err := l.peek(1)
	peek := rune(peeks[0])
	if err != nil {
		if err == io.EOF { //文件末尾
			tokenid = consts.TokenMap["EOF"]
		} else {
			tokenid = consts.TokenMap["ILLEGAL"]
			log.Println(err)
		}
		return false, tokenid, token
	}

	switch r {
	case '+':
		token += string(r)
		tokenid = consts.TokenMap["+"]
		if peek == '+' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap["++"]
		} else if peek == '=' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap["+="]
		}
		return true, tokenid, token
	case '-':
		token += string(r)
		tokenid = consts.TokenMap["-"]
		if peek == '-' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap["--"]
		} else if peek == '=' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap["-="]
		}
		return true, tokenid, token
	case '*':
		token += string(r)
		tokenid = consts.TokenMap["*"]
		if peek == '=' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap["*="]
		}
		return true, tokenid, token
	case '%':
		token += string(r)
		tokenid = consts.TokenMap["%"]
		if peek == '=' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap["%="]
		}
		return true, tokenid, token
	case '!':
		token += string(r)
		tokenid = consts.TokenMap["!"]
		if peek == '=' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap["!="]
		}
		return true, tokenid, token
	case '>':
		token += string(r)
		tokenid = consts.TokenMap[">"]
		if peek == '=' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap[">="]
		}
		return true, tokenid, token
	case '<':
		token += string(r)
		tokenid = consts.TokenMap["<"]
		if peek == '=' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap["<="]
		}
		return true, tokenid, token
	case '&':
		token += string(r)
		tokenid = consts.TokenMap["&"]
		if peek == '&' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap["&&"]
		} else if peek == '=' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap["&="]
		}
		return true, tokenid, token
	case '|':
		token += string(r)
		tokenid = consts.TokenMap["|"]
		if peek == '|' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap["||"]
		} else if peek == '=' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap["|="]
		}
		return true, tokenid, token
	case '=':
		token += string(r)
		tokenid = consts.TokenMap["="]
		if peek == '=' {
			l.reader.ReadRune()
			token += string(peek)
			tokenid = consts.TokenMap["=="]
		}
		return true, tokenid, token
	default:
		return false, consts.TokenMap["ILLEGAL"], ""
	}
}

// lexSpecificNumber 根据不同进制进行数字的识别
func (l *Lexer) lexSpecificNumber(state int) (consts.Token, string) {
	var tokenid consts.Token
	token := ""

	for state != -1 {
		peeks, err := l.peek(1)
		peek := rune(peeks[0])
		r, _, err := l.reader.ReadRune()
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
			if l.isNumber(r) {
				token += string(r)
			} else if r == '.' { //浮点数
				state = 1
				token += string(r)
			} else if r == 'e' || r == 'E' || (l.numberFlag == 16 && (r == 'p' || r == 'P')) { //指数形式
				state = 3
				token += string(r)
			} else if l.isFinish(peek) { //一个整数读取完成
				state = -1
				l.backup()
				if tokenid != consts.TokenMap["ILLEGAL"] {
					tokenid = consts.TokenMap["integer"]
				}
			} else {
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}

		case 1:
			if l.isNumber(r) {
				state = 2
				token += string(r)
			} else {
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}

		case 2:
			if l.isNumber(r) {
				token += string(r)
			} else if r == 'e' || r == 'E' || (l.numberFlag == 16 && (r == 'p' || r == 'P')) { //指数形式
				state = 3
				token += string(r)
			} else if l.isFinish(peek) { //一个小数读取完成
				state = -1
				l.backup()
				if tokenid != consts.TokenMap["ILLEGAL"] {
					tokenid = consts.TokenMap["floatnumber"]
				}
			} else {
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}

		case 3:
			if r == '+' || r == '-' || l.isNumber(r) {
				state = 4
				token += string(r)
			} else {
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}

		case 4:
			if l.isNumber(r) {
				token += string(r)
			} else if l.isFinish(peek) { //一个指数形式的数读取完成
				state = -1
				l.backup()
				if tokenid != consts.TokenMap["ILLEGAL"] {
					tokenid = consts.TokenMap["floatnumber"]
				}
			} else {
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}
		}
	}
	return tokenid, token
}

// lexNumber 扫描一串数字
func (l *Lexer) lexNumber() (consts.Token, string) {
	var tokenid consts.Token
	token := ""
	state := 0
	for state != -1 {
		peeks, err := l.peek(1)
		peek := rune(peeks[0])
		r, _, err := l.reader.ReadRune()
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
				state = 1
				token += string(r)
			} else if unicode.IsDigit(r) { //十进制数
				state = -1
				l.numberFlag = 10 //标记为十进制数
				l.backup()
				tokenid, token = l.lexSpecificNumber(0)
			} else {
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}

		case 1:
			if r >= '0' && r <= '7' { //八进制
				state = -1
				l.numberFlag = 8 //标记为八进制数
				l.backup()
				tokenid, token = l.lexSpecificNumber(0)
				token = "0" + token
			} else if r == 'x' || r == 'X' { //十六进制
				state = -1
				l.numberFlag = 16 //标记为十六进制数
				tokenid, token = l.lexSpecificNumber(0)
				token = "0x" + token
			} else if r == 'b' || r == 'B' { //二进制
				state = -1
				l.numberFlag = 2 //标记为二进制数
				tokenid, token = l.lexSpecificNumber(0)
				token = "0b" + token
			} else if r == '.' { //0.xxx 为十进制小数
				state = -1
				l.numberFlag = 10 //标记为十进制数
				tokenid, token = l.lexSpecificNumber(1)
				token = "0." + token
			} else if l.isFinish(peek) { //整数0
				state = -1
				l.backup()
				if tokenid != consts.TokenMap["ILLEGAL"] {
					tokenid = consts.TokenMap["integer"]
				}
			} else {
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}
		}
	}
	return tokenid, token
}

// lexIDKey 识别标识符和关键字
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
			if r == '_' || l.isLetter(r) { //标识符必须以字母或'_'组成
				state = 1 //转换为状态1
				token += string(r)

			} else {
				token += string(r)
				l.backup() //回退一个字符
				tokenid = consts.TokenMap["ILLEGAL"]
			}
		case 1:
			if l.isSpace(r) || l.isOperator(r) || l.isDelimiters(r) { //合法分隔符
				state = -1
				l.backup()
			} else if !(r == '_' || l.isLetter(r) || unicode.IsDigit(r)) { //非下划线，非字母，非数字
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			} else {
				token += string(r)
			}
		}
	}
	if tokenid != consts.TokenMap["ILLEGAL"] {
		if t, ok := consts.TokenMap[token]; ok {
			tokenid = t
		} else {
			tokenid = consts.TokenMap["identifier"]
		}
	}

	return tokenid, token
}

// lexDivision 识别除号，单行注释和多行注释
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
				} else {
					tokenid = consts.TokenMap["ILLEGAL"]
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

// lexChar 识别一个字符
func (l *Lexer) lexChar() (consts.Token, string) {
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
			if r == '\'' {
				state = 1
				token += string(r)
			} else {
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}

		case 1:
			if r == '\\' { //反斜杠'\'
				state = 3
				token += string(r)
			} else if r == '\n' { //字符内不允许换行，换行需要输入转义符\n
				state = -1
				l.backup()
				tokenid = consts.TokenMap["ILLEGAL"]
			} else {
				state = 2
				token += string(r)
			}
		case 2:
			if r == '\'' { //识别为字符
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["character"]
			} else if r == '\n' { //字符串内不允许换行，换行需要输入转义符\n
				state = -1
				l.backup()
				tokenid = consts.TokenMap["ILLEGAL"]
			} else { //非法字符
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}
		case 3:
			if r == 'n' || r == 'r' || r == 't' || r == '\'' || r == '\\' { //有效的转义
				state = 2
				token += string(r)
			} else if r == '\n' { //字符串内不允许换行，换行需要输入转义符\n
				state = -1
				l.backup()
				tokenid = consts.TokenMap["ILLEGAL"]
			} else { //无效的转义
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}
		}
	}

	return tokenid, token
}

// lexString 识别一个字符串
func (l *Lexer) lexString() (consts.Token, string) {
	var tokenid consts.Token
	token := ""
	state := 0 //初始状态
	for state != -1 {
		r, _, err := l.reader.ReadRune() //读取一个字节
		if err != nil {
			if err == io.EOF { //文件末尾
				if len(token) == 0 {
					tokenid = consts.TokenMap["EOF"]
				} else { //文件末尾一定是换行符\n，如果读到EOF说明该字符串缺少一边双引号
					tokenid = consts.TokenMap["ILLEGAL"]
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
			if r == '"' {
				state = 1
				token += string(r)
			} else {
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}
		case 1:
			if r == '\\' { //反斜杠'\'
				state = 2
				token += string(r)
			} else if r == '"' { //字符串正常结束
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["stringer"]
			} else if r == '\n' { //字符串内不允许换行，换行需要输入转义符\n
				state = -1
				l.backup()
				tokenid = consts.TokenMap["ILLEGAL"]
			} else {
				token += string(r)
			}

		case 2:
			if r == 'n' || r == 'r' || r == 't' || r == '"' || r == '\\' { //有效的转义
				state = 1
				token += string(r)
			} else if r == '\n' { //字符串内不允许换行，换行需要输入转义符\n
				state = -1
				l.backup()
				tokenid = consts.TokenMap["ILLEGAL"]
			} else { //无效的转义
				state = -1
				token += string(r)
				tokenid = consts.TokenMap["ILLEGAL"]
			}
		}
	}

	return tokenid, token
}

// lexIllegal 识别一个非法token
func (l *Lexer) lexIllegal() (consts.Token, string) {
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
			if l.isSpace(r) || l.isOperator(r) || l.isDelimiters(r) { //遇到合法分割符结束扫描
				state = -1
				l.backup() //回退一个字符
				tokenid = consts.TokenMap["ILLEGAL"]
			} else {
				token += string(r)
			}

		}
	}

	return tokenid, token
}
