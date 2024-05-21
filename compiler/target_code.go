package compiler

import (
	"complier/pkg/consts"
	"complier/pkg/logger"
	"complier/util"
	"fmt"
	"strconv"
	"strings"
)

type Target struct {
	Qf             *util.QuaFormList            // 四元式列表
	SymbolTable    *SymbolTable                 // 符号表
	CurrentQuaForm *util.QuaForm                // 当前四元式
	CurrentFunc    string                       // 当前函数名
	CurrentId      int                          // 当前四元式的索引
	Asm            strings.Builder              // 汇编代码字符串
	logger         *logger.Logger               // 日志记录器
	FuncMap        map[string]map[string]string // 函数参数和局部变量的地址映射
	FuncParamLen   int                          // 当前函数参数和局部变量的长度
	FuncParamNum   int                          // 函数形参个数
	FuncTempNum    int                          // 函数临时变量个数（包括局部变量以及临时参数）
}

func NewTarget(qf *util.QuaFormList, table *SymbolTable) *Target {
	return &Target{
		Qf:          qf,
		Asm:         strings.Builder{},
		SymbolTable: table,
		logger:      logger.NewLogger(),
		FuncMap:     make(map[string]map[string]string),
	}
}

// GenerateAsmCode 生成目标代码
func (t *Target) GenerateAsmCode() {
	t.CurrentFunc = "main"

	// 生成汇编代码头
	t.Asm.WriteString(consts.ASM_HEAD)

	// 生成全局变量
	for funcName, table := range t.SymbolTable.VarTable {
		if funcName != consts.ALL && funcName != "main" {
			continue
		}
		for name, _ := range table {
			t.Asm.WriteString(fmt.Sprintf("\t_%s dw 0\n", name))
		}

	}

	// 生成全局常量
	for funcName, table := range t.SymbolTable.ConstTable {
		if funcName != consts.ALL && funcName != "main" {
			continue
		}
		for name, info := range table {
			t.Asm.WriteString(fmt.Sprintf("\t_%s dw %s\n", name, info.Value))
		}
	}

	// 生成汇编代码入口
	t.Asm.WriteString(consts.ASM_START)

	for i, form := range t.Qf.QuaForms {
		op := form.Op
		arg1 := form.Arg1
		arg2 := form.Arg2
		result := form.Result

		if op == "main" {
			continue
		}

		switch op {
		case "=":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tMOV %s,AX\n", i, t.DataAdress(arg1), t.DataAdress(result)))
		case "+":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tADD AX,%s\n\tMOV %s,AX\n", i, t.DataAdress(arg1), t.DataAdress(arg2), t.DataAdress(result)))
		case "-", "@":
			if op == "@" { //求负运算，0-arg
				arg2 = arg1
				arg1 = "0"
			}
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tSUB AX,%s\n\tMOV %s,AX\n", i, t.DataAdress(arg1), t.DataAdress(arg2), t.DataAdress(result)))
		case "*":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tMOV BX,%s\n\tMUL BX\n\tMOV %s,AX\n", i, t.DataAdress(arg1), t.DataAdress(arg2), t.DataAdress(result)))
		case "/":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tMOV DX,0\n\tMOV BX,%s\n\tDIV BX\n\tMOV %s,AX\n", i, t.DataAdress(arg1), t.DataAdress(arg2), t.DataAdress(result)))
		case "%":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tMOV DX,0\n\tMOV BX,%s\n\tDIV BX\n\tMOV %s,DX\n", i, t.DataAdress(arg1), t.DataAdress(arg2), t.DataAdress(result)))
		case "<":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV DX,1\n\tMOV AX,%s\n\tCMP AX,%s\n\tJL _GT_%d\n\tMOV DX,0\n_GT_%d:\tMOV %s,DX\n", i, t.DataAdress(arg1), t.DataAdress(arg2), i, i, t.DataAdress(result)))
		case "<=":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV DX,1\n\tMOV AX,%s\n\tCMP AX,%s\n\tJLE _LE_%d\n\tMOV DX,0\n_LE_%d:\tMOV %s,DX\n", i, t.DataAdress(arg1), t.DataAdress(arg2), i, i, t.DataAdress(result)))
		case ">":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV DX,1\n\tMOV AX,%s\n\tCMP AX,%s\n\tJG _LT_%d\n\tMOV DX,0\n_LT_%d:\tMOV %s,DX\n", i, t.DataAdress(arg1), t.DataAdress(arg2), i, i, t.DataAdress(result)))
		case ">=":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV DX,1\n\tMOV AX,%s\n\tCMP AX,%s\n\tJGE _GE_%d\n\tMOV DX,0\n_GE_%d:\tMOV %s,DX\n", i, t.DataAdress(arg1), t.DataAdress(arg2), i, i, t.DataAdress(result)))
		case "==":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV DX,1\n\tMOV AX,%s\n\tCMP AX,%s\n\tJE _EQ_%d\n\tMOV DX,0\n_EQ_%d:\tMOV %s,DX\n", i, t.DataAdress(arg1), t.DataAdress(arg2), i, i, t.DataAdress(result)))
		case "!=":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV DX,1\n\tMOV AX,%s\n\tCMP AX,%s\n\tJNE _NE_%d\n\tMOV DX,0\n_NE_%d:\tMOV %s,DX\n", i, t.DataAdress(arg1), t.DataAdress(arg2), i, i, t.DataAdress(result)))
		case "j<":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tCMP AX,%s\n\tjl _%d\n", i, t.DataAdress(arg1), t.DataAdress(arg2), result))
		case "j>=":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tCMP AX,%s\n\tjge _%d\n", i, t.DataAdress(arg1), t.DataAdress(arg2), result))
		case "j>":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tCMP AX,%s\n\tjg _%d\n", i, t.DataAdress(arg1), t.DataAdress(arg2), result))
		case "j<=":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tCMP AX,%s\n\tjle _%d\n", i, t.DataAdress(arg1), t.DataAdress(arg2), result))
		case "j==":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tCMP AX,%s\n\tje _%d\n", i, t.DataAdress(arg1), t.DataAdress(arg2), result))
		case "j!=":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tCMP AX,%s\n\tjne _%d\n", i, t.DataAdress(arg1), t.DataAdress(arg2), result))
		case "&&":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV DX,0\n\tMOV AX,%s\n\tCMP AX,0\n\tJE _AND_%d\n\tMOV AX,%s\n\tCMP AX,0\n\tJE _AND_%d\n\tMOV DX,1\n_AND_%d:\tMOV %s,DX\n", i, t.DataAdress(arg1), i, t.DataAdress(arg2), i, i, t.DataAdress(result)))
		case "||":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV DX,1\n\tMOV AX,%s\n\tCMP AX,0\n\tJNE _OR_%d\n\tMOV AX,%s\n\tCMP AX,0\n\tJNE _OR_%d\n\tMOV DX,0\n_OR_%d:\tMOV %s,DX\n", i, t.DataAdress(arg1), i, t.DataAdress(arg2), i, i, t.DataAdress(result)))
		case "!":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV DX,1\n\tMOV AX,%s\n\tCMP AX,0\n\tJE _NOT_%d\n\tMOV DX,0\n_NOT_%d:\tMOV %s,DX\n", i, t.DataAdress(arg1), i, i, t.DataAdress(result)))
		case "jmp":
			jmp := "_" + strconv.Itoa(result.(int))
			// 跳转的位置为程序结束时，跳转到退出程序的位置
			if t.Qf.GetQuaForm(result.(int)).Op == "sys" {
				jmp = "quit"
			}
			t.Asm.WriteString(fmt.Sprintf("_%d:\tJMP far ptr %s\n", i, jmp))
		case "jz":
			jmp := "_" + result.(string)
			// 跳转的位置为程序结束时，跳转到退出程序的位置
			if t.Qf.GetQuaForm(result.(int)).Op == "sys" {
				jmp = "quit"
			}
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tCMP AX,0\n\tJNE _NE_%d\n\tJMP far ptr %s\n_NE_%d:\tNOP\n", i, t.DataAdress(arg1), i, jmp, i))
		case "jnz":
			jmp := "_" + result.(string)
			// 跳转的位置为程序结束时，跳转到退出程序的位置
			if t.Qf.GetQuaForm(result.(int)).Op == "sys" {
				jmp = "quit"
			}
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tCMP AX,0\n\tJE _EZ_%d\n\tJMP far ptr %s\n_EZ_%d:\tNOP\n", i, t.DataAdress(arg1), i, jmp, i))
		case "para":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tPUSH AX\n", i, t.DataAdress(arg1)))
		case "call":
			t.Asm.WriteString(fmt.Sprintf("_%d:\tCALL %s\n", i, arg1))
			if result != nil { // 函数调用有返回值
				t.Asm.WriteString(fmt.Sprintf("\tMOV %s,AX\n", t.DataAdress(result)))
			}
		case "ret":
			if result != nil { // 函数返回有返回值
				t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV AX,%s\n\tMOV SP,BP\n\tPOP BP\n\tRET\n", i, t.DataAdress(result)))
			} else {
				t.Asm.WriteString(fmt.Sprintf("_%d:\tMOV SP,BP\n\tPOP BP\n\tRET\n", i))
			}
		case "sys":
			t.Asm.WriteString("quit:\tMOV AH,4Ch\n\tINT 21h\n")
		default: // 函数定义
			t.CurrentFunc = op.(string)
			t.CurrentId = i + 1
			t.getFuncParamLen()
			t.Asm.WriteString(fmt.Sprintf("%s:\tPUSH BP\n\tMOV BP,SP\n\tSUB SP,%d\n", op, t.FuncParamLen))
		}
	}

	t.Asm.WriteString(consts.ASM_END)
}

