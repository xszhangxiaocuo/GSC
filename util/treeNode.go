package util

import (
	"fmt"
	"github.com/awalterschulze/gographviz"
	"strconv"
)

var treeStr string
var count int

type TreeNode struct {
	Token    *TokenNode
	Value    string
	Children []*TreeNode
}

func NewTreeNode(token *TokenNode, value string) *TreeNode {
	return &TreeNode{Token: token, Value: value}
}

// AddChild 添加子节点
func (node *TreeNode) AddChild(child *TreeNode) {
	if child != nil {
		node.Children = append(node.Children, child)
	}

}

// GetTree 获取树结构的字符串
func GetTree(node *TreeNode) string {
	count = 0
	treeStr = ""
	PrintTree(node, "", true)

	////绘制图片
	//graphAst := gographviz.NewEscape()
	//graphAst.SetName("syntax_tree")
	//graphAst.SetDir(true)
	//AddNode(graphAst, node, "")
	//
	//graph := graphAst.String()
	//
	//cmd := exec.Command("dot", "-Tpng", "-o", "pkg/tree_img/tree_"+GetTIme()+".png")
	//var stderr bytes.Buffer
	//cmd.Stderr = &stderr // 捕获标准错误输出
	//cmd.Stdin = strings.NewReader(graph)
	//if err := cmd.Run(); err != nil {
	//	fmt.Println("Error:", err, "Stderr:", stderr.String())
	//}

	return treeStr
}

// PrintTree 递归遍历并打印树
func PrintTree(node *TreeNode, prefix string, isLast bool) {
	if node == nil {
		return
	}
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

func AddNode(graph *gographviz.Escape, node *TreeNode, parent string) string {
	nodeName := "node" + strconv.Itoa(count)
	count++
	graph.AddNode("G", nodeName, map[string]string{"label": node.Value})
	if parent != "" {
		graph.AddEdge(parent, nodeName, true, nil)
	}
	for _, child := range node.Children {
		AddNode(graph, child, nodeName)
	}
	return nodeName
}
