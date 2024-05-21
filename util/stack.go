package util

import (
	"complier/pkg/consts"
	"log"
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

// LogicStack 逻辑栈
type LogicStack struct {
	TrueStack  *Stack[any] //真出口栈
	FalseStack *Stack[any] //假出口栈
	qf         *QuaFormList
}

// NewLogicStack 创建逻辑栈
func NewLogicStack(qf *QuaFormList) *LogicStack {
	return &LogicStack{
		TrueStack:  NewStack(),
		FalseStack: NewStack(),
		qf:         qf,
	}
}

// ClearTrueStack 清空真出口栈
func (l *LogicStack) ClearTrueStack(nextId int) {
	for !l.TrueStack.IsEmpty() {
		id := l.TrueStack.Pop().(int)
		l.qf.QuaForms[id].Result = nextId
	}
}

// ClearFalseStack 清空假出口栈
func (l *LogicStack) ClearFalseStack(nextId int) {
	for !l.FalseStack.IsEmpty() {
		id := l.FalseStack.Pop().(int)
		l.qf.QuaForms[id].Result = nextId
	}
}

// CalStack 计算栈
type CalStack struct {
	NumStack   *Stack[any]
	OpStack    *Stack[any]
	qf         *QuaFormList
	Result     any
	IfQuaStack *Stack[any] //if四元式栈，这个栈放的是一个if语句结束后需要跳转的四元式，跳转到一个完整的if语句的结束位置
	currentOp  any         //当前运算符
	LogicStack *LogicStack //逻辑栈
}

// NewCalStack 创建计算栈
func NewCalStack(qf *QuaFormList) *CalStack {
	return &CalStack{
		NumStack:   NewStack(),
		OpStack:    NewStack(),
		qf:         qf,
		LogicStack: NewLogicStack(qf),
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
	case consts.QUA_MOVE, consts.QUA_NORELA:
		return 7
	case consts.QUA_AND:
		return 8
	case consts.QUA_OR:
		return 9
	case consts.QUA_PARAM:
		return 10
	case consts.QUA_CALL:
		return 11
	case consts.QUA_ASSIGNMENT:
		return 12
	default:
		return 13
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
	if c.qf.IfFlag {
		c.PushIfOp(ope)
		return
	}

	top := c.OpStack.Top()
	for !(top == consts.QUA_PARAM && ope == consts.QUA_PARAM) && c.priority(top) <= c.priority(ope) {
		c.Cal()
		top = c.OpStack.Top()
	}
	c.OpStack.Push(ope)
}

// PushIfOp 入操作符栈
func (c *CalStack) PushIfOp(ope int) {
	c.currentOp = ope //当前要入栈的运算符
	defer func() {
		c.currentOp = nil
	}()
	if c.isRelOp(ope) {
		c.qf.RelaOp = true
	}

	top := c.OpStack.Top()
	for c.priority(top) <= c.priority(ope) {
		c.Cal()
		top = c.OpStack.Top()
	}

	//&&和||运算符不入栈
	if ope != consts.QUA_AND && ope != consts.QUA_OR {
		c.OpStack.Push(ope)
	} else if c.qf.IfFlag && c.qf.RelaOp == false {
		c.OpStack.Push(consts.QUA_NORELA)
		c.Cal()
	}
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
		if !c.qf.FuncCall { //函数调用语句，不需要保存返回值
			c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_FUNCCALL], num2, nil, nil)
		} else { //函数调用表达式，需要保存返回值
			result := c.qf.GetTemp()
			c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_FUNCCALL], num2, nil, result)
			c.NumStack.Push(result)
		}
		c.qf.FuncCall = false
		return
	}

	if op == consts.QUA_RETURN {
		c.Result = num2
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

	if op == consts.QUA_MOVE {
		c.move()
		return
	}

	num2 := c.NumStack.Pop()

	if op == consts.QUA_NORELA {
		if num2 == nil {
			return
		}
		if consts.QUA_AND == c.currentOp {
			id := c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JT], num2, nil, nil)
			c.qf.QuaForms[id].Result = id + 2
			c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, nil)
			c.LogicStack.FalseStack.Push(id + 1)
			return

		}

		if consts.QUA_OR == c.currentOp {
			id := c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JF], num2, nil, nil)
			c.qf.QuaForms[id].Result = id + 2
			c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, nil)
			c.LogicStack.TrueStack.Push(id + 1)
			return

		}

		// 当前没有运算符入栈
		if c.currentOp == nil {
			id := c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JT], num2, nil, nil)
			c.LogicStack.TrueStack.Push(id)
			c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JF], num2, nil, nil)
			c.LogicStack.FalseStack.Push(id + 1)
			return
		}

	}

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

	if c.isRelOp(op) {
		c.qf.RelaOp = false
		if c.currentOp == consts.QUA_AND {
			id := c.qf.AddQuaForm(c.whichLogicOp(op), num1, num2, nil)
			c.qf.QuaForms[id].Result = id + 2
			c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, nil)
			c.LogicStack.FalseStack.Push(id + 1)
		} else if c.currentOp == consts.QUA_OR {
			id := c.qf.AddQuaForm(c.whichLogicOp(op), num1, num2, nil)
			c.LogicStack.TrueStack.Push(id)
			c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, nil)
			c.qf.QuaForms[id+1].Result = id + 2
			// 遇到||运算符，清空假出口栈
			c.LogicStack.ClearFalseStack(c.qf.NextQuaFormId())
		} else if c.currentOp == nil {
			id := c.qf.AddQuaForm(c.whichLogicOp(op), num1, num2, nil)
			c.LogicStack.TrueStack.Push(id)
			c.qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, nil)
			c.LogicStack.FalseStack.Push(id + 1)
		}

		return
	}

	result := c.qf.GetTemp()
	c.qf.AddQuaForm(consts.QuaFormMap[op.(int)], num1, num2, result)
	c.NumStack.Push(result)

}

