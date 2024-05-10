package util

import (
	"fmt"
)

// quaForm 四元式
type quaForm struct {
	Id     int //四元式编号
	Op     any
	Arg1   any
	Arg2   any
	Result any
}

// QuaFormList 四元式列表
type QuaFormList struct {
	QuaForms []*quaForm
	Count    int  // 临时变量计数
	IfFlag   bool //标记当前是否在处理if语句的判断条件
}

// NewQuaFormList 创建四元式列表
func NewQuaFormList() *QuaFormList {
	return &QuaFormList{
		QuaForms: make([]*quaForm, 0),
		Count:    0,
	}
}

// AddQuaForm 创建四元式,并返回四元式编号
func (q *QuaFormList) AddQuaForm(op, arg1, arg2, result any) int {
	form := &quaForm{
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
	return fmt.Sprintf("T%d", q.Count)
}

// GetQuaFormList 获取四元式列表
func (q *QuaFormList) GetQuaFormList() []*quaForm {
	return q.QuaForms
}

// GetQuaForm 获取四元式
func (q *QuaFormList) GetQuaForm(index int) *quaForm {
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
