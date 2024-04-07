package util

import (
	"fmt"
)

var treeStr string

type TreeNode struct {
	Value    string
	Children []*TreeNode
}

func NewTreeNode(value string) *TreeNode {
	return &TreeNode{Value: value}
}

// AddChild 添加子节点
func (node *TreeNode) AddChild(child *TreeNode) {
	node.Children = append(node.Children, child)
}

// GetTree 获取树结构的字符串
func GetTree(node *TreeNode) string {
	PrintTree(node, "", true)
	return treeStr
}

// PrintTree 递归遍历并打印树
func PrintTree(node *TreeNode, prefix string, isLast bool) {
	// 打印当前节点
	if isLast {
		//fmt.Printf("%s└── %s\n", prefix, node.Value)
		treeStr += fmt.Sprintf("%s└── %s\n", prefix, node.Value)
		prefix += "    "
	} else {
		//fmt.Printf("%s├── %s\n", prefix, node.Value)
		treeStr += fmt.Sprintf("%s├── %s\n", prefix, node.Value)
		prefix += "│   "
	}

	// 递归打印子节点
	for i, child := range node.Children {
		PrintTree(child, prefix, i == len(node.Children)-1)
	}
}