func (c *CalStack) move() {
	stack, ok := c.NumStack.Pop().(*LogicStack) // 此处取出的必须是逻辑栈
	if !ok {
		log.Println("move error")
		return
	}
	// 当前运算符为&&，将真出口栈中的四元式的结果设置为下一个四元式的id
	if c.currentOp == consts.QUA_AND {
		for !stack.TrueStack.IsEmpty() {
			id := stack.TrueStack.Pop().(int)
			c.qf.QuaForms[id].Result = c.qf.NextQuaFormId()
		}
		// 将上一个括号内传递出来的假出口栈中的四元式id放入当前逻辑栈的假出口栈中
		for !stack.FalseStack.IsEmpty() {
			id := stack.FalseStack.Pop().(int)
			c.LogicStack.FalseStack.Push(id)
		}
		return
	}
	// 当前运算符为||，将假出口栈中的四元式的结果设置为下一个四元式的id
	if c.currentOp == consts.QUA_OR {
		for !stack.FalseStack.IsEmpty() {
			id := stack.FalseStack.Pop().(int)
			c.qf.QuaForms[id].Result = c.qf.NextQuaFormId()
		}
		// 将上一个括号内传递出来的真出口栈中的四元式id放入当前逻辑栈的真出口栈中
		for !stack.TrueStack.IsEmpty() {
			id := stack.TrueStack.Pop().(int)
			c.LogicStack.TrueStack.Push(id)
		}
		return
	}
	// 当前没有运算符入栈，将上一个括号传递出来的真出口和假出口中的四元式id放到当前逻辑栈中
	if c.currentOp == nil {
		for !stack.TrueStack.IsEmpty() {
			id := stack.TrueStack.Pop().(int)
			c.LogicStack.TrueStack.Push(id)
		}
		for !stack.FalseStack.IsEmpty() {
			id := stack.FalseStack.Pop().(int)
			c.LogicStack.FalseStack.Push(id)
		}
		return
	}
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

// CalAllUtilReturn 计算直到return用符出栈
func (c *CalStack) CalAllUtilReturn() {
	for c.OpStack.Top() != consts.QUA_RETURN {
		c.Cal()
	}
	c.Cal()
}

func (c *CalStack) Clear() {
	c.NumStack = NewStack()
	c.OpStack = NewStack()
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
	CurrentStack      *CalStack   // 当前计算栈
	BracketStack      *Stack[any] // 括号栈，每遇到一个左括号，就新建一个计算栈，遇到右括号，就将计算栈出栈
	IfQuaStack        *Stack[any] // if四元式栈，这个栈放的是一个if语句结束后需要跳转的四元式
	CurrentIfQuaStack *Stack[any] // 当前if四元式栈，用于存放还不能确定跳转位置的四元式
	LogicStack        *Stack[any] // 逻辑栈
	CurrentLogicStack *LogicStack // 当前if语句的逻辑栈
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
		IfQuaStack:   NewStack(),
		LogicStack:   NewStack(),
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
		if c.qf.IfFlag {
			c.LogicStack.Push(current.LogicStack)
			c.CurrentLogicStack = current.LogicStack
		}

		c.BracketStack.Push(current)
		c.CurrentStack = current

		return
	}
	if ope == consts.QUA_RIGHTSMALLBRACKET {
		if c.qf.RelaOp == false && c.qf.IfFlag {
			c.CurrentStack.OpStack.Push(consts.QUA_NORELA)
			c.CurrentStack.Cal()
		}
		c.CurrentStack.CalAll()
		tempResult := c.CurrentStack.NumStack.Top()
		logicStack := c.CurrentStack.LogicStack
		c.BracketStack.Pop()
		c.CurrentStack = c.BracketStack.Top().(*CalStack)
		if c.qf.IfFlag {
			c.CurrentStack.NumStack.Push(logicStack) // 将括号内的逻辑栈传递给上一个计算栈
			c.CurrentStack.OpStack.Push(consts.QUA_MOVE)
		}

		if c.CurrentStack != nil && !c.qf.IfFlag { // 栈不为空，说明当前只是计算完了一个括号内的表达式，要生成一个临时变量放入当前数字栈
			c.CurrentStack.PushNum(tempResult)
		}
		return
	}
	c.CurrentStack.PushOp(ope)
}

