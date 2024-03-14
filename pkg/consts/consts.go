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
	INTEGER    = 400
	CHARACTER  = 500
	STRING     = 600
	IDENTIFIER = 700
	REALNUMBER = 800
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

var TokenMap = map[string]Token{
	"EOF":        EOF,
	"ILLGEAL":    ILLEGAL,
	"char":       CHAR,
	"int":        INT,
	"float":      FLOAT,
	"break":      BREAK,
	"const":      CONST,
	"return":     RETURN,
	"void":       VOID,
	"continue":   CONTINUE,
	"do":         DO,
	"while":      WHILE,
	"if":         IF,
	"else":       ELSE,
	"for":        FOR,
	"{":          LEFTBRACE,
	"}":          RIGHTBRACE,
	";":          SEMICOLON,
	",":          COMMA,
	"integer":    INTEGER,
	"character":  CHARACTER,
	"string":     STRING,
	"identifier": IDENTIFIER,
	"realnumber": REALNUMBER,
	"(":          LEFTSMALLBRACKET,
	")":          RIGHTSMALLBRACKET,
	"[":          LEFTMIDBRACKET,
	"]":          RIGHTMIDBRACKET,
	"!":          EXCLAMATIONPOINT,
	"*":          MULTIPLESIGN,
	"/":          DIVISIONSIGN,
	"%":          PERCENT,
	"+":          PLUS,
	"-":          MINUS,
	"<":          LESSTHANSIGN,
	"<=":         LESSTHANEQUALSIGN,
	">":          GREATERTHANSIGN,
	">=":         GREATERTHANEQUALSIGN,
	"==":         EQUAL,
	"!=":         UNEQUAL,
	"&&":         AND,
	"||":         OR,
	"=":          EVALUATION,
	".":          DOT,
}
