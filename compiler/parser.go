package compiler

import (
	"complier/pkg/consts"
	"complier/pkg/logger"
	"complier/util"
)

type Parser struct {
	Token  []util.TokenNode
	Index  int //当前的token下标
	Logger *logger.Logger
}

func NewParser() *Parser {
	return &Parser{Logger: logger.NewLogger()}
}

// StartParse 开始解析token生成语法树返回
func (p *Parser) StartParse() string {
	root := p.program()

	return util.GetTree(root)
}

// nextToken 取得下一个token
func (p *Parser) nextToken() (token util.TokenNode) {
	if p.Index < len(p.Token) {
		token = p.Token[p.Index]
		p.Index++
		return
	}
	return util.TokenNode{Type: consts.TokenMap["EOF"]}
}

// peek 查看下n个token
func (p *Parser) peek(n int) util.TokenNode {
	if p.Index+n-1 < len(p.Token) {
		return p.Token[p.Index+n-1]
	}
	return util.TokenNode{Type: consts.TokenMap["EOF"]}
}

// match 判断传入的token种别码与下一个token种别码是否匹配
func (p *Parser) match(token util.TokenNode, expectToken consts.Token) bool {
	return token.Type == expectToken
}

// isFinish 判断程序是否读取结束
func (p *Parser) isFinish(token util.TokenNode) bool {
	return p.match(token, consts.TokenMap["EOF"])
}

// isFuncType 判断token是否是函数类型
func (p *Parser) isFuncType(token util.TokenNode) bool {
	t := token.Type
	return t == consts.TokenMap["int"] || t == consts.TokenMap["char"] || t == consts.TokenMap["float"] || t == consts.TokenMap["void"]
}

// isConstType 判断token是否是常数类型
func (p *Parser) isConstType(token util.TokenNode) bool {
	t := token.Type
	return t == consts.TokenMap["integer"] || t == consts.TokenMap["floatnumber"] || t == consts.TokenMap["character"]
}

// isVarType 判断token是否是变量类型
func (p *Parser) isVarType(token util.TokenNode) bool {
	t := token.Type
	return t == consts.TokenMap["int"] || t == consts.TokenMap["float"] || t == consts.TokenMap["char"]
}

// isRelaOpe 判断token是否是关系运算符
func (p *Parser) isRelaOpe(token util.TokenNode) bool {
	t := token.Type
	return t == consts.TokenMap[">"] || t == consts.TokenMap["<"] || t == consts.TokenMap[">="] || t == consts.TokenMap["<="] || t == consts.TokenMap["=="] || t == consts.TokenMap["!="]
}

// isStatement 判断token是否是值声明语句
func (p *Parser) isDeclarationValue(token util.TokenNode) bool {
	t := token.Type
	return t == consts.TokenMap["const"] || t == consts.TokenMap["var"]
}

// isStatement 判断token是否是执行语句
func (p *Parser) isExeStatement(token util.TokenNode) bool {
	t := token.Type
	return t == consts.TokenMap["{"] || t == consts.TokenMap["identifier"] || t == consts.TokenMap["if"] || t == consts.TokenMap["do"] || t == consts.TokenMap["while"] || t == consts.TokenMap["for"] || t == consts.TokenMap["return"] || t == consts.TokenMap["continue"] || t == consts.TokenMap["break"]
}

// isControlStatement 判断token是否是控制语句
func (p *Parser) isControlStatement(token util.TokenNode) bool {
	t := token.Type
	return t == consts.TokenMap["if"] || t == consts.TokenMap["do"] || t == consts.TokenMap["while"] || t == consts.TokenMap["for"] || t == consts.TokenMap["return"] || t == consts.TokenMap["continue"] || t == consts.TokenMap["break"]
}

// backup 回退一个token
func (p *Parser) backup() {
	if p.Index > 0 {
		p.Index--
	}
}

