package util

import (
	"complier/pkg/consts"
	"errors"
)

type Stack[T any] struct {
	data []T
}

func NewStack() *Stack[any] {
	return &Stack[any]{}
}

// Size 查看栈大小
func (stack *Stack[any]) Size() int {
	return len(stack.data)
}

// IsEmpty 查看栈是否为空
func (stack *Stack[any]) IsEmpty() bool {
	return len(stack.data) == 0
}

// Cap 查看栈容量
func (stack *Stack[any]) Cap() int {
	return cap(stack.data)
}

// Push 入栈
func (stack *Stack[any]) Push(value any) {
	stack.data = append(stack.data, value)
}

// Top 查看栈顶元素
func (stack *Stack[any]) Top() interface{} {
	if len(stack.data) == 0 {
		return nil
	}
	return stack.data[len(stack.data)-1]
}

// Pop 出栈
func (stack *Stack[any]) Pop() (interface{}, error) {
	theStack := stack.data
	if len(theStack) == 0 {
		return nil, errors.New("栈为空")
	}
	value := theStack[len(theStack)-1]
	stack.data = theStack[:len(theStack)-1]
	return value, nil
}

// CalStack 计算栈
type CalStack struct {
	NumStack *Stack[any]
	OpStack  *Stack[any]
	qf       *QuaFormList
}

// NewCalStack 创建计算栈
func NewCalStack(qf *QuaFormList) *CalStack {
	return &CalStack{
		NumStack: NewStack(),
		OpStack:  NewStack(),
		qf:       qf,
	}
}

// priority 优先级判断，数字越小优先级越高,当栈顶运算符优先级大于等于当前运算符时，栈顶运算符出栈并进行一次运算
func (c *CalStack) priority(op any) int {
	switch op {
	case consts.QUA_LEFTSMALLBRACKET, consts.QUA_RIGHTSMALLBRACKET:
		return 1
	case consts.QUA_NOT, consts.QUA_NEGATIVE:
		return 2
	case consts.QUA_MUL, consts.QUA_DIV, consts.QUA_MOD:
		return 3
	case consts.QUA_ADD, consts.QUA_SUB:
		return 4
	case consts.QUA_GT, consts.QUA_GE, consts.QUA_LT, consts.QUA_LE:
		return 5
	case consts.QUA_EQ, consts.QUA_NE:
		return 6
	case consts.QUA_AND:
		return 7
	case consts.QUA_OR:
		return 8
	case consts.QUA_ASSIGNMENT:
		return 9
	default:
		return 10
	}

}

// PushNum 入数字栈
func (c *CalStack) PushNum(value any) {
	c.NumStack.Push(value)
}

// PushOp 入操作符栈
func (c *CalStack) PushOp(ope int) {
	top := c.OpStack.Top()
	if c.priority(top) <= c.priority(ope) {
		c.Cal()
	}
	c.OpStack.Push(ope)
}

// Cal 对数字栈顶两个元素进行一次计算,遇到"@","!"只取栈顶一个元素进行一次计算
func (c *CalStack) Cal() {
	op, _ := c.OpStack.Pop()
	num1, _ := c.NumStack.Pop()
	if consts.QUA_NOT == op || consts.QUA_NEGATIVE == op {
		result := c.qf.GetTemp()
		c.qf.AddQuaForm(consts.QuaFormMap[op.(int)], num1, nil, result)
		c.NumStack.Push(result)
		return
	}

	num2, _ := c.NumStack.Pop()
	// 遇到赋值运算符，num2为变量，num1为值
	if op == consts.QUA_ASSIGNMENT {
		c.qf.AddQuaForm(consts.QuaFormMap[op.(int)], num1, nil, num2)
		return
	}

	result := c.qf.GetTemp()
	c.qf.AddQuaForm(consts.QuaFormMap[op.(int)], num1, num2, result)
	c.NumStack.Push(result)
}

// CalAll 计算所有
func (c *CalStack) CalAll() {
	for !c.OpStack.IsEmpty() {
		c.Cal()
	}
}

// CalStacks 计算栈集合
type CalStacks struct {
	CurrentStack *CalStack   //当前计算栈
	BracketStack *Stack[any] //括号栈，每遇到一个左括号，就新建一个计算栈，遇到右括号，就将计算栈出栈
	qf           *QuaFormList
}

// NewCalStacks 创建计算栈集合
func NewCalStacks(qf *QuaFormList) *CalStacks {
	current := NewCalStack(qf)
	c := &CalStacks{
		CurrentStack: current,
		BracketStack: NewStack(),
		qf:           qf,
	}
	c.BracketStack.Push(current)
	return c
}

func (c *CalStacks) PushNum(value any) {
	c.CurrentStack.PushNum(value)
}

func (c *CalStacks) PushOpe(ope int) {
	if ope == consts.QUA_LEFTSMALLBRACKET {
		current := NewCalStack(c.qf)
		c.BracketStack.Push(current)
		c.CurrentStack = current
		return
	}
	if ope == consts.QUA_RIGHTSMALLBRACKET {
		c.CurrentStack.CalAll()
		tempResult := c.CurrentStack.NumStack.Top()
		c.BracketStack.Pop()
		c.CurrentStack = c.BracketStack.Top().(*CalStack)
		if c.CurrentStack != nil { //栈不为空，说明当前只是计算完了一个括号内的表达式，要生成一个临时变量放入当前数字栈
			c.CurrentStack.PushNum(tempResult)
		}
		return
	}
	c.CurrentStack.PushOp(ope)
}

func (c *CalStacks) CalAll() {
	c.CurrentStack.CalAll()
}
