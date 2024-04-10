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
	return t == consts.TokenMap["{"] || t == consts.TokenMap["identifier"] || t == consts.TokenMap["if"] || t == consts.TokenMap["do"] || t == consts.TokenMap["while"] || t == consts.TokenMap["for"] || t == consts.TokenMap["return"]
}

// isControlStatement 判断token是否是控制语句
func (p *Parser) isControlStatement(token util.TokenNode) bool {
	t := token.Type
	return t == consts.TokenMap["if"] || t == consts.TokenMap["do"] || t == consts.TokenMap["while"] || t == consts.TokenMap["for"] || t == consts.TokenMap["return"]
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
	state := 0
	for state != -1 {
		if p.isFinish(token) {
			if state == 0 {
				p.Logger.AddErr("缺少main函数")
			}
			state = -1
			break
		}
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["main"]) {
				state = 1
				continue
			}
			statement, node := p.declarationStatement()
			if statement {
				root.AddChild(node)
			} else {
				state = 1
				p.Logger.AddParserErr(token, nodeName)
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
				p.Logger.AddParserErr(token, nodeName, "左括号缺失")
			}
		case 3:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = 4
				root.AddChild(util.NewTreeNode(")"))
			} else {
				state = 4
				p.Logger.AddParserErr(token, nodeName, "右括号缺失")
			}
		case 4:
			statement, node := p.compoundStatement()
			if statement {
				state = 5
				root.AddChild(node)
			} else {
				state = 5
				p.Logger.AddParserErr(token, nodeName)
			}
		case 5:
			block, node := p.functionBlock()
			if block {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
				p.Logger.AddParserErr(token, nodeName)
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
		if p.isFinish(token) {
			state = -1
			break
		}
		switch state {
		case 0:
			token = p.peek(1)
			if p.match(token, consts.TokenMap["var"]) || p.match(token, consts.TokenMap["const"]) { //值声明
				state = 1
			} else if p.isFuncType(token) {
				state = 2
			} else {
				state = -1
				ok = false
				root.AddChild(util.NewTreeNode("ε"))
			}
		case 1:
			if flag, node = p.declarationValue(); flag {
				state = -1
				root.AddChild(node)
			}
		case 2:
			if flag, node = p.declarationFunctionStatement(); flag {
				state = -1
				root.AddChild(node)
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
				state = -1
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
				state = -1
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
	var flagNull bool

	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.statement(); flag {
				state = 1
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 1:
			if flagNull, flag, node = p.statementTable0(); flag { //不为空且没有错误
				state = -1
				if !flagNull {
					root.AddChild(node)
				}

			} else {
				state = -1
			}
		}
	}
	return
}

// statementTable0 <语句表0>
func (p *Parser) statementTable0() (null bool, ok bool, root *util.TreeNode) {
	null = false
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
				null = true
			}
			//TODO: 还不确定这里能不能推断为空，如果不能要进行报错
		case 1:
			if flag, node = p.statementTable(); flag {
				state = -1
				root.AddChild(node)
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
			}
		case 2:
			if flag, node = p.exeStatement(); flag {
				state = -1
				root.AddChild(node)
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
				p.Logger.AddParserErr(token, nodeName)
			}
		case 2:
			if flag, node = p.assignmentStatement(); flag {
				state = -1
				root.AddChild(node)
			}
		case 3:
			if flag, node = p.funcCallStatement(); flag {
				state = -1
				root.AddChild(node)
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
				state = -1
				ok = false
			}
		case 1:
			if flag, node = p.Var(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = -1
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
				p.Logger.AddParserErr(token, nodeName, "缺少 (")
			}
		case 3:
			if flag, node = p.formalParamList(); flag {
				state = 4
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 4:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = 5
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = -1
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
			token = p.nextToken()
			if p.match(token, consts.TokenMap["const"]) {
				state = 1
				p.backup()
			} else if p.match(token, consts.TokenMap["var"]) {
				state = 2
				p.backup()
			} else {
				state = -1
				ok = false
			}
		case 1:
			if flag, node = p.declarationConst(); flag {
				state = -1
				root.AddChild(node)
			}
		case 2:
			if flag, node = p.declarationVar(); flag {
				state = -1
				root.AddChild(node)
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
				state = -1
				ok = false
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = -1
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "函数声明语句缺少 ; ")
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
			token = p.nextToken()
			if p.isFuncType(token) {
				state = 1
				node = util.NewTreeNode(token.Value)
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "函数类型错误")
			}
		case 1:
			if flag, node = p.Var(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = -1
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
			if flag, node = p.formalParamList(); flag {
				state = 4
				root.AddChild(node)
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
func (p *Parser) declarationConst() (bool, *util.TreeNode) {
	nodeName := "<常量声明>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 1:
			if flag, node = p.constType(); flag {
				state = 2
				root.AddChild(node)
			}
		case 2:
			if flag, node = p.declarationConstTable(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// constType <常量类型>
func (p *Parser) constType() (bool, *util.TreeNode) {
	nodeName := "<常量类型>"
	root := util.NewTreeNode(nodeName)
	state := 0
	var node *util.TreeNode
	var token util.TokenNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["int"]) {
				state = -1
				node = util.NewTreeNode("int")
				root.AddChild(node)
			} else if p.match(token, consts.TokenMap["char"]) {
				state = -1
				node = util.NewTreeNode("char")
				root.AddChild(node)
			} else if p.match(token, consts.TokenMap["float"]) {
				state = -1
				node = util.NewTreeNode("float")
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// declarationConstTable <常量声明表>
func (p *Parser) declarationConstTable() (bool, *util.TreeNode) {
	nodeName := "<常量声明表>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["="]) {
				state = 2
				node = util.NewTreeNode("=")
				root.AddChild(node)
			}
		case 2:
			if flag, node = p.declarationConstTable0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// declarationConstTable0 <常量声明表0>
func (p *Parser) declarationConstTable0() (bool, *util.TreeNode) {
	nodeName := "<常量声明表0>"
	root := util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.declarationConstTableValue(); flag {
				state = 1
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.declarationConstTable1(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// declarationConstTable1 <常量声明表1>
func (p *Parser) declarationConstTable1() (bool, *util.TreeNode) {
	nodeName := "<常量声明表1>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 1:
			if flag, node = p.declarationConstTable(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// declarationConstTableValue <常量声明表值>
func (p *Parser) declarationConstTableValue() (bool, *util.TreeNode) {
	nodeName := "<常量声明表值>"
	root := util.NewTreeNode(nodeName)
	state := 0
	var token util.TokenNode
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["identifier"]) {
				state = 1
				p.backup()
			} else if p.isConstType(token) {
				state = 2
				p.backup()
			}
		case 1:
			if flag, node = p.Var(); flag {
				state = -1
				root.AddChild(node)
			}
		case 2:
			if flag, node = p.Const(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// Var <变量>
func (p *Parser) Var() (bool, *util.TreeNode) {
	nodeName := "<变量>"
	root := util.NewTreeNode(nodeName)
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
			}
		}
	}
	return true, root
}

// Const <常量>
func (p *Parser) Const() (bool, *util.TreeNode) {
	nodeName := "<常量>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 1:
			if flag, node = p.charConst(); flag {
				state = -1
				root.AddChild(node)
			}
		case 2:
			if flag, node = p.numberConst(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// numberConst <数值型常量>
func (p *Parser) numberConst() (bool, *util.TreeNode) {
	nodeName := "<数值型常量>"
	root := util.NewTreeNode(nodeName)
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
			}
		}
	}
	return true, root
}

// charConst <字符型常量>
func (p *Parser) charConst() (bool, *util.TreeNode) {
	nodeName := "<字符型常量>"
	root := util.NewTreeNode(nodeName)
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
			}
		}
	}
	return true, root
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
				p.Logger.AddParserErr(token, nodeName)
			}
		}
	}
	return
}

// declarationVar <变量声明>
func (p *Parser) declarationVar() (bool, *util.TreeNode) {
	nodeName := "<变量声明>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 1:
			if flag, node = p.varType(); flag {
				state = 2
				root.AddChild(node)
			}
		case 2:
			if flag, node = p.declarationVarTable(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
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
				p.Logger.AddParserErr(token, nodeName)
			}
		}
	}
	return
}