func (c *CalStacks) CalIf() {
	c.CurrentStack.Cal() // 计算一次move操作

	for !c.CurrentStack.LogicStack.TrueStack.IsEmpty() {
		id := c.CurrentStack.LogicStack.TrueStack.Pop().(int)
		c.CurrentLogicStack.TrueStack.Push(id)
	}
	for !c.CurrentStack.LogicStack.FalseStack.IsEmpty() {
		id := c.CurrentStack.LogicStack.FalseStack.Pop().(int)
		c.CurrentLogicStack.FalseStack.Push(id)
	}
}

func (c *CalStacks) CalAll() {
	c.CurrentStack.CalAll()
	//c.CurrentStack.Result = c.CurrentStack.NumStack.Top()
	c.Result = c.CurrentStack.Result
}

func (c *CalStacks) CalAllUtilCall() {
	c.CurrentStack.CalAllUtilCall()
}

func (c *CalStacks) CalAllUtilReturn() {
	c.CurrentStack.CalAllUtilReturn()
	c.Result = c.CurrentStack.Result
}

func (c *CalStacks) Clear() {
	c.CurrentStack.Clear()
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

func (c *CalStacks) PushIfStack(stack *Stack[any]) {
	c.IfQuaStack.Push(stack)
	c.CurrentIfQuaStack = stack
	c.CurrentStack.IfQuaStack = stack

}

func (c *CalStacks) PushLogicStack(stack *LogicStack) {
	c.LogicStack.Push(stack)
	c.CurrentLogicStack = stack
	c.CurrentStack.LogicStack = stack
}

func (c *CalStacks) PopCurrentLogicStack() {
	c.LogicStack.Pop()
	if c.LogicStack.Top() != nil {
		c.CurrentLogicStack = c.LogicStack.Top().(*LogicStack)
		c.CurrentStack.LogicStack = c.CurrentLogicStack
	}
}

func (c *CalStacks) ClearTrueStack(id int) {
	c.CurrentLogicStack.ClearTrueStack(id)
}

func (c *CalStacks) ClearFalseStack(id int) {
	c.CurrentLogicStack.ClearFalseStack(id)
}
