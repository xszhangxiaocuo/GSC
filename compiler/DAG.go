package compiler

import (
	"bytes"
	"complier/util"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"os/exec"
	"strconv"
	"strings"
)

type DAGNode struct {
	Id        int // 节点编号
	Op        any
	MainLabel any          // 主标签（如果是叶子结点，则主标签为初值）
	Labels    map[any]bool // 附加标签
	Left      *DAGNode
	Right     *DAGNode
	Parents   []*DAGNode
	IsConst   bool
}

type DAG struct {
	Qf           *util.QuaFormList  // 四元式列表
	DAGQf        *util.QuaFormList  // DAG优化后的四元式列表
	Roots        []*DAGNode         // 根节点
	entryPoints  map[int]bool       // 入口语句
	blockEntry   []int              // 优化后的基本块入口四元式id
	blocks       [][]*util.QuaForm  // 基本块列表
	currentBlock []*util.QuaForm    // 当前基本块
	LabelMap     map[any][]*DAGNode // 标签映射
	NodeList     [][]*DAGNode       // 节点列表
	CurrentList  []*DAGNode         // 当前节点列表
	JmpMap       map[int]int        // 跳转语句映射,key为跳转语句在输入四元式中的编号，value为跳转语句在优化后的四元式中的编号
}

// NewDAG 创建一个新的DAG
func NewDAG(qf *util.QuaFormList) *DAG {
	if qf == nil {
		qf = util.NewQuaFormList()
	}
	block := make([]*util.QuaForm, 0)
	blocks := make([][]*util.QuaForm, 0)
	return &DAG{
		Qf:           util.NewQuaFormList(),
		DAGQf:        util.NewQuaFormList(),
		entryPoints:  make(map[int]bool),
		LabelMap:     make(map[any][]*DAGNode),
		blocks:       blocks,
		currentBlock: block,
		NodeList:     make([][]*DAGNode, 0),
		CurrentList:  make([]*DAGNode, 0),
		JmpMap:       make(map[int]int),
	}
}

func (d *DAG) StartDAG(input string) {
	d.parseQuaternions(input)
	d.partitionBasicBlocks()
	d.Optimize()
	//绘制图片
	graphAst := gographviz.NewEscape()
	graphAst.SetName("syntax_tree")
	graphAst.SetDir(true)

	for _, nodes := range d.NodeList {
		nodeMap = make(map[*DAGNode]string)
		for _, node := range nodes { // 初始化节点
			AddNode(graphAst, node)
		}
		for _, node := range nodes { // 添加边
			AddEdge(graphAst, node)
		}
		graph := graphAst.String()

		cmd := exec.Command("dot", "-Tpng", "-o", "pkg/dag_img/DAG_"+util.GetTIme()+".png")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr // 捕获标准错误输出
		cmd.Stdin = strings.NewReader(graph)
		if err := cmd.Run(); err != nil {
			fmt.Println("Error:", err, "Stderr:", stderr.String())
		}
	}

}

// NextNodeId 下一个节点id
func (d *DAG) NextNodeId() int {
	return len(d.NodeList)
}

func (d *DAG) isInt(s any) bool {
	_, ok := s.(int)
	return ok
}

func (d *DAG) isFloat(s string) bool {
	_, err := strconv.ParseFloat(s, 16)
	return err == nil
}

// 解析输入的四元式代码
func (d *DAG) parseQuaternions(input string) {
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		param := make([]any, 4)
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
			if atoi, err := strconv.Atoi(parts[i]); err == nil {
				param[i] = atoi
			} else if d.isFloat(parts[i]) {
				param[i], _ = strconv.ParseFloat(parts[i], 16)
			} else if parts[i] == "_" || parts[i] == "" {
				param[i] = nil
			} else {
				param[i] = parts[i]
			}
		}
		d.Qf.AddQuaForm(param[0], param[1], param[2], param[3])
	}
	fmt.Println(d.Qf.PrintQuaFormList())
}

// 判断是否是转移语句
func isTransferStatement(op string) bool {
	if op == "jz" || op == "jnz" || op == "jmp" || op == "j>" || op == "j<" || op == "j<=" || op == "j>=" || op == "j!=" || op == "j==" {
		return true
	}
	return false
}

// 判断是否是操作符，否则为函数名
func isOp(op string) bool {
	if op == "=" || op == "+" || op == "-" || op == "*" || op == "/" || op == "%" || op == "&" || op == "|" || op == "&&" || op == "||" || op == "<" || op == ">" || op == "<=" || op == ">=" || op == "==" || op == "!=" || op == "para" || op == "call" || op == "sys" || op == "ret" || isTransferStatement(op) {
		return true
	}
	return false
}