// program <程序>
func (p *Parser) program() *util.TreeNode {
	nodeName := "<程序>"
	root := util.NewTreeNode(nodeName)
	var token util.TokenNode
	var node *util.TreeNode
	var flag bool
	state := 0
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["main"]) {
				state = 1
				continue
			}
			flag, node = p.declarationStatement()
			if flag {
				root.AddChild(node)
			} else {
				state = 1
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["main"]) {
				state = 2
				root.AddChild(util.NewTreeNode("main"))
			} else {
				state = 2
				p.Logger.AddParserErr(token, nodeName, "缺少main函数")
			}
		case 2:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["("]) {
				state = 3
				root.AddChild(util.NewTreeNode("("))
			} else {
				state = 3
				p.Logger.AddParserErr(token, nodeName, "缺少 ( ")
			}
		case 3:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = 4
				root.AddChild(util.NewTreeNode(")"))
			} else {
				state = 4
				p.Logger.AddParserErr(token, nodeName, "缺少 ) ")
			}
		case 4:
			flag, node = p.compoundStatement()
			if flag {
				state = 5
				root.AddChild(node)
			} else {
				state = 5
			}
		case 5:
			flag, node = p.functionBlock()
			if flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
			}
		}
	}
	return root
}

// declarationStatement <声明语句>
func (p *Parser) declarationStatement() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<声明语句>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	var node *util.TreeNode
	state := 0
	var flag bool
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["var"]) || p.match(token, consts.TokenMap["const"]) { //值声明
				state = 1
			} else if p.isFuncType(token) {
				state = 2
			} else {
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.declarationValue(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 2:
			if flag, node = p.declarationFunctionStatement(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}

	return
}

// compoundStatement <复合语句>
func (p *Parser) compoundStatement() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<复合语句>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	var node *util.TreeNode
	state := 0
	var flag bool

	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["{"]) {
				state = 1
				node = util.NewTreeNode("{")
				root.AddChild(node)
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 { ")
			}
		case 1:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["}"]) {
				state = 2
			} else if flag, node = p.statementTable(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["}"]) {
				state = -1
				node = util.NewTreeNode("}")
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 } ")
			}
		}
	}
	return
}

