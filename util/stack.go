package util

import (
	"complier/pkg/consts"
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
func (stack *Stack[any]) Pop() interface{} {
	theStack := stack.data
	if len(theStack) == 0 {
		return nil
	}
	value := theStack[len(theStack)-1]
	stack.data = theStack[:len(theStack)-1]
	return value
}

// CalStack 计算栈
type CalStack struct {
	NumStack   *Stack[any]
	OpStack    *Stack[any]
	qf         *QuaFormList
	Result     any
	QuaStack   *Stack[any] //四元式栈，用于解决if和while语句嵌套问题，这个栈放的是判断条件产生的四元式
	IfQuaStack *Stack[any] //if四元式栈，这个栈放的是一个if语句结束后需要跳转的四元式
}

// NewCalStack 创建计算栈
func NewCalStack(qf *QuaFormList) *CalStack {
	return &CalStack{
		NumStack: NewStack(),
		OpStack:  NewStack(),
		qf:       qf,
	}
}

// whichLogicOp 判断运算符类型
func (c *CalStack) whichLogicOp(op any) string {
	switch op {
	case consts.QUA_GT:
		return consts.QuaFormMap[consts.QUA_JMPGT]
	case consts.QUA_GE:
		return consts.QuaFormMap[consts.QUA_JMPGE]
	case consts.QUA_LT:
		return consts.QuaFormMap[consts.QUA_JMPLT]
	case consts.QUA_LE:
		return consts.QuaFormMap[consts.QUA_JMPLE]
	case consts.QUA_EQ:
		return consts.QuaFormMap[consts.QUA_JMPEQ]
	case consts.QUA_NE:
		return consts.QuaFormMap[consts.QUA_JMPNE]
	default:
		return ""
	}
}

// isRelOp 判断是否为关系运算符
func (c *CalStack) isRelOp(op any) bool {
	switch op {
	case consts.QUA_GT, consts.QUA_GE, consts.QUA_LT, consts.QUA_LE, consts.QUA_EQ, consts.QUA_NE:
		return true
	default:
		return false
	}
}

// isLogicOp 判断是否为逻辑运算符
func (c *CalStack) isLogicOp(op any) bool {
	switch op {
	case consts.QUA_AND, consts.QUA_OR:
		return true
	default:
		return false
	}
}

// priority 优先级判断，数字越小优先级越高,当栈顶运算符优先级大于等于当前运算符时，栈顶运算符出栈并进行一次运算
func (c *CalStack) priority(op any) int {
	switch op {
	case consts.QUA_LEFTSMALLBRACKET, consts.QUA_RIGHTSMALLBRACKET:
		return 1
	case consts.QUA_NOT, consts.QUA_NEGATIVE, consts.QUA_POSITIVE:
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
	case consts.QUA_PARAM:
		return 9
	case consts.QUA_CALL:
		return 10
	case consts.QUA_ASSIGNMENT:
		return 11
	default:
		return 12
	}

}

// PushFuncCall 压入一个函数调用
func (c *CalStack) PushFuncCall(funcName string) {
	c.OpStack.Push(consts.QUA_CALL)
	c.NumStack.Push(funcName)
}

// PushNum 入数字栈
func (c *CalStack) PushNum(value any) {
	c.NumStack.Push(value)
}

// PushOp 入操作符栈
func (c *CalStack) PushOp(ope int) {
	if c.qf.IfFlag && c.isLogicOp(ope) {
		c.OpStack.Push(ope)
		c.CalIf()
		return
	}

	top := c.OpStack.Top()
	for c.priority(top) <= c.priority(ope) {
		c.Cal()
		top = c.OpStack.Top()
	}
	c.OpStack.Push(ope)
}

// Cal 对数字栈顶两个元素进行一次计算,遇到"#","@","!"只取栈顶一个元素进行一次计算
func (c *CalStack) Cal() {
	if c.qf.IfFlag {
		c.CalIf()
		return
	}
	op := c.OpStack.Pop()
	num2 := c.NumStack.Pop()
	if consts.QUA_NOT == op || consts.QUA_NEGATIVE == op {
		result := c.qf.GetTemp()
		c.qf.AddQuaForm(consts.QuaFormMap[op.(int)], num2, nil, result)
		c.NumStack.Push(result)
		return
	}
	if consts.QUA_PARAM == op {
		c.qf.AddQuaForm(consts.QuaFormMap[op.(int)], num2, nil, nil)
		return
	}

	if consts.QUA_CALL == op {
		if c.NumStack.IsEmpty() {
			c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_FUNCCALL], num2, nil, nil)
		} else {
			result := c.qf.GetTemp()
			c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_FUNCCALL], num2, nil, result)
			c.NumStack.Push(result)
		}
		return
	}

	num1 := c.NumStack.Pop()
	// 遇到赋值运算符，num1为变量，num2为值
	if op == consts.QUA_ASSIGNMENT {
		c.qf.AddQuaForm(consts.QuaFormMap[op.(int)], num2, nil, num1)
		c.Result = num2
		return
	}

	result := c.qf.GetTemp()
	c.qf.AddQuaForm(consts.QuaFormMap[op.(int)], num1, num2, result)
	c.NumStack.Push(result)
}