// declarationVarTable <变量声明表>
func (p *Parser) declarationVarTable() (bool, *util.TreeNode) {
	nodeName := "<变量声明表>"
	root := util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.declarationSingleVar(); flag {
				state = 1
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.declarationVarTable0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// declarationSingleVar <单变量声明>
func (p *Parser) declarationSingleVar() (bool, *util.TreeNode) {
	nodeName := "<单变量声明>"
	root := util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.Var(); flag {
				state = 1
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.declarationSingleVar0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// declarationVarTable0 <变量声明表0>
func (p *Parser) declarationVarTable0() (bool, *util.TreeNode) {
	nodeName := "<变量声明表0>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 1:
			if flag, node = p.declarationVarTable(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// declarationSingleVar0 <单变量声明0>
func (p *Parser) declarationSingleVar0() (bool, *util.TreeNode) {
	nodeName := "<单变量声明0>"
	root := util.NewTreeNode(nodeName)
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
				state = -1
				p.backup()
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.boolExp(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// arithmeticExp <算术表达式>
func (p *Parser) arithmeticExp() (bool, *util.TreeNode) {
	nodeName := "<算术表达式>"
	root := util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.item(); flag {
				state = 1
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.arithmeticExp0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// arithmeticExp0 <算术表达式0>
func (p *Parser) arithmeticExp0() (bool, *util.TreeNode) {
	nodeName := "<算术表达式0>"
	root := util.NewTreeNode(nodeName)
	state := 0
	var flag bool
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
				state = -1
				p.backup()
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.item(); flag {
				state = 2
				root.AddChild(node)
			}
		case 2:
			if flag, node = p.arithmeticExp0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// item <项>
func (p *Parser) item() (bool, *util.TreeNode) {
	nodeName := "<项>"
	root := util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.factor(); flag {
				state = 1
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.item0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// item0 <项0>
func (p *Parser) item0() (bool, *util.TreeNode) {
	nodeName := "<项0>"
	root := util.NewTreeNode(nodeName)
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
				state = -1
				p.backup()
				node = util.NewTreeNode("ε")
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.factor(); flag {
				state = 2
				root.AddChild(node)
			}
		case 2:
			if flag, node = p.item0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// factor <因子>
func (p *Parser) factor() (bool, *util.TreeNode) {
	nodeName := "<因子>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 1:
			if flag, node = p.arithmeticExp(); flag {
				state = 4
				root.AddChild(node)
			}
		case 2:
			if flag, node = p.Const(); flag {
				state = -1
				root.AddChild(node)
			}
		case 3:
			if flag, node = p.Var(); flag {
				state = -1
				root.AddChild(node)
			}
		case 4:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = -1
				node = util.NewTreeNode(")")
				root.AddChild(node)
			}
		case 5:
			if flag, node = p.funcCall(); flag {
				state = -1
				root.AddChild(node)
			}
		case 6:
			if flag, node = p.factor0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// factor0 <因子0>
func (p *Parser) factor0() (bool, *util.TreeNode) {
	nodeName := "<因子0>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 1:
			if flag, node = p.factor(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// relationalExp <关系表达式>
func (p *Parser) relationalExp() (bool, *util.TreeNode) {
	nodeName := "<关系表达式>"
	root := util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.arithmeticExp(); flag {
				state = 1
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.relationalOpe(); flag {
				state = 2
				root.AddChild(node)
			}
		case 2:
			if flag, node = p.arithmeticExp(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
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
			}
		}
	}
	return
}

// boolExp <布尔表达式>
func (p *Parser) boolExp() (bool, *util.TreeNode) {
	nodeName := "<布尔表达式>"
	root := util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.boolItem(); flag {
				state = 1
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.boolExp0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// boolExp0 <布尔表达式0>
func (p *Parser) boolExp0() (bool, *util.TreeNode) {
	nodeName := "<布尔表达式0>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 2:
			if flag, node = p.boolExp0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// boolItem <布尔项>
func (p *Parser) boolItem() (bool, *util.TreeNode) {
	nodeName := "<布尔项>"
	root := util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.boolFactor(); flag {
				state = 1
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.boolItem0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// boolItem0 <布尔项0>
func (p *Parser) boolItem0() (bool, *util.TreeNode) {
	nodeName := "<布尔项0>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 2:
			if flag, node = p.boolItem0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// boolFactor <布尔因子>
func (p *Parser) boolFactor() (bool, *util.TreeNode) {
	nodeName := "<布尔因子>"
	root := util.NewTreeNode(nodeName)
	state := 0
	var flag bool
	var node *util.TreeNode
	for state != -1 {
		switch state {
		case 0:
			if flag, node = p.arithmeticExp(); flag {
				state = 1
				root.AddChild(node)
			}
		case 1:
			if flag, node = p.boolFactor0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// boolFactor0 <布尔因子0>
func (p *Parser) boolFactor0() (bool, *util.TreeNode) {
	nodeName := "<布尔因子0>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 2:
			if flag, node = p.arithmeticExp(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// assignmentStatement <赋值语句>
func (p *Parser) assignmentStatement() (bool, *util.TreeNode) {
	nodeName := "<赋值语句>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = -1
				node = util.NewTreeNode(";")
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// assignmentExp <赋值表达式>
func (p *Parser) assignmentExp() (bool, *util.TreeNode) {
	nodeName := "<赋值表达式>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["="]) {
				state = 2
				node = util.NewTreeNode("=")
				root.AddChild(node)
			}
		case 2:
			if flag, node = p.assignmentExp0(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
}

// assignmentExp0 <赋值表达式0>
func (p *Parser) assignmentExp0() (bool, *util.TreeNode) {
	nodeName := "<赋值表达式0>"
	root := util.NewTreeNode(nodeName)
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
			}
		case 2:
			if flag, node = p.funcCall(); flag {
				state = -1
				root.AddChild(node)
			}
		}
	}
	return true, root
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
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "函数变量名推断错误")
			}
		case 1:
			if flag, node = p.funcCall(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = -1
			}
		case 2:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = -1
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				state = -1
				p.Logger.AddParserErr(token, nodeName, "函数调用语句缺失 ; ")
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
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "函数变量名推断错误")
			}
		case 1:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["("]) {
				state = 2
				node = util.NewTreeNode("(")
				root.AddChild(node)
			} else {
				state = -1
				p.Logger.AddParserErr(token, nodeName, " ( 缺失")
			}
		case 2:
			if flag, node = p.actualParamList(); flag {
				state = 3
				root.AddChild(node)
			} else {
				state = -1
			}
		case 3:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = -1
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = -1
				p.Logger.AddParserErr(token, nodeName, " ) 缺失")
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
			}
		case 2:
			if flag, node = p.actualParam0(); flag {
				state = -1
				root.AddChild(node)
			} else {
				state = -1
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
			}
		}
	}
	return
}

// formalParamList <形参列表>
func (p *Parser) formalParamList() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<形参列表>"
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
			if flag, node = p.formalParam(); flag {
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

// formalParam <形参>
func (p *Parser) formalParam() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<形参>"
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
				state = -1
				ok = false
			}
		case 2:
			if flag, node = p.formalParam0(); flag {
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

// formalParam0 <形参0>
func (p *Parser) formalParam0() (ok bool, root *util.TreeNode) {
	ok = true
	nodeName := "<形参0>"
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
			if flag, node = p.formalParam(); flag {
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
			} else {
				state = -1
				ok = false
				node = util.NewTreeNode("ε")
				root.AddChild(node)
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
				state = -1
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
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, " if 缺少 ( ")
			}
		case 2:
			if flag, node = p.boolExp(); flag {
				state = 3
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 3:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = 4
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, " 缺少 ) ")
			}
		case 4:
			if flag, node = p.compoundStatement(); flag {
				state = 5
				root.AddChild(node)
			} else {
				state = -1
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
				state = -1
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
				state = -1
				ok = false
			}
		case 3:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = 4
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 ; ")
			}
		case 4:
			if flag, node = p.boolExp(); flag {
				state = 5
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 5:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = 6
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 ; ")
			}
		case 6:
			if flag, node = p.assignmentExp(); flag {
				state = 7
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 7:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = 8
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = -1
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
				state = -1
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
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, " while 缺少 ( ")
			}
		case 2:
			if flag, node = p.boolExp(); flag {
				state = 3
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 3:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = 4
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = -1
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
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少do")
			}
		case 1:
			if flag, node = p.compoundStatement(); flag {
				state = 2
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 2:
			token = p.nextToken()
			if p.match(token, consts.TokenMap["while"]) {
				state = 3
				node = util.NewTreeNode("while")
				root.AddChild(node)
			} else {
				state = -1
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
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "缺少 ( ")
			}
		case 4:
			if flag, node = p.boolExp(); flag {
				state = 5
				root.AddChild(node)
			} else {
				state = -1
				ok = false
			}
		case 5:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[")"]) {
				state = 6
				node = util.NewTreeNode(")")
				root.AddChild(node)
			} else {
				state = -1
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
				state = -1
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
				state = -1
				ok = false
			}
		case 2:
			token = p.nextToken()
			if p.match(token, consts.TokenMap[";"]) {
				state = -1
				node = util.NewTreeNode(";")
				root.AddChild(node)
			} else {
				state = -1
				ok = false
				p.Logger.AddParserErr(token, nodeName, "return 缺少 ; ")
			}
		}
	}
	return
}