// statementTable <语句表>
func (p *Parser) statementTable() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<语句表>"
	root = util.NewTreeNode(nodeName)
	var node *util.TreeNode
	state := 0
	var flag bool

	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.statement(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			if flag, node = p.statementTable0(); flag { //不为空且没有错误
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// statementTable0 <语句表0>
func (p *Parser) statementTable0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<语句表0>"
	root = util.NewTreeNode(nodeName)
	var node *util.TreeNode
	state := 0
	var flag bool
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.isDeclarationValue(token) || p.isExeStatement(token) {
				state = 1
			} else { //推断为空
				state = -1
			}
		case 1:
			if flag, node = p.statementTable(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// statement <语句>
func (p *Parser) statement() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<语句>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	var node *util.TreeNode
	state := 0
	var flag bool

	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.isDeclarationValue(token) { //在复合语句中只能进行值声明，不能进行函数声明
				state = 1
			} else if p.isExeStatement(token) {
				state = 2
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName)
			}
		case 1:
			if flag, node = p.declarationValue(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 2:
			if flag, node = p.exeStatement(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// exeStatement <执行语句>
func (p *Parser) exeStatement() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<执行语句>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	var node *util.TreeNode
	state := 0
	var flag bool

	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["{"]) {
				state = 1
			} else if p.match(token, consts.TokenMap["identifier"]) {
				state = 2
			} else if p.isControlStatement(token) {
				state = 3
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName)
			}
		case 1:
			if flag, node = p.compoundStatement(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 2:
			if flag, node = p.dataHandleStatement(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 3:
			if flag, node = p.controlStatement(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// dataHandleStatement <数据处理语句>
func (p *Parser) dataHandleStatement() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<数据处理语句>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	var node *util.TreeNode
	state := 0
	var flag bool

	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["identifier"]) {
				state = 1
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少标识符")
			}
		case 1:
			token = p.peek(2)
			if p.match(token, consts.TokenMap["="]) {
				state = 2
			} else if p.match(token, consts.TokenMap["("]) {
				state = 3
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 = 或 ( ")
			}
		case 2:
			if flag, node = p.assignmentStatement(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 3:
			if flag, node = p.funcCallStatement(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// functionBlock <函数块>
func (p *Parser) functionBlock() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<函数块>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	var node *util.TreeNode
	state := 0
	var flag bool
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.isFuncType(token) {
				state = 1
			} else {
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.functionDefine(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 2:
			if flag, node = p.functionBlock(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}

	return
}

// functionDefine <函数定义>
func (p *Parser) functionDefine() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<函数定义>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	var node *util.TreeNode
	state := 0
	var flag bool
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.funcType(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			if flag, node = p.Var(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["("]) {
				state = 3
				node = util.NewTreeNode("(")
				root.AddChild(node)
			} else {
				state = 3
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 (")
			}
		case 3:
			if flag, node = p.defineFormalParamList(); flag {
				state = 4
				root.AddChild(node)
			} else {
				state = 4
				ok = false
			}
		case 4:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = 5
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = 5
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 )")
			}
		case 5:
			if flag, node = p.compoundStatement(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}

	return
}

// declarationValue <值声明>
func (p *Parser) declarationValue() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<值声明>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	var node *util.TreeNode
	state := 0
	var flag bool
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["const"]) {
				state = 1
			} else if p.match(token, consts.TokenMap["var"]) {
				state = 2
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少值声明关键字")
			}
		case 1:
			if flag, node = p.declarationConst(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 2:
			if flag, node = p.declarationVar(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}

	return
}

// declarationFunctionStatement <函数声明语句>
func (p *Parser) declarationFunctionStatement() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<函数声明语句>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.declarationFunction(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = -1
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				p.backup()
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 ; ")
			}
		}
	}
	return
}

// declarationFunction <函数声明>
func (p *Parser) declarationFunction() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<函数声明>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.funcType(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			if flag, node = p.Var(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["("]) {
				state = 3
				node = util.NewTreeNode("(")
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, " ( 缺失")
			}
		case 3:
			if flag, node = p.declFormalParamList(); flag {
				state = 4
				root.AddChild(node)
			} else {
				state = 4
				ok = false
			}
		case 4:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = -1
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, " ) 缺失")
			}
		}
	}
	return
}

// declarationConst <常量声明>
func (p *Parser) declarationConst() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<常量声明>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			token := p.nextToken()
			if p.match(token, consts.TokenMap["const"]) {
				state = 1
				node = util.NewTreeNode("const")
				root.AddChild(node)
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少关键字const")
			}
		case 1:
			if flag, node = p.constType(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			if flag, node = p.declarationConstTable(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// constType <常量类型>
func (p *Parser) constType() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<常量类型>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.isVarType(token) {
				state = -1
				node = util.NewTreeNode(token.Value)
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "类型缺失")
			}
		}
	}
	return
}

// declarationConstTable <常量声明表>
func (p *Parser) declarationConstTable() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<常量声明表>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.Var(); flag { //标识符
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["="]) {
				state = 2
				node = util.NewTreeNode("=")
				root.AddChild(node)
			} else {
				state = 2
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 = ")
			}
		case 2:
			if flag, node = p.declarationConstTable0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// declarationConstTable0 <常量声明表0>