// CalIf 对数字栈顶两个元素进行一次计算,遇到"#","@","!"只取栈顶一个元素进行一次计算
func (c *CalStack) CalIf() {
	op := c.OpStack.Pop()
	num2 := c.NumStack.Pop()
	if consts.QUA_NOT == op || consts.QUA_NEGATIVE == op {
		result := c.qf.GetTemp()
		c.qf.AddQuaForm(consts.QuaFormMap[op.(int)], num2, nil, result)
		c.NumStack.Push(result)
		return
	}
	if consts.QUA_PARAM == op {
		c.qf.AddQuaForm(consts.QuaFormMap[op.(int)], num2, nil, nil)
		return
	}

	if consts.QUA_CALL == op {
		if c.NumStack.IsEmpty() {
			c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_FUNCCALL], num2, nil, nil)
		} else {
			result := c.qf.GetTemp()
			c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_FUNCCALL], num2, nil, result)
			c.NumStack.Push(result)
		}
		return
	}

	if consts.QUA_AND == op {
		id := c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JT], num2, nil, nil)
		c.qf.QuaForms[id].Result = id + 2
		c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, nil)
		c.QuaStack.Push(id + 1)
		return

	}

	if consts.QUA_OR == op {
		id := c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JF], num2, nil, nil)
		c.qf.QuaForms[id].Result = id + 2
		c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, nil)
		c.QuaStack.Push(id + 1)
		return

	}

	num1 := c.NumStack.Pop()
	// 遇到赋值运算符，num1为变量，num2为值
	if op == consts.QUA_ASSIGNMENT {
		c.qf.AddQuaForm(consts.QuaFormMap[op.(int)], num2, nil, num1)
		c.Result = num2
		return
	}

	if c.isRelOp(op) {
		id := c.qf.AddQuaForm(c.whichLogicOp(op), num1, num2, nil)
		c.qf.QuaForms[id].Result = id + 2
		c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, nil)
		c.QuaStack.Push(id + 1)
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

// CalAllUtilCall 计算直到函数调用符出栈
func (c *CalStack) CalAllUtilCall() {
	for c.OpStack.Top() != consts.QUA_CALL {
		c.Cal()
	}
	c.Cal()
}

func (c *CalStack) Clear() {
	c.NumStack = NewStack()
	c.OpStack = NewStack()
}

func (c *CalStack) ClearCurrentQuaStack() {
	id := c.qf.NextQuaFormId()
	for !c.QuaStack.IsEmpty() {
		i := c.QuaStack.Pop().(int)
		c.qf.QuaForms[i].Result = id
	}
}

func (c *CalStack) ClearCurrentIfStack() {
	id := c.qf.NextQuaFormId()
	for !c.IfQuaStack.IsEmpty() {
		i := c.IfQuaStack.Pop().(int)
		c.qf.QuaForms[i].Result = id
	}
}

