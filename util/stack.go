package util

import "errors"

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
func (stack *Stack[any]) Push(value interface{}) {
	stack.data = append(stack.data, value)
}

// Top 查看栈顶元素
func (stack *Stack[any]) Top() (interface{}, error) {
	if len(stack.data) == 0 {
		return nil, errors.New("栈为空")
	}
	return stack.data[len(stack.data)-1], nil
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