// partitionBasicBlocks 划分基本块
func (d *DAG) partitionBasicBlocks() {
	if len(d.Qf.QuaForms) == 0 {
		return
	}

	// 1. 确定各个基本块的入口语句
	d.entryPoints[0] = true
	for i, q := range d.Qf.QuaForms {
		if isTransferStatement(q.Op.(string)) {
			target := q.Result.(int)
			if target < len(d.Qf.QuaForms) {
				d.entryPoints[target] = true
			}
			if i+1 < len(d.Qf.QuaForms) {
				d.entryPoints[i+1] = true
			}
		}
	}

	// 2. 构造基本块
	for i, q := range d.Qf.QuaForms {
		// 如果是入口语句或者转移语句并且当前基本块不为空，则将当前基本块加入基本块列表
		if d.entryPoints[i] && len(d.currentBlock) > 0 {
			d.blocks = append(d.blocks, d.currentBlock)
			d.currentBlock = make([]*util.QuaForm, 0)
		}
		d.currentBlock = append(d.currentBlock, q)
		// 如果是转移语句，则将当前基本块加入基本块列表
		if isTransferStatement(q.Op.(string)) {
			d.blocks = append(d.blocks, d.currentBlock)
			d.currentBlock = make([]*util.QuaForm, 0)
		}
	}
	if len(d.currentBlock) > 0 {
		d.blocks = append(d.blocks, d.currentBlock)
	}

}

// addNode 向DAG中添加一个新节点
func (d *DAG) addNode(op any, mainLabel any, left, right *DAGNode, isConst bool) *DAGNode {
	node := &DAGNode{
		Id:        d.NextNodeId(),
		Op:        op,
		MainLabel: mainLabel,
		Labels:    make(map[any]bool),
		Left:      left,
		Right:     right,
		IsConst:   isConst,
	}
	d.CurrentList = append(d.CurrentList, node)
	d.LabelMap[mainLabel] = append(d.LabelMap[mainLabel], node)
	d.deleteLabel(node, mainLabel) // 变量值做了修改，删除原来的附加标签
	return node
}

// findNode 查找具有相同操作符和子节点的现有节点
func (d *DAG) findNode(op any, left, right *DAGNode) *DAGNode {
	// 需要逆序遍历节点列表，因为最新的节点在列表的最后面
	for i := len(d.CurrentList) - 1; i >= 0; i-- {
		node := d.CurrentList[i]
		if node.Op == op && node.Left == left && node.Right == right {
			return node
		}
	}
	return nil
}

// Optimize 优化所有基本块
func (d *DAG) Optimize() {
	for _, block := range d.blocks {
		d.optimizeBlock(block)
		// 清空节点列表和标签映射
		d.NodeList = append(d.NodeList, d.CurrentList)
		d.CurrentList = make([]*DAGNode, 0)
		d.LabelMap = make(map[any][]*DAGNode)
	}
}

