package consts

type Token int

// 关键字
const (
	EOF     = 0
	ILLEGAL = -1
	CHAR    = iota + 101
	INT
	FLOAT
	BREAK
	CONST
	RETURN
	VOID
	CONTINUE
	DO
	WHILE
	IF
	ELSE
	FOR
)

// 界符
const (
	LEFTBRACE = iota + 301
	RIGHTBRACE
	SEMICOLON
	COMMA
)

// 单词类别
const (
	INTEGER = iota + 400
	BIN
	OCT
	HEX

	CHARACTER   = 500
	STRING      = 600
	IDENTIFIER  = 700
	FLOATNUMBER = 800
	EXPONENT    = 900
)

// 运算符
const (
	LEFTSMALLBRACKET = iota + 201
	RIGHTSMALLBRACKET
	LEFTMIDBRACKET
	RIGHTMIDBRACKET
	EXCLAMATIONPOINT
	MULTIPLESIGN
	DIVISIONSIGN
	PERCENT
	PLUS
	MINUS
	LESSTHANSIGN
	LESSTHANEQUALSIGN
	GREATERTHANSIGN
	GREATERTHANEQUALSIGN
	EQUAL
	UNEQUAL
	AND
	OR
	EVALUATION
	DOT
)

// 注释符
const (
	singlecomment = iota + 10001
	leftmulticomment
	rightmulticomment
)

var TokenMap = map[string]Token{
	"EOF":     EOF,     //文件结束
	"ILLEGAL": ILLEGAL, //非法格式
	//关键字
	"char":     CHAR,
	"int":      INT,
	"float":    FLOAT,
	"break":    BREAK,
	"const":    CONST,
	"return":   RETURN,
	"void":     VOID,
	"continue": CONTINUE,
	"do":       DO,
	"while":    WHILE,
	"if":       IF,
	"else":     ELSE,
	"for":      FOR,
	//界符
	"{": LEFTBRACE,
	"}": RIGHTBRACE,
	";": SEMICOLON,
	",": COMMA,
	//类型
	"integer":     INTEGER,     //整型
	"bin":         BIN,         //二进制
	"oct":         OCT,         //八进制
	"hex":         HEX,         //十六进制
	"character":   CHARACTER,   //字符
	"string":      STRING,      //字符串
	"identifier":  IDENTIFIER,  //标识符
	"floatnumber": FLOATNUMBER, //浮点数
	"exponent":    EXPONENT,    //指数形式的数
	//运算符
	"(":  LEFTSMALLBRACKET,
	")":  RIGHTSMALLBRACKET,
	"[":  LEFTMIDBRACKET,
	"]":  RIGHTMIDBRACKET,
	"!":  EXCLAMATIONPOINT,
	"*":  MULTIPLESIGN,
	"/":  DIVISIONSIGN,
	"%":  PERCENT,
	"+":  PLUS,
	"-":  MINUS,
	"<":  LESSTHANSIGN,
	"<=": LESSTHANEQUALSIGN,
	">":  GREATERTHANSIGN,
	">=": GREATERTHANEQUALSIGN,
	"==": EQUAL,
	"!=": UNEQUAL,
	"&&": AND,
	"||": OR,
	"=":  EVALUATION,
	".":  DOT,
	//注释
	"//": singlecomment,
	"/*": leftmulticomment,
	"*/": rightmulticomment,
}
