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
	Count    int // 临时变量计数
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
func (q *QuaFormList) PrintQuaFormList() {
	str := ""
	for i, qf := range q.QuaForms {
		str += fmt.Sprintf("%d: ", i)
		str += fmt.Sprintf("%s ", qf.Op)
		arg1, ok := qf.Arg1.(string)
		if ok {
			str += fmt.Sprintf("%s ", arg1)
		} else {
			if qf.Arg1 == nil {
				str += fmt.Sprintf("<nil> ")
			} else {
				str += fmt.Sprintf("%c ", qf.Arg1)
			}
		}
		arg2, ok := qf.Arg2.(string)
		if ok {
			str += fmt.Sprintf("%s ", arg2)
		} else {
			if qf.Arg2 == nil {
				str += fmt.Sprintf("<nil> ")
			} else {
				str += fmt.Sprintf("%c ", qf.Arg2)
			}
		}
		result, ok := qf.Result.(string)
		if ok {
			str += fmt.Sprintf("%s\n", result)
		} else {
			str += fmt.Sprintf("%c\n", qf.Result)
		}
		//fmt.Printf("%d: %s %s %s %s\n", i, qf.Op, arg1, arg2, result)
		//fmt.Printf("%d: %s %s %s %s\n", i, qf.Op, qf.Arg1, qf.Arg1, qf.Result)
	}
	fmt.Println(str)
}