// CalStacks 计算栈集合
type CalStacks struct {
	CurrentStack      *CalStack   //当前计算栈
	BracketStack      *Stack[any] //括号栈，每遇到一个左括号，就新建一个计算栈，遇到右括号，就将计算栈出栈
	CurrentQuaStack   *Stack[any] //当前四元式栈，用于存放还不能确定跳转位置的四元式
	QuaStack          *Stack[any] //四元式栈，用于解决if和while语句嵌套问题，这个栈放的是判断条件产生的四元式
	CurrentIfQuaStack *Stack[any] //当前if四元式栈，用于存放还不能确定跳转位置的四元式
	IfQuaStack        *Stack[any] //if四元式栈，这个栈放的是一个if语句结束后需要跳转的四元式
	qf                *QuaFormList
	Result            any
}

// NewCalStacks 创建计算栈集合
func NewCalStacks(qf *QuaFormList) *CalStacks {
	current := NewCalStack(qf)
	c := &CalStacks{
		CurrentStack: current,
		BracketStack: NewStack(),
		qf:           qf,
		QuaStack:     NewStack(),
		IfQuaStack:   NewStack(),
	}
	c.BracketStack.Push(current)
	return c
}

func (c *CalStacks) PushFuncCall(funcName string) {
	c.CurrentStack.PushFuncCall(funcName)
}

func (c *CalStacks) PushNum(value any) {
	c.CurrentStack.PushNum(value)
}

func (c *CalStacks) PushOpe(ope int) {
	if ope == consts.QUA_LEFTSMALLBRACKET {
		current := NewCalStack(c.qf)
		current.QuaStack = c.CurrentQuaStack
		current.IfQuaStack = c.CurrentIfQuaStack
		c.BracketStack.Push(current)
		c.CurrentStack = current

		return
	}
	if ope == consts.QUA_RIGHTSMALLBRACKET {
		c.CurrentStack.CalAll()
		tempResult := c.CurrentStack.NumStack.Top()
		c.BracketStack.Pop()
		c.CurrentStack = c.BracketStack.Top().(*CalStack)
		if c.CurrentStack != nil && !c.qf.IfFlag { //栈不为空，说明当前只是计算完了一个括号内的表达式，要生成一个临时变量放入当前数字栈
			c.CurrentStack.PushNum(tempResult)
		}
		return
	}
	c.CurrentStack.PushOp(ope)
}

func (c *CalStacks) CalAll() {
	c.CurrentStack.CalAll()
	c.Result = c.CurrentStack.Result
}

func (c *CalStacks) CalAllUtilCall() {
	c.CurrentStack.CalAllUtilCall()
}

func (c *CalStacks) Clear() {
	c.CurrentStack.Clear()
}

func (c *CalStacks) ClearCurrentQuaStack() {
	c.CurrentStack.ClearCurrentQuaStack()
	c.QuaStack.Pop()
	if c.QuaStack.Top() != nil {
		c.CurrentQuaStack = c.QuaStack.Top().(*Stack[any])
		c.CurrentStack.QuaStack = c.CurrentQuaStack
	}
}

func (c *CalStacks) ClearCurrentIfStack() {
	c.CurrentStack.ClearCurrentIfStack()

}

func (c *CalStacks) PopCurrentIfStack() {
	c.IfQuaStack.Pop()
	if c.IfQuaStack.Top() != nil {
		c.CurrentIfQuaStack = c.IfQuaStack.Top().(*Stack[any])
		c.CurrentStack.IfQuaStack = c.CurrentIfQuaStack
	}
}

func (c *CalStacks) PushQuaStack(stack *Stack[any]) {
	c.QuaStack.Push(stack)
	c.CurrentQuaStack = stack
	c.CurrentStack.QuaStack = stack

}

func (c *CalStacks) PushIfStack(stack *Stack[any]) {
	c.IfQuaStack.Push(stack)
	c.CurrentIfQuaStack = stack
	c.CurrentStack.IfQuaStack = stack

}
