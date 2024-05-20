package compiler

import (
	"complier/util"
	"fmt"
	"strconv"
	"strings"
)

type DAGNode struct {
	Op      string
	Value   any
	Labels  []string
	Left    *DAGNode
	Right   *DAGNode
	Parents []*DAGNode
	IsConst bool // 是否是常数
}

type DAG struct {
	Qf           *util.QuaFormList   // 四元式列表
	DAGQf        *util.QuaFormList   // DAG优化后的四元式列表
	Roots        []*DAGNode          // 根节点
	entryPoints  map[int]bool        // 入口语句
	blocks       [][]*util.QuaForm   // 基本块列表
	currentBlock []*util.QuaForm     // 当前基本块
	LabelMap     map[string]*DAGNode // 标签映射
}

func NewDAG(qf *util.QuaFormList) *DAG {
	if qf == nil {
		qf = util.NewQuaFormList()
	}
	block := make([]*util.QuaForm, 0)
	blocks := make([][]*util.QuaForm, 0)
	return &DAG{
		Qf:           qf,
		entryPoints:  make(map[int]bool),
		blocks:       blocks,
		currentBlock: block,
		LabelMap:     make(map[string]*DAGNode),
	}
}

func (d *DAG) StartDAG(input string) {
	d.parseQuaternions(input)
	d.partitionBasicBlocks()
}

func (d *DAG) isInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
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
			if d.isInt(parts[i]) {
				param[i], _ = strconv.Atoi(parts[i])
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
	if op == "jmp" || op == "j>" || op == "j<" || op == "j<=" || op == "j>=" || op == "j!=" || op == "j==" {
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

// 优化基本块
func (d *DAG) optimizeBasicBlocks() {
	for _, block := range d.blocks {
		d.optimizeBasicBlock(block)
	}
}

// 优化基本块
func (d *DAG) optimizeBasicBlock(block []*util.QuaForm) {
	// 1. 构造DAG
	d.buildDAG(block)
	// 2. 优化DAG
	d.optimizeDAG()
	// 3. 生成优化后的四元式
	d.generateOptimizedQuaternions()

}

// 构造DAG
func (d *DAG) buildDAG(block []*util.QuaForm) {
	// 1. 构造DAG
	for _, q := range block {
		if q.Op == "=" {
			// 如果是赋值语句，则将赋值语句的结果作为DAG的根节点
			node := &DAGNode{
				Op:      "=",
				Value:   q.Arg1,
				Labels:  []string{q.Result.(string)},
				Parents: make([]*DAGNode, 0),
			}
			if n, ok := q.Arg1.(int); ok {
				node.Value = n
				node.IsConst = true
			}
			d.Roots = append(d.Roots, node)
			d.LabelMap[q.Result.(string)] = node // 将标签映射到节点
		} else {
			// 如果是其他语句，则将语句的结果作为DAG的根节点
			node := &DAGNode{
				Op:      q.Op.(string),
				Value:   q.Result,
				Labels:  []string{q.Result.(string)},
				Parents: make([]*DAGNode, 0),
			}
			d.Roots = append(d.Roots, node)
		}
	}
}

// 优化DAG
func (d *DAG) optimizeDAG() {
	// 1. 优化DAG
	for _, root := range d.Roots {
		d.optimizeNode(root)
	}
}

// 优化节点
func (d *DAG) optimizeNode(node *DAGNode) {
	// 1. 优化左子树
	if node.Left != nil {
		d.optimizeNode(node.Left)
	}
	// 2. 优化右子树
	if node.Right != nil {
		d.optimizeNode(node.Right)
	}
	// 3. 优化当前节点
	d.optimizeCurrentNode(node)

}

// 优化当前节点
func (d *DAG) optimizeCurrentNode(node *DAGNode) {
	// 1. 如果当前节点的左右子树都是叶子节点，则判断是否可以合并
	if node.Left != nil && node.Right != nil {
		if len(node.Left.Labels) == 1 && len(node.Right.Labels) == 1 {
			// 如果左右子树的标签相同，则可以合并
			if node.Left.Labels[0] == node.Right.Labels[0] {
				node.Value = node.Left.Value
				node.Op = "="
				node.Left = nil
				node.Right = nil
			}
		}
	}

}

// 生成优化后的四元式
func (d *DAG) generateOptimizedQuaternions() {
	// 1. 生成优化后的四元式
	for _, root := range d.Roots {
		d.generateOptimizedQuaternionsFromNode(root)
	}
}

// 从节点生成优化后的四元式
func (d *DAG) generateOptimizedQuaternionsFromNode(node *DAGNode) {
	// 1. 从左子树生成优化后的四元式
	if node.Left != nil {
		d.generateOptimizedQuaternionsFromNode(node.Left)
	}
	// 2. 从右子树生成优化后的四元式
	if node.Right != nil {
		d.generateOptimizedQuaternionsFromNode(node.Right)
	}
	// 3. 从当前节点生成优化后的四元式
	d.generateOptimizedQuaternionsFromCurrentNode(node)
}

// 从当前节点生成优化后的四元式
func (d *DAG) generateOptimizedQuaternionsFromCurrentNode(node *DAGNode) {
	// 1. 如果当前节点是赋值语句，则直接生成四元式
	if node.Op == "=" {
		d.DAGQf.AddQuaForm("=", node.Value, nil, node.Labels[0])
	} else {
		// 如果当前节点不是赋值语句，则生成四元式
		d.DAGQf.AddQuaForm(node.Op, node.Left.Labels[0], node.Right.Labels[0], node.Labels[0])
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