// optimizeBlock 优化单个基本块
func (d *DAG) optimizeBlock(block []*util.QuaForm) {
	d.currentBlock = block
	for _, qf := range block {
		switch qf.Op {
		case "+", "-", "*", "/", "%", "&", "|", "&&", "||", "<", ">", "<=", ">=", "==", "!=":
			if d.isInt(qf.Arg1) && d.isInt(qf.Arg2) { // 如果两个操作数都是整数，则直接计算结果
				var result int
				switch qf.Op {
				case "+":
					result = qf.Arg1.(int) + qf.Arg2.(int)
				case "-":
					result = qf.Arg1.(int) - qf.Arg2.(int)
				case "*":
					result = qf.Arg1.(int) * qf.Arg2.(int)
				case "/":
					result = qf.Arg1.(int) / qf.Arg2.(int)
				case "%":
					result = qf.Arg1.(int) % qf.Arg2.(int)
				case "&":
					result = qf.Arg1.(int) & qf.Arg2.(int)
				case "|":
					result = qf.Arg1.(int) | qf.Arg2.(int)
				case "&&":
					if qf.Arg1.(int) != 0 && qf.Arg2.(int) != 0 {
						result = 1
					} else {
						result = 0
					}
				case "||":
					if qf.Arg1.(int) != 0 || qf.Arg2.(int) != 0 {
						result = 1
					} else {
						result = 0
					}
				case "<":
					if qf.Arg1.(int) < qf.Arg2.(int) {
						result = 1
					} else {
						result = 0
					}
				case ">":
					if qf.Arg1.(int) > qf.Arg2.(int) {
						result = 1
					} else {
						result = 0
					}
				case "<=":
					if qf.Arg1.(int) <= qf.Arg2.(int) {
						result = 1
					} else {
						result = 0
					}
				case ">=":
					if qf.Arg1.(int) >= qf.Arg2.(int) {
						result = 1
					} else {
						result = 0
					}
				case "==":
					if qf.Arg1.(int) == qf.Arg2.(int) {
						result = 1
					} else {
						result = 0
					}
				case "!=":
					if qf.Arg1.(int) != qf.Arg2.(int) {
						result = 1
					} else {
						result = 0
					}
				}
				node := d.getOrAddNode(result)
				d.addLabel(node, qf.Result) // 添加附加标签
			} else {
				left := d.getOrAddNode(qf.Arg1)
				right := d.getOrAddNode(qf.Arg2)
				existingNode := d.findNode(qf.Op, left, right)
				if existingNode != nil {
					d.addLabel(existingNode, qf.Result) // 添加附加标签
				} else {
					d.addNode(qf.Op, qf.Result, left, right, false)
				}
			}
		case "=":
			node := d.getOrAddNode(qf.Arg1)
			d.addLabel(node, qf.Result) // 添加附加标签
		case "@", "!":
			var result int
			if v, ok := qf.Arg1.(int); ok { // 如果操作数是整数，则直接计算结果
				if qf.Op == "@" {
					result = -v
				} else if qf.Op == "!" {
					if v == 0 {
						result = 1
					} else {
						result = 0
					}
				}
				node := d.getOrAddNode(result)
				d.addLabel(node, qf.Result) // 添加附加标签
			} else {
				node := d.getOrAddNode(qf.Arg1)
				existingNode := d.findNode(qf.Op, node, nil)
				if existingNode != nil {
					d.addLabel(existingNode, qf.Result) // 添加附加标签
				} else {
					d.addNode(qf.Op, qf.Result, node, nil, false)
				}
			}
		default:
			// 其他未处理的操作也直接添加到优化后的四元式列表中
			d.JmpMap[qf.Id] = d.getBlockId(qf.Id) // 先记录在原四元式中的编号与基本块的对应信息，在优化结束后根据基本块号更新编号
		}
	}
	d.generateOptimizedQuaForms()
}

// getBlockId 根据四元式id获取基本块编号
func (d *DAG) getBlockId(id int) int {
	for i, block := range d.blocks {
		for _, q := range block {
			if q.Id == id {
				return i
			}
		}
	}
	return -1
}

// getOrAddNode 获取已存在的节点或添加一个新节点
func (d *DAG) getOrAddNode(label any) *DAGNode {
	if label == nil {
		return nil
	}
	// 需要逆序遍历节点列表，因为最新的节点在列表的最后面
	for i := len(d.CurrentList) - 1; i >= 0; i-- {
		// 如果节点的主标签等于label或者节点的附加标签中包含label，则返回该节点
		if d.CurrentList[i].MainLabel == label || d.CurrentList[i].Labels[label] {
			return d.CurrentList[i]
		}

	}
	if _, ok := label.(int); ok {
		return d.addNode(nil, label, nil, nil, true)
	}
	return d.addNode(nil, label, nil, nil, false)
}

// 判断标签的优先级
func getPriority(label string) int {
	if _, err := strconv.Atoi(label); err == nil {
		return 3 // 常量优先级最高
	} else if strings.HasPrefix(label, "T") || strings.HasPrefix(label, "$") {
		return 1 // 临时变量优先级最低
	}
	return 2 // 变量优先级中等
}

// addLabel 添加附加标签
func (d *DAG) addLabel(node *DAGNode, label any) {
	if node.MainLabel == label { // 如果label等于当前节点的主标签，则直接返回
		return
	}
	// 根据优先级判断是否需要更新主标签，如果label的优先级大于当前节点的主标签优先级，则将label赋值给当前节点的主标签，优先级顺序：常量 > 变量 > 临时变量
	if node.MainLabel == nil { // 如果当前节点的主标签为空，则直接将label赋值给主标签
		node.MainLabel = label
	} else if _, ok := label.(int); ok { // 如果label是常量，则直接将label赋值给主标签
		if _, ok := node.MainLabel.(int); ok { // 如果当前节点的主标签也是常量，直接赋值
			delete(d.LabelMap, node.MainLabel)
			node.MainLabel = label
		} else { // 否则将主标签移动到附加标签中
			node.Labels[node.MainLabel] = true
			node.MainLabel = label
		}
	} else if _, ok := node.MainLabel.(int); ok { // 如果当前节点的主标签是常量，直接将label添加到附加标签
		node.Labels[label] = true
		d.deleteLabel(node, label) // 变量值做了修改，删除原来的附加标签
	} else if getPriority(label.(string)) > getPriority(node.MainLabel.(string)) { // 如果label的优先级大于当前节点的主标签优先级，则将label赋值给当前节点的主标签
		node.Labels[node.MainLabel] = true
		node.MainLabel = label
		d.deleteLabel(node, label) // 变量值做了修改，删除原来的附加标签
	} else {
		node.Labels[label] = true
		d.deleteLabel(node, label) // 变量值做了修改，删除原来的附加标签
	}

	d.LabelMap[label] = append(d.LabelMap[label], node)
}

