package util

import (
	"fmt"
)

// QuaForm 四元式
type QuaForm struct {
	Id     int // 四元式编号
	Op     any
	Arg1   any
	Arg2   any
	Result any
}

// ForJmpPos 记录循环的条件判断位置，以及第二个赋值表达式位置
type ForJmpPos struct {
	ConditionPos int //条件判断位置
	AssignPos    int //第二个赋值表达式位置
	ContinuePos  int //遇到continue的退出位置
}

func NewForJmpPos() *ForJmpPos {
	return &ForJmpPos{
		ConditionPos: -1,
		AssignPos:    -1,
		ContinuePos:  -1,
	}
}

// QuaFormList 四元式列表
type QuaFormList struct {
	QuaForms             []*QuaForm
	Count                int         // 临时变量计数
	IfFlag               bool        // 标记当前是否在处理if语句的判断条件
	JmpPoint             *Stack[any] // 标记循环的起始位置的四元式编号
	BreakStacks          *Stack[any]
	ContinueStacks       *Stack[any]
	CurrentBreakStack    *Stack[any] // 需要回填的break四元式编号
	CurrentContinueStack *Stack[any] // 需要回填的continue四元式编号
	RelaOp               bool        // 标记当前运算过程中是否有关系运算符
}

// NewQuaFormList 创建四元式列表
func NewQuaFormList() *QuaFormList {
	return &QuaFormList{
		QuaForms:       make([]*QuaForm, 0),
		Count:          0,
		JmpPoint:       NewStack(),
		BreakStacks:    NewStack(),
		ContinueStacks: NewStack(),
	}
}

// PushBreakStack 压入break栈
func (q *QuaFormList) PushBreakStack(stack *Stack[any]) {
	q.BreakStacks.Push(stack)
	q.CurrentBreakStack = stack
}

// ClearBreakStack 清空break栈
func (q *QuaFormList) ClearBreakStack(id int) {
	for q.CurrentBreakStack.Top() != nil {
		top := q.CurrentBreakStack.Pop().(int)
		q.QuaForms[top].Result = id
	}
	q.BreakStacks.Pop()
	if q.BreakStacks.Top() != nil {
		q.CurrentBreakStack = q.BreakStacks.Top().(*Stack[any])
	}
}

// PushContinue 压入continue栈
func (q *QuaFormList) PushContinue(stack *Stack[any]) {
	q.ContinueStacks.Push(stack)
	q.CurrentContinueStack = stack
}

// ClearContinueStack 清空continue栈
func (q *QuaFormList) ClearContinueStack(id int) {
	for q.CurrentContinueStack.Top() != nil {
		top := q.CurrentContinueStack.Pop().(int)
		q.QuaForms[top].Result = id
	}
	q.ContinueStacks.Pop()
	if q.ContinueStacks.Top() != nil {
		q.CurrentContinueStack = q.ContinueStacks.Top().(*Stack[any])
	}
}

// AddQuaForm 创建四元式,并返回四元式编号
func (q *QuaFormList) AddQuaForm(op, arg1, arg2, result any) int {
	form := &QuaForm{
		Id:     q.NextQuaFormId(),
		Op:     op,
		Arg1:   arg1,
		Arg2:   arg2,
		Result: result,
	}
	q.QuaForms = append(q.QuaForms, form)
	return form.Id
}

// NextQuaFormId 获取下一个四元式编号
func (q *QuaFormList) NextQuaFormId() int {
	return len(q.QuaForms)
}

// GetTemp 获取临时变量
func (q *QuaFormList) GetTemp() string {
	defer func() {
		q.Count++
	}()
	return fmt.Sprintf("$T%d", q.Count)
}

// GetQuaFormList 获取四元式列表
func (q *QuaFormList) GetQuaFormList() []*QuaForm {
	return q.QuaForms
}

// GetQuaForm 获取四元式
func (q *QuaFormList) GetQuaForm(index int) *QuaForm {
	return q.QuaForms[index]
}

// GetQuaFormLength 获取四元式列表长度
func (q *QuaFormList) GetQuaFormLength() int {
	return len(q.QuaForms)
}

// PrintQuaFormList 打印四元式列表
func (q *QuaFormList) PrintQuaFormList() string {
	str := "四元式列表：\nid\top\t\targ1\t\targ2\t\tresult\n"
	for i, qf := range q.QuaForms {
		str += fmt.Sprintf("%d\t", i)
		str += fmt.Sprintf("%s\t\t", qf.Op)
		arg1, ok := qf.Arg1.(string)
		if ok {
			str += fmt.Sprintf("%s\t\t", arg1)
		} else {
			if qf.Arg1 == nil {
				str += fmt.Sprintf("<nil>\t\t")
			} else {
				str += fmt.Sprintf("%c\t\t", qf.Arg1)
			}
		}
		arg2, ok := qf.Arg2.(string)
		if ok {
			str += fmt.Sprintf("%s\t\t", arg2)
		} else {
			if qf.Arg2 == nil {
				str += fmt.Sprintf("<nil>\t\t")
			} else {
				str += fmt.Sprintf("%c\t\t", qf.Arg2)
			}
		}
		result, ok := qf.Result.(string)
		if ok {
			str += fmt.Sprintf("%s\n", result)
		} else {
			if result, ok := qf.Result.(int); ok {
				str += fmt.Sprintf("%d\n", result)
			} else if qf.Result == nil {
				str += fmt.Sprintf("<nil>\n")
			} else {
				str += fmt.Sprintf("%c\n", qf.Result)
			}
		}
	}
	return str
}