func (p *Parser) declarationConstTable0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<常量声明表0>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.declarationConstTableValue(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			if flag, node = p.declarationConstTable1(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// declarationConstTable1 <常量声明表1>
func (p *Parser) declarationConstTable1() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<常量声明表1>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = -1
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else if p.match(token, consts.TokenMap[","]) {
				state = 1
				node = util.NewTreeNode(",")
				root.AddChild(node)
			} else {
				p.backup()
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 ; 或 ,")
			}
		case 1:
			if flag, node = p.declarationConstTable(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// declarationConstTableValue <常量声明表值>
func (p *Parser) declarationConstTableValue() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<常量声明表值>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var token util.TokenNode
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["identifier"]) {
				state = 1
			} else if p.isConstType(token) {
				state = 2
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName)
			}
		case 1:
			if flag, node = p.Var(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 2:
			if flag, node = p.Const(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// Var <变量>
func (p *Parser) Var() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<变量>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	state := 0
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["identifier"]) { //标识符
				state = -1
				node = util.NewTreeNode(token.Value)
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少标识符")
			}
		}
	}
	return
}

// Const <常量>
func (p *Parser) Const() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<常量>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var node *util.TreeNode
	var flag bool
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["character"]) {
				state = 1
			} else if p.isConstType(token) {
				state = 2
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少常量")
			}
		case 1:
			if flag, node = p.charConst(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 2:
			if flag, node = p.numberConst(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// numberConst <数值型常量>
func (p *Parser) numberConst() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<数值型常量>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	state := 0
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["integer"]) { //整型
				state = -1
				node = util.NewTreeNode(token.Value)
				root.AddChild(node)
			} else if p.match(token, consts.TokenMap["floatnumber"]) {
				state = -1
				node = util.NewTreeNode(token.Value)
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少数值型常量")
			}
		}
	}
	return
}

// charConst <字符型常量>
func (p *Parser) charConst() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<字符型常量>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	state := 0
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["character"]) { //标识符
				state = -1
				node = util.NewTreeNode(token.Value)
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少字符型常量")
			}
		}
	}
	return
}

// funcType <函数类型>
func (p *Parser) funcType() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<函数类型>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	state := 0
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.isFuncType(token) {
				state = -1
				node = util.NewTreeNode(token.Value)
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少函数类型")
			}
		}
	}
	return
}

// declarationVar <变量声明>
func (p *Parser) declarationVar() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<变量声明>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["var"]) {
				state = 1
				node = util.NewTreeNode("var")
				root.AddChild(node)
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少关键字 var ")
			}
		case 1:
			if flag, node = p.varType(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			if flag, node = p.declarationVarTable(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// varType <变量类型>
func (p *Parser) varType() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<变量类型>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.isVarType(token) {
				state = -1
				node = util.NewTreeNode(token.Value)
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少变量类型")
			}
		}
	}
	return
}