// deleteLabel 删除附加标签 TODO: 删除d.LabelMap中的标签
func (d *DAG) deleteLabel(node *DAGNode, label any) {
	for _, n := range d.CurrentList {
		if n == node {
			continue
		}
		if _, ok := n.Labels[label]; ok { // 附加标签存在，删除该标签
			delete(n.Labels, label)
		}
	}
}

// generateOptimizedQuaForms 生成优化后的四元式列表
func (d *DAG) generateOptimizedQuaForms() {
	for _, node := range d.CurrentList {
		if node.Op == nil { // 没有操作符且附加标签中有自定义变量，需要将自定义变量赋值给主标签
			for label, _ := range node.Labels {
				if getPriority(label.(string)) == 2 {
					d.DAGQf.AddQuaForm("=", node.MainLabel, nil, label)
				}
			}
		} else { // 有操作符，直接添加到优化后的四元式列表中
			left := node.Left
			right := node.Right
			if node.Op == "@" || node.Op == "!" || node.Op == "jz" || node.Op == "jnz" || node.Op == "=" {
				d.DAGQf.AddQuaForm(node.Op, left.MainLabel, nil, node.MainLabel)
			} else if node.Op == "para" {
				d.DAGQf.AddQuaForm(node.Op, node.MainLabel, nil, nil)
			} else if node.Op == "call" || node.Op == "ret" {
				if node.MainLabel != nil { //有返回值
					d.DAGQf.AddQuaForm(node.Op, left.MainLabel, nil, node.MainLabel)
				} else {
					d.DAGQf.AddQuaForm(node.Op, left.MainLabel, nil, nil)
				}
			} else if node.Op == "sys" || !isOp(node.Op.(string)) { // 系统结束或者函数定义
				d.DAGQf.AddQuaForm(node.Op, nil, nil, nil)
			} else if node.Op == "jmp" {
				d.DAGQf.AddQuaForm(node.Op, nil, nil, node.MainLabel)
			} else {
				d.DAGQf.AddQuaForm(node.Op, left.MainLabel, right.MainLabel, node.MainLabel)
			}
		}
	}
}

// PrintBasicBlocks 打印基本块
func (d *DAG) PrintBasicBlocks() string {
	str := ""
	for i, block := range d.blocks {
		str += fmt.Sprintf("基本块 %d:\n", i)
		for _, qf := range block {
			str += fmt.Sprintf("%d\t", qf.Id)
			str += fmt.Sprintf("%s\t\t", qf.Op)
			arg1, ok := qf.Arg1.(string)
			if ok {
				str += fmt.Sprintf("%s\t\t", arg1)
			} else {
				if qf.Arg1 == nil {
					str += fmt.Sprintf("<nil>\t\t")
				} else if num, ok := qf.Arg1.(int); ok {
					str += fmt.Sprintf("%d\t\t", num)
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
				} else if num, ok := qf.Arg2.(int); ok {
					str += fmt.Sprintf("%d\t\t", num)
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
			result += fmt.Sprintf("\n")
		}
	}
	return str
}

var count int
var nodeMap map[*DAGNode]string

func AddNode(graph *gographviz.Escape, node *DAGNode) {
	if node == nil {
		return
	}
	nodeName := "node" + strconv.Itoa(count)
	count++
	var str string
	str += "\""
	if node.Op != nil && node.Op != "" {
		str += fmt.Sprintf("%s|", node.Op)
	}
	if node.MainLabel != nil {
		str += fmt.Sprintf("%v", node.MainLabel)
	}

	//if len(node.Labels) != 0 {
	//	str += "|"
	//}
	//for label, _ := range node.Labels {
	//	str += fmt.Sprintf("%v ", label)
	//}

	str += "\""
	graph.AddNode("G", nodeName, map[string]string{"label": str})
	nodeMap[node] = nodeName
}

func AddEdge(graph *gographviz.Escape, node *DAGNode) {
	if node == nil {
		return
	}
	if node.Left != nil {
		graph.AddEdge(nodeMap[node], nodeMap[node.Left], true, nil)
	}
	if node.Right != nil && node.Left != node.Right {
		graph.AddEdge(nodeMap[node], nodeMap[node.Right], true, nil)
	}
}