// isFuncDef 判断当前四元式是否为函数定义
func (t *Target) isFuncDef(op any) bool {
	ope := op.(string)
	if ope == "=" || ope == "+" || ope == "-" || ope == "*" || ope == "/" || ope == "%" || ope == "<" || ope == "<=" || ope == ">" || ope == ">=" || ope == "==" || ope == "!=" || ope == "j<" || ope == "j>=" || ope == "j>" || ope == "j<=" || ope == "j==" || ope == "j!=" || ope == "&&" || ope == "||" || ope == "!" || ope == "jmp" || ope == "jz" || ope == "jnz" || ope == "para" || ope == "call" || ope == "ret" || ope == "sys" {
		return false
	}
	return true
}

// getFuncParamAddr 获取函数的参数对应的地址
func (t *Target) getFuncParamAddr(param string) (string, bool) {
	addr, ok := t.FuncMap[t.CurrentFunc][param]
	return addr, ok
}

// setFuncParamAddr 设置函数的参数对应的地址
func (t *Target) setFuncParamAddr(param any) {
	if param == nil {
		return
	}
	var p string
	var ok bool
	if p, ok = param.(string); !ok {
		return
	}
	if _, ok = t.getFuncParamAddr(p); !ok && !t.isGlobalVar(p) { // 查询不到参数地址并且不是全局变量
		if t.SymbolTable.VarTable[t.CurrentFunc][p] != nil && t.SymbolTable.VarTable[t.CurrentFunc][p].ParamFlag { // 是函数形参
			t.FuncParamLen += 2
			t.FuncMap[t.CurrentFunc][p] = fmt.Sprintf("ss:[bp+%d]", 4+t.FuncParamNum*2) // 函数形参地址, 从bp+4开始,bp+2为返回地址,bp+0为bp
			t.FuncParamNum++
		} else if !t.isDigit(p) { // 非常量数字，是局部变量或临时变量
			t.FuncParamLen += 2
			t.FuncMap[t.CurrentFunc][p] = fmt.Sprintf("ss:[bp-%d]", 2+t.FuncTempNum*2) // 局部变量地址, 从bp-2开始
			t.FuncTempNum++
		}
	}
}