// declarationVarTable <变量声明表>
func (p *Parser) declarationVarTable() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<变量声明表>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.declarationSingleVar(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			if flag, node = p.declarationVarTable0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// declarationSingleVar <单变量声明>
func (p *Parser) declarationSingleVar() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<单变量声明>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.Var(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			if flag, node = p.declarationSingleVar0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// declarationVarTable0 <变量声明表0>
func (p *Parser) declarationVarTable0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<变量声明表0>"
	root = util.NewTreeNode(nodeName)
	var token util.TokenNode
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = -1
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else if p.match(token, consts.TokenMap[","]) {
				state = 1
				node = util.NewTreeNode(",")
				root.AddChild(node)
			} else {
				p.backup()
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 ; 或 ,")
			}
		case 1:
			if flag, node = p.declarationVarTable(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// declarationSingleVar0 <单变量声明0>
func (p *Parser) declarationSingleVar0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<单变量声明0>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["="]) {
				state = 1
				node = util.NewTreeNode("=")
				root.AddChild(node)
			} else {
				p.backup()
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.boolExp(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// arithmeticExp <算术表达式>
func (p *Parser) arithmeticExp() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<算术表达式>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.item(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			if flag, node = p.arithmeticExp0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// arithmeticExp0 <算术表达式0>
func (p *Parser) arithmeticExp0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<算术表达式0>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag, flagNull bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["+"]) {
				state = 1
				node = util.NewTreeNode("+")
				root.AddChild(node)
			} else if p.match(token, consts.TokenMap["-"]) {
				state = 1
				node = util.NewTreeNode("-")
				root.AddChild(node)
			} else {
				p.backup()
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.item(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			if flag, node = p.arithmeticExp0(); flag {
				state = -1
				if !flagNull {
					root.AddChild(node)
				}
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// item <项>
func (p *Parser) item() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<项>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag, flagNull bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.factor(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			if flag, node = p.item0(); flag {
				state = -1
				if !flagNull {
					root.AddChild(node)
				}
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// item0 <项0>
func (p *Parser) item0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<项0>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["*"]) {
				state = 1
				node = util.NewTreeNode("*")
				root.AddChild(node)
			} else if p.match(token, consts.TokenMap["/"]) {
				state = 1
				node = util.NewTreeNode("/")
				root.AddChild(node)
			} else if p.match(token, consts.TokenMap["%"]) {
				state = 1
				node = util.NewTreeNode("%")
				root.AddChild(node)
			} else {
				p.backup()
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.factor(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			if flag, node = p.item0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// factor <因子>
func (p *Parser) factor() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<因子>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["("]) {
				p.nextToken()
				state = 1
				node = util.NewTreeNode("(")
				root.AddChild(node)
			} else if p.isConstType(token) {
				state = 2
			} else if p.match(token, consts.TokenMap["identifier"]) {
				if p.match(p.peek(2), consts.TokenMap["("]) {
					state = 5
				} else {
					state = 3
				}
			} else if p.match(token, consts.TokenMap["-"]) || p.match(token, consts.TokenMap["+"]) || p.match(token, consts.TokenMap["!"]) {
				state = 6
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName)
			}
		case 1:
			if flag, node = p.arithmeticExp(); flag {
				state = 4
				root.AddChild(node)
			} else {
				state = 4
				ok = false
			}
		case 2:
			if flag, node = p.Const(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 3:
			if flag, node = p.Var(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 4:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = -1
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 ) ")
			}
		case 5:
			if flag, node = p.funcCall(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 6:
			if flag, node = p.factor0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// factor0 <因子0>
func (p *Parser) factor0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<因子0>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["+"]) || p.match(token, consts.TokenMap["-"]) || p.match(token, consts.TokenMap["!"]) {
				state = 1
				node = util.NewTreeNode(token.Value)
				root.AddChild(node)
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "因子0缺少 + 或 - 或 !")
			}
		case 1:
			if flag, node = p.factor(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// relationalExp <关系表达式>
func (p *Parser) relationalExp() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<关系表达式>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.arithmeticExp(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			if flag, node = p.relationalOpe(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			if flag, node = p.arithmeticExp(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// relationalOpe <关系运算符>
func (p *Parser) relationalOpe() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<关系运算符>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[">"]) {
				state = -1
				node = util.NewTreeNode(">")
				root.AddChild(node)
			} else if p.match(token, consts.TokenMap["<"]) {
				state = -1
				node = util.NewTreeNode("<")
				root.AddChild(node)
			} else if p.match(token, consts.TokenMap[">="]) {
				state = -1
				node = util.NewTreeNode(">=")
				root.AddChild(node)
			} else if p.match(token, consts.TokenMap["<="]) {
				state = -1
				node = util.NewTreeNode("<=")
				root.AddChild(node)
			} else if p.match(token, consts.TokenMap["=="]) {
				state = -1
				node = util.NewTreeNode("==")
				root.AddChild(node)
			} else if p.match(token, consts.TokenMap["!="]) {
				state = -1
				node = util.NewTreeNode("!=")
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少关系运算符")
			}
		}
	}
	return
}

// boolExp <布尔表达式>
func (p *Parser) boolExp() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<布尔表达式>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.boolItem(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			if flag, node = p.boolExp0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		}
	}
	return
}

// boolExp0 <布尔表达式0>
func (p *Parser) boolExp0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<布尔表达式0>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["||"]) {
				state = 1
				node = util.NewTreeNode("||")
				root.AddChild(node)
			} else {
				state = -1
				p.backup()
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.boolItem(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			if flag, node = p.boolExp0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// boolItem <布尔项>
func (p *Parser) boolItem() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<布尔项>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.boolFactor(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			if flag, node = p.boolItem0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// boolItem0 <布尔项0>
func (p *Parser) boolItem0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<布尔项0>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["&&"]) {
				state = 1
				node = util.NewTreeNode("&&")
				root.AddChild(node)
			} else {
				state = -1
				p.backup()
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.boolFactor(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			if flag, node = p.boolItem0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// boolFactor <布尔因子>
func (p *Parser) boolFactor() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<布尔因子>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.arithmeticExp(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			if flag, node = p.boolFactor0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// boolFactor0 <布尔因子0>
func (p *Parser) boolFactor0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<布尔因子0>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.isRelaOpe(token) {
				state = 1
				root.AddChild(node)
			} else {
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.relationalOpe(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 2:
			if flag, node = p.arithmeticExp(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// assignmentStatement <赋值语句>
func (p *Parser) assignmentStatement() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<赋值语句>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.assignmentExp(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = -1
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				p.backup()
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 ; ")
			}
		}
	}
	return
}

// assignmentExp <赋值表达式>
func (p *Parser) assignmentExp() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<赋值表达式>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["identifier"]) {
				state = 1
				node = util.NewTreeNode(token.Value)
				root.AddChild(node)
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少标识符")
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["="]) {
				state = 2
				node = util.NewTreeNode("=")
				root.AddChild(node)
			} else {
				state = 2
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 = ")
			}
		case 2:
			if flag, node = p.assignmentExp0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// assignmentExp0 <赋值表达式0>
func (p *Parser) assignmentExp0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<赋值表达式0>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["identifier"]) && p.match(p.peek(2), consts.TokenMap["("]) {
				state = 2
			} else {
				state = 1
			}
		case 1:
			if flag, node = p.boolExp(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 2:
			if flag, node = p.funcCall(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// funcCallStatement <函数调用语句>
func (p *Parser) funcCallStatement() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<函数调用语句>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["identifier"]) {
				state = 1
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少函数变量名")
			}
		case 1:
			if flag, node = p.funcCall(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = -1
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				p.backup()
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "函数调用语句缺少 ; ")
			}
		}
	}
	return
}

// funcCall <函数调用>
func (p *Parser) funcCall() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<函数调用>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.Var(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["("]) {
				state = 2
				node = util.NewTreeNode("(")
				root.AddChild(node)
			} else {
				state = 2
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 ( ")
			}
		case 2:
			if flag, node = p.actualParamList(); flag {
				state = 3
				root.AddChild(node)
			} else {
				state = 3
				ok = false
			}
		case 3:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = -1
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 ) ")
			}
		}
	}
	return
}

