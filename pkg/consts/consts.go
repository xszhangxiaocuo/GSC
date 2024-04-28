package consts

type Token int

const (
	NULL      = "ε"
	ALL       = "@all" //全局作用域标记
	TYPEVOID  = "void"
	TYPEINT   = "int"
	TYPEFLOAT = "float"
	TYPECHAR  = "char"
	TYPEFUNC  = "func"
	TYPECONST = "const"
	TYPEVAR   = "var"
)

// 关键字
const (
	EOF     = -1
	ILLEGAL = -2
	CHAR    = iota + 101
	STRING
	INT
	FLOAT
	TRUE
	FALSE
	BREAK
	CONST
	RETURN
	VAR
	VOID
	MAIN
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
	STRINGER    = 600
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
	MULTIPLESIGNEQUAL
	DIVISIONSIGN
	DIVISIONSIGNEQUAL
	PERCENT
	PERCENTEQUAL
	PLUS
	PLUSPLUS
	PLUSEQUAL
	MINUS
	MINUSMINUS
	MINUSEQUAL
	LESSTHANSIGN
	LESSTHANEQUALSIGN
	GREATERTHANSIGN
	GREATERTHANEQUALSIGN
	EQUAL
	UNEQUAL
	AND
	OR
	SINGLEAND
	SINGLEOR
	ANDEQUAL
	OREQUAL
	EVALUATION
	DOT
)

// 注释符
const (
	SINGLECOMMENT = iota + 10001
	MULTICOMMENT
)

var TokenMap = map[string]Token{
	"EOF":     EOF,     //文件结束
	"ILLEGAL": ILLEGAL, //非法格式
	//关键字
	"char":     CHAR,
	"string":   STRING,
	"int":      INT,
	"float":    FLOAT,
	"true":     TRUE,
	"false":    FALSE,
	"break":    BREAK,
	"const":    CONST,
	"var":      VAR,
	"return":   RETURN,
	"void":     VOID,
	"main":     MAIN,
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
	"stringer":    STRINGER,    //字符串
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
	"*=": MULTIPLESIGNEQUAL,
	"/":  DIVISIONSIGN,
	"/=": DIVISIONSIGNEQUAL,
	"%":  PERCENT,
	"%=": PERCENTEQUAL,
	"+":  PLUS,
	"++": PLUSPLUS,
	"+=": PLUSEQUAL,
	"-":  MINUS,
	"--": MINUSMINUS,
	"-=": MINUSEQUAL,
	"<":  LESSTHANSIGN,
	"<=": LESSTHANEQUALSIGN,
	">":  GREATERTHANSIGN,
	">=": GREATERTHANEQUALSIGN,
	"==": EQUAL,
	"!=": UNEQUAL,
	"&&": AND,
	"||": OR,
	"&":  SINGLEAND,
	"|":  SINGLEOR,
	"&=": ANDEQUAL,
	"|=": OREQUAL,
	"=":  EVALUATION,
	".":  DOT,
	//注释
	"//":   SINGLECOMMENT,
	"/**/": MULTICOMMENT,
}

const (
	PROGRAM              string = "<程序>"
	DECLARATION          string = "<声明语句>"
	VALUE_DECLARATION    string = "<值声明>"
	CONST_DECLARATION    string = "<常量声明>"
	CONST_TYPE           string = "<常量类型>"
	CONST_TABLE          string = "<常量声明表>"
	CONST_TABLE_0        string = "<常量声明表0>"
	CONST_TABLE_1        string = "<常量声明表1>"
	CONST_TABLE_VALUE    string = "<常量声明表值>"
	VARIABLE             string = "<变量>"
	CONSTANT             string = "<常量>"
	NUM_CONSTANT         string = "<数值型常量>"
	CHAR_CONSTANT        string = "<字符型常量>"
	VARIABLE_DECL        string = "<变量声明>"
	VARIABLE_TYPE        string = "<变量类型>"
	VARIABLE_TABLE       string = "<变量声明表>"
	VARIABLE_TABLE_0     string = "<变量声明表0>"
	SINGLE_VARIABLE      string = "<单变量声明>"
	SINGLE_VARIABLE_0    string = "<单变量声明0>"
	FUNCTION_DECL_STMT   string = "<函数声明语句>"
	FUNCTION_DECL        string = "<函数声明>"
	FUNCTION_TYPE        string = "<函数类型>"
	FUNCTION_PARAMS      string = "<函数声明形参列表>"
	FUNCTION_PARAM       string = "<函数声明形参>"
	FUNCTION_PARAM_0     string = "<函数声明形参0>"
	COMPOUND_STMT        string = "<复合语句>"
	STATEMENT_TABLE      string = "<语句表>"
	STATEMENT_TABLE_0    string = "<语句表0>"
	STATEMENT            string = "<语句>"
	EXECUTION_STMT       string = "<执行语句>"
	DATA_PROCESS_STMT    string = "<数据处理语句>"
	FUNCTION_CALL_STMT   string = "<函数调用语句>"
	CONTROL_STMT         string = "<控制语句>"
	FUNCTION_CALL        string = "<函数调用>"
	ARGUMENTS            string = "<实参列表>"
	ARGUMENT             string = "<实参>"
	ARGUMENT_0           string = "<实参0>"
	IF_STMT              string = "<if语句>"
	IF_TAIL              string = "<ifTail语句>"
	IF_TAIL_0            string = "<ifTail0语句>"
	FOR_STMT             string = "<for语句>"
	WHILE_STMT           string = "<while语句>"
	DO_WHILE_STMT        string = "<DoWHILE语句>"
	RETURN_STMT          string = "<return语句>"
	RETURN_STMT_0        string = "<return语句0>"
	BREAK_STMT           string = "<break语句>"
	CONTINUE_STMT        string = "<continue语句>"
	FUNCTION_BLOCK       string = "<函数块>"
	FUNCTION_DEF         string = "<函数定义>"
	FUNCTION_PARAMS_DEF  string = "<函数定义形参列表>"
	FUNCTION_PARAM_DEF   string = "<函数定义形参>"
	FUNCTION_PARAM_0_DEF string = "<函数定义形参0>"
	ASSIGNMENT_STMT      string = "<赋值语句>"
	ASSIGNMENT_EXPR      string = "<赋值表达式>"
	ASSIGNMENT_EXPR_0    string = "<赋值表达式0>"
	BOOLEAN_EXPR         string = "<布尔表达式>"
	BOOLEAN_EXPR_0       string = "<布尔表达式0>"
	BOOLEAN_ITEM         string = "<布尔项>"
	BOOLEAN_ITEM_0       string = "<布尔项0>"
	BOOLEAN_FACTOR       string = "<布尔因子>"
	BOOLEAN_FACTOR_0     string = "<布尔因子0>"
	ARITHMETIC_EXPR      string = "<算术表达式>"
	ARITHMETIC_EXPR_0    string = "<算术表达式0>"
	TERM                 string = "<项>"
	TERM_0               string = "<项0>"
	FACTOR               string = "<因子>"
	FACTOR_0             string = "<因子0>"
	RELATION_EXPR        string = "<关系表达式>"
	RELATION_OPERATOR    string = "<关系运算符>"
)