// initFuncParamAddr 初始化函数的形参对应的地址
func (t *Target) initFuncParamAddr() {
	for _, name := range t.SymbolTable.FuncTable[t.CurrentFunc].ParsName {
		t.FuncParamLen += 2
		t.FuncMap[t.CurrentFunc][name] = fmt.Sprintf("ss:[bp+%d]", 4+t.FuncParamNum*2) // 函数形参地址, 从bp+4开始,bp+2为返回地址,bp+0为bp
		t.FuncParamNum++
	}
}

// getFuncParamLen 获取函数形参、局部变量及临时变量的长度
func (t *Target) getFuncParamLen() {
	t.FuncParamLen = 0
	t.FuncTempNum = 0
	t.FuncParamNum = 0
	t.FuncMap[t.CurrentFunc] = make(map[string]string)

	// 初始化函数的形参对应的地址
	t.initFuncParamAddr()
	//遍历函数定义之后的四元式，获取函数的参数和局部变量，没有用到的局部变量不会被分配地址
	for i := t.CurrentId; i < len(t.Qf.QuaForms); i++ {
		form := t.Qf.GetQuaForm(i)
		if t.isFuncDef(form.Op) {
			break
		}
		if form.Op != consts.QuaFormMap[consts.QUA_CALL] { // 函数调用第一个参数为函数名，不需要分配地址
			t.setFuncParamAddr(form.Arg1)
		}
		t.setFuncParamAddr(form.Arg2)
		t.setFuncParamAddr(form.Result)
	}
}

// isGlobalVar 判断是否为全局变量或全局常量
func (t *Target) isGlobalVar(varName string) bool {
	if info, ok := t.SymbolTable.VarTable[consts.ALL][varName]; ok {
		if info.Scope == consts.ALL {
			return true
		}
	} else if info, ok = t.SymbolTable.ConstTable[consts.ALL][varName]; ok {
		if info.Scope == consts.ALL {
			return true
		}
	}
	return false
}

// DataAdress 获取参数地址
func (t *Target) DataAdress(arg any) string {
	param := arg.(string)
	p := ""
	if t.CurrentFunc == "main" {
		if param[0] == '$' { // 临时变量，从扩展段的栈中取值
			p = fmt.Sprintf("es:[%d]", t.toInt(param[2:])*2)
		} else if t.isDigit(param) { // 数字，直接取值
			p = param
		} else { // 变量，从数据段中取值
			p = fmt.Sprintf("ds:[_%s]", param)
		}
	} else { // 当前函数不是main函数，从栈中取值
		if t.isDigit(param) { // 数字，直接取值
			p = param
		} else if t.isGlobalVar(param) { // 全局变量或全局常量
			p = fmt.Sprintf("ds:[_%s]", param)
		} else { // 当前函数形参以及局部变量
			p = t.FuncMap[t.CurrentFunc][param]
		}
	}
	return p
}

func (t *Target) equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (t *Target) toInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func (t *Target) isDigit(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
