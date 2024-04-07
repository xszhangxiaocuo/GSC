package compiler

import (
	"complier/pkg/consts"
	"complier/pkg/logger"
	"complier/util"
	"fmt"
)

type Parser struct {
	Token  []TokenNode
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
func (p *Parser) nextToken() (token TokenNode) {
	if p.Index < len(p.Token) {
		token = p.Token[p.Index]
		p.Index++
		return
	}
	return TokenNode{Type: consts.TokenMap["EOF"]}
}

// peek 查看下一个token
func (p *Parser) peek() TokenNode {
	if p.Index < len(p.Token) {
		return p.Token[p.Index]
	}
	return TokenNode{Type: consts.TokenMap["EOF"]}
}

// match 判断传入的token种别码与下一个token种别码是否匹配
func (p *Parser) match(expectToken consts.Token) bool {
	return p.nextToken().Type == expectToken
}

// program <程序>
func (p *Parser) program() *util.TreeNode {
	root := util.NewTreeNode("<程序>")
	token := p.peek()
	state := 0
	for state != -1 {
		switch state {
		case 0:
			if p.declarationStatement() {
				state = 1
				root.AddChild(util.NewTreeNode("<声明语句>"))
			} else {
				state = 2
				p.Logger.AddErr(fmt.Sprintf("%d:%d\t\t%d\t\t%s\t\t<程序>-><声明语句> error\n", token.Pos.Line, token.Pos.Column, token.Type, token.Value))
			}
		case 1:
			if p.match(consts.TokenMap["main"]) {
				state = 2
				root.AddChild(util.NewTreeNode("main"))
			} else {
				break
			}
		case 2:
			if p.match(consts.TokenMap["("]) {
				state = 3
				root.AddChild(util.NewTreeNode("("))
			} else {
				break
			}
		case 3:
			if p.match(consts.TokenMap[")"]) {
				state = 4
				root.AddChild(util.NewTreeNode(")"))
			} else {
				break
			}
		case 4:
			if p.compoundStatement() {
				state = 5
				root.AddChild(util.NewTreeNode("<复合语句>"))
			} else {
				break
			}
		case 5:
			if p.functionBlock() {
				state = -1
				root.AddChild(util.NewTreeNode("函数块"))
			} else {
				break
			}
		}
	}
	return root
}

// declarationStatement <声明语句>
func (p *Parser) declarationStatement() bool {

	return true
}

// compoundStatement <复合语句>
func (p *Parser) compoundStatement() bool {

	return true
}

// functionBlock <函数块>
func (p *Parser) functionBlock() bool {

	return true
}