// actualParamList <实参列表>
func (p *Parser) actualParamList() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<实参列表>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.isConstType(token) || p.match(token, consts.TokenMap["identifier"]) || p.match(token, consts.TokenMap["("]) {
				state = 1
			} else {
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.actualParam(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// actualParam <实参>
func (p *Parser) actualParam() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<实参>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.isConstType(token) || p.match(token, consts.TokenMap["identifier"]) || p.match(token, consts.TokenMap["("]) {
				state = 1
			} else {
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.boolExp(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 2:
			if flag, node = p.actualParam0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// actualParam0 <实参0>
func (p *Parser) actualParam0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<实参0>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap[","]) {
				token = p.nextToken()
				state = 1
				node = util.NewTreeNode(",")
				root.AddChild(node)
			} else {
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.actualParam(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// declFormalParamList <函数声明形参列表>
func (p *Parser) declFormalParamList() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<函数声明形参列表>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.isVarType(token) {
				state = 1
			} else {
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.declFormalParam(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// declFormalParam <函数声明形参>
func (p *Parser) declFormalParam() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<函数声明形参>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.varType(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = 1
				ok = false
			}
		case 1:
			if flag, node = p.declFormalParam0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// declFormalParam0 <函数声明形参0>
func (p *Parser) declFormalParam0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<函数声明形参0>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap[","]) {
				p.nextToken()
				state = 1
				node = util.NewTreeNode(",")
				root.AddChild(node)
			} else {
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.declFormalParam(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// defineFormalParamList <函数定义形参列表>
func (p *Parser) defineFormalParamList() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<函数定义形参列表>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.isVarType(token) {
				state = 1
			} else {
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.defineFormalParam(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// defineFormalParam <函数定义形参>
func (p *Parser) defineFormalParam() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<函数定义形参>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.varType(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 1:
			if flag, node = p.Var(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			if flag, node = p.defineFormalParam0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// defineFormalParam0 <函数定义形参0>
func (p *Parser) defineFormalParam0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<函数定义形参0>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap[","]) {
				p.nextToken()
				state = 1
				node = util.NewTreeNode(",")
				root.AddChild(node)
			} else {
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.defineFormalParam(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// controlStatement <控制语句>
func (p *Parser) controlStatement() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<控制语句>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["if"]) {
				state = 1
			} else if p.match(token, consts.TokenMap["for"]) {
				state = 2
			} else if p.match(token, consts.TokenMap["while"]) {
				state = 3
			} else if p.match(token, consts.TokenMap["do"]) {
				state = 4
			} else if p.match(token, consts.TokenMap["return"]) {
				state = 5
			} else if p.match(token, consts.TokenMap["break"]) {
				state = 6
			} else if p.match(token, consts.TokenMap["continue"]) {
				state = 7
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少控制语句关键字")
			}
		case 1:
			if flag, node = p.IF(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 2:
			if flag, node = p.FOR(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 3:
			if flag, node = p.WHILE(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 4:
			if flag, node = p.DoWHILE(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 5:
			if flag, node = p.Return(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 6:
			if flag, node = p.Break(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 7:
			if flag, node = p.Continue(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// IF <if语句>
func (p *Parser) IF() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<if语句>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["if"]) {
				state = 1
				node = util.NewTreeNode("if")
				root.AddChild(node)
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少if")
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["("]) {
				state = 2
				node = util.NewTreeNode("(")
				root.AddChild(node)
			} else {
				state = 2
				ok = false
				p.Logger.AddParserErr(token, nodeName, " if 缺少 ( ")
			}
		case 2:
			if flag, node = p.boolExp(); flag {
				state = 3
				root.AddChild(node)
			} else {
				state = 3
				ok = false
			}
		case 3:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = 4
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = 4
				ok = false
				p.Logger.AddParserErr(token, nodeName, "if 缺少 ) ")
			}
		case 4:
			if flag, node = p.compoundStatement(); flag {
				state = 5
				root.AddChild(node)
			} else {
				state = 5
				ok = false
			}
		case 5:
			if flag, node = p.IfTail(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// IfTail <IfTail语句>
func (p *Parser) IfTail() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<IfTail语句>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["else"]) {
				state = 1
			} else {
				state = -1
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["else"]) {
				state = 2
				node = util.NewTreeNode("else")
				root.AddChild(node)
			} else {
				state = 2
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少else")
			}
		case 2:
			if flag, node = p.IfTail0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// IfTail0 <IfTail0语句>
func (p *Parser) IfTail0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<IfTail0语句>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["{"]) {
				state = 1
			} else if p.match(token, consts.TokenMap["if"]) {
				state = 2
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "else 缺少 { 或 if")
			}
		case 1:
			if flag, node = p.compoundStatement(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 2:
			if flag, node = p.IF(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// FOR <for语句>
func (p *Parser) FOR() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<for语句>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["for"]) {
				state = 1
				node = util.NewTreeNode("for")
				root.AddChild(node)
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少for")
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["("]) {
				state = 2
				node = util.NewTreeNode("(")
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, " for 缺少 ( ")
			}
		case 2:
			if flag, node = p.assignmentExp(); flag {
				state = 3
				root.AddChild(node)
			} else {
				state = 3
				ok = false
			}
		case 3:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = 4
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				p.backup()
				state = 4
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 ; ")
			}
		case 4:
			if flag, node = p.boolExp(); flag {
				state = 5
				root.AddChild(node)
			} else {
				state = 5
				ok = false
			}
		case 5:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = 6
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				p.backup()
				state = 6
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 ; ")
			}
		case 6:
			if flag, node = p.assignmentExp(); flag {
				state = 7
				root.AddChild(node)
			} else {
				state = 7
				ok = false
			}
		case 7:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = 8
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = 8
				ok = false
				p.Logger.AddParserErr(token, nodeName, "for 缺少 ) ")
			}
		case 8:
			if flag, node = p.compoundStatement(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// WHILE <while语句>
func (p *Parser) WHILE() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<while语句>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["while"]) {
				state = 1
				node = util.NewTreeNode("while")
				root.AddChild(node)
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少while")
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["("]) {
				state = 2
				node = util.NewTreeNode("(")
				root.AddChild(node)
			} else {
				state = 2
				ok = false
				p.Logger.AddParserErr(token, nodeName, " while 缺少 ( ")
			}
		case 2:
			if flag, node = p.boolExp(); flag {
				state = 3
				root.AddChild(node)
			} else {
				state = 3
				ok = false
			}
		case 3:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = 4
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = 4
				ok = false
				p.Logger.AddParserErr(token, nodeName, "while 缺少 ) ")
			}
		case 4:
			if flag, node = p.compoundStatement(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// DoWHILE <DoWHILE语句>
func (p *Parser) DoWHILE() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<DoWHILE语句>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["do"]) {
				state = 1
				node = util.NewTreeNode("do")
				root.AddChild(node)
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少do")
			}
		case 1:
			if flag, node = p.compoundStatement(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["while"]) {
				state = 3
				node = util.NewTreeNode("while")
				root.AddChild(node)
			} else {
				state = 3
				ok = false
				p.Logger.AddParserErr(token, nodeName, " do 缺少 while")
			}
		case 3:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["("]) {
				state = 4
				node = util.NewTreeNode("(")
				root.AddChild(node)
			} else {
				state = 4
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 ( ")
			}
		case 4:
			if flag, node = p.boolExp(); flag {
				state = 5
				root.AddChild(node)
			} else {
				state = 5
				ok = false
			}
		case 5:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = 6
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = 6
				ok = false
				p.Logger.AddParserErr(token, nodeName, "while 缺少 ) ")
			}
		case 6:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = -1
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				p.backup()
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "do while 缺少 ; ")
			}
		}
	}
	return
}

// Return <return语句>
func (p *Parser) Return() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<Return语句>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["return"]) {
				state = 1
				node = util.NewTreeNode("return")
				root.AddChild(node)
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少return")
			}
		case 1:
			if flag, node = p.Return0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		}
	}
	return
}

// Return0 <return语句0>
func (p *Parser) Return0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<return语句0>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap[";"]) {
				state = 2
			} else {
				state = 1
			}
		case 1:
			if flag, node = p.boolExp(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = 2
				ok = false
			}
		case 2:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = -1
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				p.backup()
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "return 缺少 ; ")
			}
		}
	}
	return
}

// Break <break语句>
func (p *Parser) Break() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<break语句>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["break"]) {
				state = 1
				node = util.NewTreeNode("break")
				root.AddChild(node)
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少break")
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = -1
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				p.backup()
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "break 缺少 ; ")
			}
		}
	}
	return
}

// Continue <continue语句>
func (p *Parser) Continue() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<continue语句>"
	root = util.NewTreeNode(nodeName)
	state := 0
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["continue"]) {
				state = 1
				node = util.NewTreeNode("continue")
				root.AddChild(node)
			} else {
				state = 1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少continue")
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = -1
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				p.backup()
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "continue 缺少 ; ")
			}
		}
	}
	return
}
