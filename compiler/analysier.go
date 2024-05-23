package compiler

import (
	"complier/pkg/consts"
	"complier/pkg/logger"
	"complier/util"
	"fmt"
	"strconv"
)

// Param 函数参数
type Param struct {
	Type string
	Name string
}

// Info 符号表信息
type Info struct {
	Scope     string   //作用域范围的函数名
	Name      string   //变量名
	Type      string   //变量类型
	Value     any      //变量值
	Level     int      //变量作用域,0表示为全局
	Pars      []string //如果是函数，需要参数列表
	ParsName  []string //参数名
	initFlag  bool     //标记当前info的value是否已经初始化
	funcFlag  bool     //标记函数是否已经定义
	ParamFlag bool     //标记是否是形参
}

func (i *Info) Copy() *Info {
	return &Info{
		Scope:    i.Scope,
		Name:     i.Name,
		Type:     i.Type,
		Value:    i.Value,
		Level:    i.Level,
		Pars:     i.Pars,
		initFlag: i.initFlag,
		funcFlag: i.funcFlag,
	}
}

// String 返回info的字符串形式
func (i *Info) String() string {
	str := fmt.Sprintf("%s\t\t\t%s\t\t%s\t\t%s\t\t%v", i.Scope, strconv.Itoa(i.Level), i.Name, i.Type, i.Value)
	if len(i.Pars) != 0 {
		str += fmt.Sprintf("\t\t%v", i.Pars)
	}
	return str
}

// SymbolTable 符号表
type SymbolTable struct {
	VarTable   map[string]map[string]*Info //变量表
	ConstTable map[string]map[string]*Info //常量表
	FuncTable  map[string]*Info            //函数表
}

// String 返回符号表的字符串形式
func (s *SymbolTable) String() string {
	str := "变量表: \n作用域\t作用域等级\t\t变量名\t变量类型\t变量值\n"
	for _, table := range s.VarTable {
		for _, v := range table {
			str += v.String() + "\n"
		}
	}
	str += "\n\n常量表: \n作用域\t作用域等级\t\t常量名\t常量类型\t常量值\n"
	for _, table := range s.ConstTable {
		for _, v := range table {
			str += v.String() + "\n"
		}
	}
	str += "\n\n函数表: \n作用域\t作用域等级\t\t函数名\t函数类型\t函数值\t参数列表\n"
	for _, v := range s.FuncTable {
		str += v.String() + "\n"
	}
	return str

}

// NewSymbolTable 创建符号表
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		VarTable:   make(map[string]map[string]*Info),
		ConstTable: make(map[string]map[string]*Info),
		FuncTable:  make(map[string]*Info),
	}

}

// AddVariable 添加变量
func (s *SymbolTable) AddVariable(info *Info) {
	s.VarTable[info.Scope][info.Name] = info
}

// FindVariable 查找变量
func (s *SymbolTable) FindVariable(scope string, name string) (*Info, bool) {
	info, found := s.VarTable[scope][name]
	if !found { //如果在当前作用域找不到变量，就在全局作用域找
		scope = consts.ALL
		info, found = s.VarTable[scope][name]
	}
	return info, found
}

// AddConstant 添加常量
func (s *SymbolTable) AddConstant(info *Info) {
	s.ConstTable[info.Scope][info.Name] = info
}

// FindConstant 查找常量
func (s *SymbolTable) FindConstant(scope string, name string) (*Info, bool) {
	info, found := s.ConstTable[scope][name]
	if !found { //如果在当前作用域找不到常量，就在全局作用域找
		scope = consts.ALL
		info, found = s.ConstTable[scope][name]
	}
	return info, found
}

// AddFunction 添加函数
func (s *SymbolTable) AddFunction(info *Info) {
	s.FuncTable[info.Name] = info
}

// FindFunction 查找函数
func (s *SymbolTable) FindFunction(name string) (*Info, bool) {
	info, found := s.FuncTable[name]
	return info, found
}

// Analyser 语义分析器
type Analyser struct {
	Ast           *util.TreeNode    //语法树
	calStacks     *util.CalStacks   //运算符栈
	SymbolTable   *SymbolTable      //符号表
	Logger        *logger.Logger    //日志记录器
	Level         int               //作用域等级
	Scope         string            //作用域
	info          *Info             //当前传递的info信息
	flag          bool              //标记当前传递的info信息是否已经完整
	err           bool              //标记是否出现错误
	paramFlag     bool              //标记是否有参数
	divFlag       bool              //标记是否有除法
	retFlag       bool              //标记是否有返回值
	divToken      *util.TokenNode   //除法的token
	node          *util.TreeNode    //当前节点
	Qf            *util.QuaFormList //四元式列表
	CurrentJmpPos *util.ForJmpPos   //当前循环的条件判断位置
	currentFunc   string            //当前函数
	params        []Param           //参数列表
}

// NewAnalyser 创建语义分析器
func NewAnalyser(ast *util.TreeNode) *Analyser {
	qf := util.NewQuaFormList()
	return &Analyser{
		Ast:         ast,
		calStacks:   util.NewCalStacks(qf),
		SymbolTable: NewSymbolTable(),
		Logger:      logger.NewLogger(),
		Level:       0,
		Scope:       consts.ALL,
		info:        nil,
		flag:        false,
		err:         false,
		node:        nil,
		Qf:          qf,
	}
}

// isLegalNode 判断是否是合法的节点
func isLegalNode(node *util.TreeNode) bool {
	return !(node == nil || len(node.Children) == 0 || node.Children[0].Value == consts.NULL)
}

// infoFlag 如果一个info信息传递完毕，将info置空
func (a *Analyser) infoFlag() {
	if a.flag {
		a.info = nil
		a.flag = false
	}
}

// addConstTable 添加常量表
func (a *Analyser) addConstTable() {
	defer func() {
		a.flag = true
		a.info = nil
	}()

	if a.info != nil {
		if a.isExist(a.info.Name) {
			a.Logger.AddErr("\t\t\t\t\t\t常量：" + a.info.Name + " 重复定义\n")
			return
		}
		if a.info.Value == nil {
			a.Logger.AddErr("\t\t\t\t\t\t常量：" + a.info.Name + " 未赋值\n")
			return
		}
		a.SymbolTable.AddConstant(a.info)
	}

}

// addVarTable 添加变量表
func (a *Analyser) addVarTable() {
	defer func() {
		a.flag = true
		a.info = nil
	}()
	if a.info != nil {
		if a.isExist(a.info.Name) {
			a.Logger.AddErr("\t\t\t\t\t\t变量：" + a.info.Name + " 重复定义\n")
			return
		}
		//TODO: 变量初始化?
		//if !a.info.initFlag {
		//	if a.isSameType(consts.TokenMap[a.info.Type], consts.TokenMap["int"]) {
		//		a.info.Value = 0
		//	} else if a.isSameType(consts.TokenMap[a.info.Type], consts.TokenMap["float"]) {
		//		a.info.Value = 0.0
		//	} else if a.isSameType(consts.TokenMap[a.info.Type], consts.TokenMap["char"]) {
		//		a.info.Value = ' '
		//	}
		//}
		a.info.Scope = a.Scope
		a.info.Level = a.Level
		a.SymbolTable.AddVariable(a.info)
	}
}

// changeVarTable 修改变量表
func (a *Analyser) changeVarTable() {
	defer func() {
		a.flag = true
		a.info = nil
	}()
	if a.info != nil {
		if !a.isExist(a.info.Name) {
			a.Logger.AddErr("\t\t\t\t\t\t变量：" + a.info.Name + " 未定义\n")
			return
		}
		info := a.SymbolTable.VarTable[a.Scope][a.info.Name].Copy()
		if info.Name != a.info.Value {
			info.Value = a.info.Value
			a.SymbolTable.AddVariable(info)
		}

	}
}

// addFuncTable 添加函数表
func (a *Analyser) addFuncTable() {
	defer func() {
		a.flag = true
		a.info = nil
	}()
	if a.isExist(a.info.Name) {
		a.Logger.AddErr("\t\t\t\t\t\t函数：" + a.info.Name + " 重复定义\n")
	} else {
		a.SymbolTable.AddFunction(a.info)
	}
}

// TODO: 还要检查作用域
// checkVar 在进行表达式运算时检查变量是否合法
func (a *Analyser) checkVar(node *util.TreeNode) bool {
	//一个变量可能是变量表中的变量，也可能是常量表中的常量
	if !a.varIsExist(node.Value) && !a.constIsExist(node.Value) {
		a.Logger.AddAnalyseErr(node.Token, "变量未定义")
		return false
	}
	//TODO: 检查变量类型是否匹配
	//if a.info.Type == "" {
	//	if a.varIsExist(a.info.Name) {
	//		info, _ := a.SymbolTable.FindVariable(a.info.Name)
	//		a.info.Type = info.Type
	//	} else if a.constIsExist(a.info.Name) {
	//		info, _ := a.SymbolTable.FindConstant(a.info.Name)
	//		a.info.Type = info.Type
	//	} else if a.funcIsExist(a.info.Name) {
	//		info, _ := a.SymbolTable.FindFunction(a.info.Name)
	//		a.info.Type = info.Pars[0]
	//	}
	//}
	//
	var v *Info
	if a.varIsExist(node.Value) {
		v, _ = a.SymbolTable.FindVariable(a.Scope, node.Value)
	} else if a.constIsExist(node.Value) {
		v, _ = a.SymbolTable.FindConstant(a.Scope, node.Value)
	} else {
		a.Logger.AddAnalyseErr(node.Token, "变量类型未知")
		return false
	}

	//if v.Type != a.info.Type {
	//	a.Logger.AddAnalyseErr(node.Token, "类型不匹配: ", a.info.Type)
	//	return false
	//}
	//检查变量作用域,只有在同一作用域下或者在更高作用域下才能访问
	if !(v.Level == 0 || v.Scope == a.info.Scope && v.Level <= a.info.Level) {
		a.Logger.AddAnalyseErr(node.Token, "变量作用域不匹配")
		return false
	}
	return true
}

// checkConstNumber 在进行表达式运算时检查常数和变量类型是否一致
func (a *Analyser) checkConstNumber(node *util.TreeNode) bool {
	//if a.info.Type == "" {
	//	if a.varIsExist(a.info.Name) {
	//		a.info.Type = a.SymbolTable.VarTable[a.info.Name].Type
	//	} else if a.funcIsExist(a.currentFunc) {
	//		a.info.Type = a.SymbolTable.FuncTable[a.currentFunc].Type
	//	} else {
	//		a.Logger.AddAnalyseErr(node.Token, "类型不匹配: ", a.info.Type)
	//		return false
	//	}
	//}
	//if node.Token == nil || !a.isSameType(node.Token.Type, consts.TokenMap[a.info.Type]) {
	//	a.Logger.AddAnalyseErr(node.Token, "类型不匹配: ", a.info.Type)
	//	return false
	//}
	return true
}

// checkFunc 在进行函数调用时检查函数是否合法
func (a *Analyser) checkFunc(node *util.TreeNode) bool {
	if !a.funcIsExist(node.Value) {
		a.Logger.AddAnalyseErr(node.Token, "函数未定义")
		return false
	}
	//TODO: 检查函数参数是否匹配
	//if a.info.Type == "" {
	//	a.info.Type = a.SymbolTable.VarTable[a.info.Name].Type
	//}
	//if a.info.Type != a.SymbolTable.FuncTable[node.Value].Type {
	//	a.Logger.AddAnalyseErr(node.Token, "函数返回值类型不匹配")
	//	return false
	//}
	return true
}

// checkFuncCall 检查函数调用和定义时参数是否匹配
func (a *Analyser) checkFuncParam(funcName string) bool {
	if !a.funcIsExist(funcName) {
		a.Logger.AddErr("\t\t\t\t\t\t函数：" + funcName + " 未定义\n")
		return false
	}
	//if len(a.SymbolTable.FuncTable[funcName].Pars) != len(a.params) {
	//	a.Logger.AddErr("\t\t\t\t\t\t函数：" + funcName + " 参数个数不匹配\n")
	//	return false
	//}
	//pars := a.SymbolTable.FuncTable[funcName].Pars
	//for i, v := range a.params {
	//	if pars[i] != v.Type {
	//		a.Logger.AddErr("\t\t\t\t\t\t函数：" + funcName + fmt.Sprintf("%v", a.params) + " 参数类型不匹配，需要类型" + fmt.Sprintf("%v", pars) + "\n")
	//		return false
	//	}
	//}

	return true
}

// checkFuncParamList 检查函数参数列表,并将参数列表存入符号表
func (a *Analyser) checkFuncParamList(funcName string) {
	if !a.funcIsExist(funcName) {
		a.Logger.AddErr("\t\t\t\t\t\t函数：" + funcName + " 未定义\n")
		return
	}
	v, _ := a.SymbolTable.FindFunction(funcName)
	if len(v.Pars) != len(a.params) {
		a.Logger.AddErr("\t\t\t\t\t\t函数：" + funcName + " 参数个数不匹配\n")
		return
	}
	for i, par := range v.Pars {
		if par != a.params[i].Type {
			a.Logger.AddErr("\t\t\t\t\t\t函数：" + funcName + fmt.Sprintf("%v", a.params) + " 参数类型不匹配，需要类型" + fmt.Sprintf("%v", v.Pars) + "\n")
			return
		}
		name := a.params[i].Name
		t := a.params[i].Type
		if a.isExist(name) {
			a.Logger.AddErr("\t\t\t\t\t\t变量：" + name + " 重复定义\n")
			return
		}

		a.SymbolTable.AddVariable(&Info{
			Scope:     a.Scope,
			Level:     a.info.Level + 1,
			Name:      name,
			Type:      t,
			Value:     a.info.Value,
			ParamFlag: true, // 标记为形参
		})
		v.ParsName = append(v.ParsName, name)
	}
}

// isExist 检查符号是否存在
func (a *Analyser) isExist(name string) bool {
	if a.varIsExist(name) || a.constIsExist(name) || a.funcIsExist(name) {
		return true
	}
	return false
}

// varIsExist 检查变量是否存在
func (a *Analyser) varIsExist(name string) bool {
	if _, ok := a.SymbolTable.FindVariable(a.Scope, name); ok {
		return true
	}
	return false
}

// constIsExist 检查常量是否存在
func (a *Analyser) constIsExist(name string) bool {
	if _, ok := a.SymbolTable.FindConstant(a.Scope, name); ok {
		return true
	}
	return false
}

// funcIsExist 检查函数是否存在
func (a *Analyser) funcIsExist(name string) bool {
	if name == "read" || name == "write" {
		return true
	}
	if _, ok := a.SymbolTable.FindFunction(name); ok {
		return true
	}
	return false
}

// isSameType 检查两个类型是否相同
func (a *Analyser) isSameType(t1, t2 consts.Token) bool {
	if t1 == t2 {
		return true
	}
	if (t1 == consts.TokenMap["int"] && t2 == consts.TokenMap["integer"]) || (t1 == consts.TokenMap["integer"] && t2 == consts.TokenMap["int"]) {
		return true
	} else if (t1 == consts.TokenMap["char"] && t2 == consts.TokenMap["character"]) || (t1 == consts.TokenMap["character"] && t2 == consts.TokenMap["char"]) {
		return true
	} else if (t1 == consts.TokenMap["float"] && t2 == consts.TokenMap["floatnumber"]) || (t1 == consts.TokenMap["floatnumber"] && t2 == consts.TokenMap["float"]) {
		return true
	}
	return false
}

// initInfo 初始化info信息
func (a *Analyser) initInfo() {
	a.info = &Info{
		Scope:    a.Scope,
		Level:    a.Level,
		initFlag: false,
		funcFlag: false,
	}
	if a.Level == 0 {
		a.Scope = consts.ALL
		a.info.Scope = a.Scope
	}
}

func (a *Analyser) initValue() {
	//TODO: 变量初始化?
	//switch a.info.Type {
	//case "int":
	//	a.info.Value = "0"
	//case "float":
	//	a.info.Value = "0.0"
	//case "char":
	//	a.info.Value = "' '"
	//}
	a.info.initFlag = true
}

// clearCalStacks 计算栈内的表达式并清空栈
func (a *Analyser) clearCalStacks() {
	a.calStacks.CalAll()
	a.info.Value = a.calStacks.Result
	if a.info.Value == nil {
		a.info.Value = a.calStacks.CurrentStack.NumStack.Pop()
	}
	a.calStacks.Clear()
}

// StartAnalyse 开始语义分析
func (a *Analyser) StartAnalyse() {
	// 初始化全局作用域
	a.SymbolTable.VarTable[consts.ALL] = make(map[string]*Info)
	a.SymbolTable.ConstTable[consts.ALL] = make(map[string]*Info)
	a.analyse(a.Ast, 0)
}

// analyse 递归遍历语法树进行语义分析
func (a *Analyser) analyse(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}

	child := node.Children[next]
	switch child.Value {
	case consts.DECLARATION:
		a.info.Scope = a.Scope
		a.info.Level = a.Level
		a.analyseDeclarationStatement(child, 0)
	case "main":
		// 初始化main函数作用域
		a.SymbolTable.VarTable["main"] = make(map[string]*Info)
		a.SymbolTable.ConstTable["main"] = make(map[string]*Info)
		//添加main函数
		a.Qf.AddQuaForm("main", nil, nil, nil)
		a.SymbolTable.AddFunction(&Info{
			Scope: consts.ALL,
			Name:  "main",
			Level: 0,
			Type:  "void",
		})
		a.Scope = "main" //作用域为main函数
		a.currentFunc = "main"
		a.info.Scope = a.Scope
	case consts.COMPOUND_STMT:
		a.analyseCompoundStatement(child, 0)
		a.Qf.AddQuaForm(consts.QuaFormMap[consts.QUA_SYS], nil, nil, nil)
	case consts.FUNCTION_BLOCK:
		a.analyseFunctionBlock(child, 0)
	}
	a.infoFlag()
	a.analyse(node, next+1)
}

// analyseDeclarationStatement 分析声明语句
func (a *Analyser) analyseDeclarationStatement(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}

	child := node.Children[next]
	switch child.Value {
	case consts.VALUE_DECLARATION:
		a.analyseDeclarationValue(child, 0)
	case consts.FUNCTION_DECL_STMT:
		a.analyseDeclarationFunctionStatement(child, 0)
	}
	a.infoFlag()
	a.analyseDeclarationStatement(node, next+1)
}

// analyseDeclarationValue 分析值声明
func (a *Analyser) analyseDeclarationValue(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}

	child := node.Children[next]
	switch child.Value {
	case consts.CONST_DECLARATION:
		a.analyseDeclarationConst(child, 0)
	case consts.VARIABLE_DECL:
		a.analyseDeclarationVar(child, 0)
	}
	a.infoFlag()
	a.analyseDeclarationValue(node, next+1)
}

// analyseDeclarationConst 分析常量声明
func (a *Analyser) analyseDeclarationConst(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}

	child := node.Children[next]
	switch child.Value {
	case consts.CONST_TYPE:
		a.info.Type = child.Children[0].Value
	case consts.CONST_TABLE:
		a.analyseDeclarationConstTable(child, 0)
	}
	a.infoFlag()
	a.analyseDeclarationConst(node, next+1)
}

// analyseDeclarationConstTable 分析常量声明表
func (a *Analyser) analyseDeclarationConstTable(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}

	child := node.Children[next]
	switch child.Value {
	case consts.VARIABLE:
		a.info.Name = child.Children[0].Value
		a.calStacks.PushNum(child.Children[0].Value) //变量入栈
	case "=":
		a.info.initFlag = true
		a.calStacks.PushOpe(consts.QUA_ASSIGNMENT) //运算符入栈
	case consts.CONST_TABLE_0:
		a.analyseDeclarationConstTable0(child, 0)
	}
	a.infoFlag()
	a.analyseDeclarationConstTable(node, next+1)
}

// analyseDeclarationConstTable0 分析常量声明表0
func (a *Analyser) analyseDeclarationConstTable0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}

	child := node.Children[next]
	switch child.Value {
	case consts.CONST_TABLE_VALUE:
		a.analyseDeclarationConstTableValue(child, 0)
	case consts.CONST_TABLE_1:
		a.analyseDeclarationConstTable1(child, 0)
	}
	a.infoFlag()
	a.analyseDeclarationConstTable0(node, next+1)
}

// analyseDeclarationConstTable1 分析常量声明表1
func (a *Analyser) analyseDeclarationConstTable1(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}

	child := node.Children[next]
	switch child.Value {
	case ";":
		if !a.err {
			//TODO: 变量初始化?
			//if !a.info.initFlag {
			//	a.initValue()
			//	a.calStacks.PushOpe(consts.QUA_ASSIGNMENT)
			//	a.calStacks.PushNum(a.info.Value)
			//}
			a.clearCalStacks()
			a.addConstTable()
		}
		a.err = false
		a.flag = true
		a.calStacks.Clear()
	case ",":
		info := a.info.Copy()
		if !a.err {
			//TODO: 变量初始化?
			//if !a.info.initFlag {
			//	a.initValue()
			//	a.calStacks.PushOpe(consts.QUA_ASSIGNMENT)
			//	a.calStacks.PushNum(a.info.Value)
			//}
			a.clearCalStacks()
			a.addConstTable()
		}
		//继续传递info信息
		a.flag = false
		a.info = &Info{
			Scope:    info.Scope,
			Level:    info.Level,
			Type:     info.Type,
			initFlag: false,
		}
		a.err = false
		a.calStacks.Clear()
	case consts.CONST_TABLE:
		a.analyseDeclarationConstTable(child, 0)
	}
	a.infoFlag()
	a.analyseDeclarationConstTable1(node, next+1)
}

// analyseDeclarationConstTableValue 分析常量声明表值
func (a *Analyser) analyseDeclarationConstTableValue(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}

	child := node.Children[next]
	switch child.Value {
	case consts.VARIABLE:
		//检查变量是否存在，类型是否匹配
		if a.checkVar(child.Children[0]) {
			v, _ := a.SymbolTable.FindVariable(a.Scope, child.Children[0].Value)
			a.info.Value = v.Value            //取出变量值
			a.calStacks.PushNum(a.info.Value) //变量值入栈
		} else {
			a.err = true
		}
		if a.info.Value == nil {
			a.Logger.AddAnalyseErr(child.Children[0].Token, "常量未赋值")
			a.err = true
		}
	case consts.CONSTANT:
		if a.checkConstNumber(child.Children[0].Children[0]) {
			a.info.Value = child.Children[0].Children[0].Value
			a.calStacks.PushNum(child.Children[0].Children[0].Value) //常数入栈
		} else {
			a.err = true
		}
	}
	a.infoFlag()
	a.analyseDeclarationConstTableValue(node, next+1)
}

// analyseDeclarationVar 分析变量声明
func (a *Analyser) analyseDeclarationVar(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.VARIABLE_TYPE:
		a.info.Type = child.Children[0].Value
	case consts.VARIABLE_TABLE:
		a.analyseDeclarationVarTable(child, 0)
	}
	a.infoFlag()
	a.analyseDeclarationVar(node, next+1)
}

// analyseDeclarationVarTable 分析变量声明表
func (a *Analyser) analyseDeclarationVarTable(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.SINGLE_VARIABLE:
		a.analyseDeclarationSingleVar(child, 0)
	case consts.VARIABLE_TABLE_0:
		a.analyseDeclarationVarTable0(child, 0)
	}
	a.infoFlag()
	a.analyseDeclarationVarTable(node, next+1)
}

// analyseDeclarationSingleVar 分析单变量声明
func (a *Analyser) analyseDeclarationSingleVar(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.VARIABLE:
		a.info.Name = child.Children[0].Value
		a.calStacks.PushNum(child.Children[0].Value) //变量名入栈
	case consts.SINGLE_VARIABLE_0:
		a.analyseDeclarationSingleVar0(child, 0)
	}
	a.infoFlag()
	a.analyseDeclarationSingleVar(node, next+1)
}

// analyseDeclarationSingleVar0 分析单变量声明0
func (a *Analyser) analyseDeclarationSingleVar0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "=":
		a.info.initFlag = true
		a.calStacks.PushOpe(consts.QUA_ASSIGNMENT)
	case consts.BOOLEAN_EXPR:
		a.analyseBoolExp(child, 0)
	}
	a.infoFlag()
	a.analyseDeclarationSingleVar0(node, next+1)
}

// analyseDeclarationVarTable0 分析变量声明表0
func (a *Analyser) analyseDeclarationVarTable0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case ";":
		if !a.err {
			//TODO: 变量初始化?
			//if !a.info.initFlag {
			//	a.initValue()
			//	a.calStacks.PushOpe(consts.QUA_ASSIGNMENT)
			//	a.calStacks.PushNum(a.info.Value)
			//}
			if a.calStacks.CurrentStack.OpStack.Top() == consts.QUA_ASSIGNMENT {
				a.clearCalStacks()
			} else {
				a.calStacks.CurrentStack.NumStack.Pop()
			}

			a.addVarTable()
		}
		a.err = false
		a.flag = true
		a.calStacks.Clear()
	case ",":
		info := a.info.Copy()
		if !a.err {
			// TODO: 变量初始化?
			//if !a.info.initFlag {
			//	a.initValue()
			//	a.calStacks.PushOpe(consts.QUA_ASSIGNMENT)
			//	a.calStacks.PushNum(a.info.Value)
			//}
			a.calStacks.CurrentStack.NumStack.Pop()
			a.addVarTable()
		}
		//继续传递info信息
		a.flag = false
		a.info = &Info{
			Scope:    info.Scope,
			Level:    info.Level,
			Type:     info.Type,
			initFlag: false,
		}
		a.err = false
		a.calStacks.Clear()
	case consts.VARIABLE_TABLE:
		a.analyseDeclarationVarTable(child, 0)
	}
	a.infoFlag()
	a.analyseDeclarationVarTable0(node, next+1)
}

// analyseBoolExp 分析布尔表达式
func (a *Analyser) analyseBoolExp(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.BOOLEAN_ITEM:
		a.analyseBoolItem(child, 0)
	case consts.BOOLEAN_EXPR_0:
		a.analyseBoolExp0(child, 0)
	}
	a.infoFlag()
	a.analyseBoolExp(node, next+1)
}

// analyseBoolItem 分析布尔项
func (a *Analyser) analyseBoolItem(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.BOOLEAN_FACTOR:
		a.analyseBoolFactor(child, 0)
	case consts.BOOLEAN_ITEM_0:
		a.analyseBoolItem0(child, 0)
	}
	a.infoFlag()
	a.analyseBoolItem(node, next+1)
}

// analyseBoolExp0 分析布尔项0
func (a *Analyser) analyseBoolItem0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "&&":
		a.calStacks.PushOpe(consts.QUA_AND)
	case consts.BOOLEAN_FACTOR:
		a.analyseBoolFactor(child, 0)
	case consts.BOOLEAN_ITEM_0:
		a.analyseBoolItem0(child, 0)
	}
	a.infoFlag()
	a.analyseBoolItem0(node, next+1)
}

// analyseBoolFactor 分析布尔因子
func (a *Analyser) analyseBoolFactor(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.ARITHMETIC_EXPR:
		a.analyseArithmeticExp(child, 0)
	case consts.BOOLEAN_FACTOR_0:
		a.analyseBoolFactor0(child, 0)
	}
	a.infoFlag()
	a.analyseBoolFactor(node, next+1)
}

// analyseArithmeticExp 分析算术表达式
func (a *Analyser) analyseArithmeticExp(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.TERM:
		a.analyseItem(child, 0)
	case consts.ARITHMETIC_EXPR_0:
		a.analyseArithmeticExp0(child, 0)
	}
	a.infoFlag()
	a.analyseArithmeticExp(node, next+1)
}

// analyseItem 分析项
func (a *Analyser) analyseItem(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.FACTOR:
		a.analyseFactor(child, 0)
	case consts.TERM_0:
		a.analyseItem0(child, 0)
	}
	a.infoFlag()
	a.analyseItem(node, next+1)
}

// analyseFactor 分析因子
func (a *Analyser) analyseFactor(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "(":
		a.calStacks.PushOpe(consts.QUA_LEFTSMALLBRACKET)
	case ")":
		a.calStacks.PushOpe(consts.QUA_RIGHTSMALLBRACKET)
	case consts.BOOLEAN_EXPR:
		a.analyseBoolExp(child, 0)
	case consts.VARIABLE:
		if a.checkVar(child.Children[0]) {
			a.calStacks.PushNum(child.Children[0].Value)
		} else {
			a.err = true
		}
	case consts.CONSTANT:
		if a.checkConstNumber(child.Children[0].Children[0]) {
			a.calStacks.PushNum(child.Children[0].Children[0].Value)
		} else {
			a.err = true
		}
	case consts.FUNCTION_CALL:
		a.analyseFuncCall(child, 0, false)
	case consts.FACTOR_0:
		a.analyseFactor0(child, 0)
	}
	a.infoFlag()
	a.analyseFactor(node, next+1)
}

// analyseFactor0 分析因子0
func (a *Analyser) analyseFactor0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "+":
		a.calStacks.PushOpe(consts.QUA_POSITIVE)
	case "-":
		a.calStacks.PushOpe(consts.QUA_NEGATIVE)
	case "!":
		a.calStacks.PushOpe(consts.QUA_NOT)
	case consts.FACTOR:
		a.analyseFactor(child, 0)
	}
	a.infoFlag()
	a.analyseFactor0(node, next+1)
}

// analyseItem0 分析项0
func (a *Analyser) analyseItem0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "*":
		a.calStacks.PushOpe(consts.QUA_MUL)
	case "/":
		a.divFlag = true
		a.divToken = child.Token
		a.calStacks.PushOpe(consts.QUA_DIV)
	case "%":
		a.calStacks.PushOpe(consts.QUA_MOD)
	case consts.FACTOR:
		a.analyseFactor(child, 0)
		if a.divFlag {
			a.divFlag = false
			if a.calStacks.CurrentStack.NumStack.Top() == "0" || a.calStacks.CurrentStack.NumStack.Top() == "0.0" {
				a.Logger.AddAnalyseErr(a.divToken, "除数不能为0")
			} else if info, ok := a.SymbolTable.FindVariable(a.Scope, a.calStacks.CurrentStack.NumStack.Top().(string)); ok && info.Value == "0" {
				a.Logger.AddAnalyseErr(a.divToken, "除数不能为0")
			} else if info, ok = a.SymbolTable.FindConstant(a.Scope, a.calStacks.CurrentStack.NumStack.Top().(string)); ok && info.Value == "0" {
				a.Logger.AddAnalyseErr(a.divToken, "除数不能为0")
			}
		}
	case consts.TERM_0:
		a.analyseItem0(child, 0)
	}
	a.infoFlag()
	a.analyseItem0(node, next+1)
}

// analyseArithmeticExp0 分析算术表达式0
func (a *Analyser) analyseArithmeticExp0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "+":
		a.calStacks.PushOpe(consts.QUA_ADD)
	case "-":
		a.calStacks.PushOpe(consts.QUA_SUB)
	case consts.TERM:
		a.analyseItem(child, 0)
	case consts.ARITHMETIC_EXPR_0:
		a.analyseArithmeticExp0(child, 0)
	}
	a.infoFlag()
	a.analyseArithmeticExp0(node, next+1)
}

// analyseBoolFactor0 分析布尔因子0
func (a *Analyser) analyseBoolFactor0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.RELATION_OPERATOR:
		a.analyseRelationOperator(child, 0)
	case consts.ARITHMETIC_EXPR:
		a.analyseArithmeticExp(child, 0)
	}
	a.infoFlag()
	a.analyseBoolFactor0(node, next+1)
}

// analyseRelationOperator 分析关系运算符
func (a *Analyser) analyseRelationOperator(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.RELATION_OPERATOR:
		a.analyseRelationOperator(child, 0)
	case ">":
		a.calStacks.PushOpe(consts.QUA_GT)
	case ">=":
		a.calStacks.PushOpe(consts.QUA_GE)
	case "<":
		a.calStacks.PushOpe(consts.QUA_LT)
	case "<=":
		a.calStacks.PushOpe(consts.QUA_LE)
	case "==":
		a.calStacks.PushOpe(consts.QUA_EQ)
	case "!=":
		a.calStacks.PushOpe(consts.QUA_NE)
	}
	a.infoFlag()
	a.analyseRelationOperator(node, next+1)
}

// analyseBoolExp0 分析布尔表达式0
func (a *Analyser) analyseBoolExp0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "||":
		a.calStacks.PushOpe(consts.QUA_OR)
	case consts.BOOLEAN_ITEM:
		a.analyseBoolItem(child, 0)
	case consts.BOOLEAN_EXPR_0:
		a.analyseBoolExp0(child, 0)
	}
	a.infoFlag()
	a.analyseBoolExp0(node, next+1)
}

// analyseDeclarationFunctionStatement 分析函数声明语句
func (a *Analyser) analyseDeclarationFunctionStatement(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case ";":
		if !a.err {
			// 初始化函数作用域
			a.SymbolTable.VarTable[a.info.Name] = make(map[string]*Info)
			a.addFuncTable()
		}
		a.err = false
	case consts.FUNCTION_DECL:
		a.analyseDeclarationFunction(child, 0)
	}
	a.infoFlag()
	a.analyseDeclarationFunctionStatement(node, next+1)
}

// analyseDeclarationFunction 分析函数声明
func (a *Analyser) analyseDeclarationFunction(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.FUNCTION_TYPE:
		a.info.Type = child.Children[0].Value
	case consts.VARIABLE:
		a.info.Name = child.Children[0].Value
	case consts.FUNCTION_PARAMS:
		a.analyseDeclFormalParamList(child, 0)
	}
	a.infoFlag()
	a.analyseDeclarationFunction(node, next+1)
}

// analyseDeclFormalParamList 分析函数声明形参列表
func (a *Analyser) analyseDeclFormalParamList(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.FUNCTION_PARAM:
		a.analyseDeclFormalParam(child, 0)
	}
	a.infoFlag()
	a.analyseDeclFormalParamList(node, next+1)
}

// analyseDeclFormalParam 分析函数声明形参
func (a *Analyser) analyseDeclFormalParam(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.VARIABLE_TYPE:
		a.info.Pars = append(a.info.Pars, child.Children[0].Value)
		//a.info.Type = child.Children[0].Value
	case consts.FUNCTION_PARAM_0:
		a.analyseDeclFormalParam0(child, 0)
	}
	a.infoFlag()
	a.analyseDeclFormalParam(node, next+1)
}

// analyseDeclFormalParam0 分析函数声明形参0
func (a *Analyser) analyseDeclFormalParam0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.FUNCTION_PARAM:
		a.analyseDeclFormalParam(child, 0)
	}
	a.infoFlag()
	a.analyseDeclFormalParam0(node, next+1)
}

// analyseCompoundStatement 分析复合语句
func (a *Analyser) analyseCompoundStatement(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "{":
		a.Level++
		a.info.Level = a.Level
	case "}":
		a.Level--
		a.info.Level = a.Level
	case consts.STATEMENT_TABLE:
		a.analyseStatementTable(child, 0)
	}
	a.infoFlag()
	a.analyseCompoundStatement(node, next+1)
}

// analyseStatementTable 分析语句表
func (a *Analyser) analyseStatementTable(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.STATEMENT:
		a.analyseStatement(child, 0)
	case consts.STATEMENT_TABLE_0:
		a.analyseStatementTable0(child, 0)
	}
	a.infoFlag()
	a.analyseStatementTable(node, next+1)
}

// analyseStatementTable0 分析语句表0
func (a *Analyser) analyseStatementTable0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.STATEMENT_TABLE:
		a.analyseStatementTable(child, 0)
	}
	a.infoFlag()
	a.analyseStatementTable0(node, next+1)
}

// analyseStatement 分析语句
func (a *Analyser) analyseStatement(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.VALUE_DECLARATION:
		a.analyseDeclarationValue(child, 0)
	case consts.EXECUTION_STMT:
		a.analyseExeStatement(child, 0)
	}
	a.infoFlag()
	a.analyseStatement(node, next+1)
}

// analyseExeStatement 分析执行语句
func (a *Analyser) analyseExeStatement(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.DATA_PROCESS_STMT:
		a.analyseDataHandleStatement(child, 0)
	case consts.CONTROL_STMT:
		a.analyseControlStatement(child, 0)
	case consts.COMPOUND_STMT:
		a.analyseCompoundStatement(child, 0)
	}
	a.infoFlag()
	a.analyseExeStatement(node, next+1)
}

// analyseDataHandleStatement 分析数据处理语句
func (a *Analyser) analyseDataHandleStatement(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.ASSIGNMENT_STMT:
		a.analyseAssignmentStatement(child, 0)
	case consts.FUNCTION_CALL_STMT:
		a.analyseFuncCallStatement(child, 0)
	}
	a.infoFlag()
	a.analyseDataHandleStatement(node, next+1)
}

// analyseControlStatement 分析控制语句
func (a *Analyser) analyseControlStatement(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.IF_STMT:
		stack := util.NewStack()
		a.calStacks.PushIfStack(stack)
		a.analyseIfStatement(child, 0)
		a.calStacks.PopCurrentIfStack()
	case consts.WHILE_STMT:
		jmpPos := util.NewForJmpPos()
		a.Qf.JmpPoint.Push(jmpPos)
		a.CurrentJmpPos = jmpPos

		breakStack := util.NewStack()
		a.Qf.PushBreakStack(breakStack)
		continueStack := util.NewStack()
		a.Qf.PushContinue(continueStack)

		a.analyseWhileStatement(child, 0)

		//回填continue出口
		a.Qf.ClearContinueStack(a.CurrentJmpPos.ContinuePos)

		a.Qf.JmpPoint.Pop()
		if !a.Qf.JmpPoint.IsEmpty() {
			a.CurrentJmpPos = a.Qf.JmpPoint.Top().(*util.ForJmpPos)
		}
		//回填break出口
		a.Qf.ClearBreakStack(a.Qf.NextQuaFormId())
	case consts.DO_WHILE_STMT:
		jmpPos := util.NewForJmpPos()
		a.Qf.JmpPoint.Push(jmpPos)
		a.CurrentJmpPos = jmpPos

		breakStack := util.NewStack()
		a.Qf.PushBreakStack(breakStack)
		continueStack := util.NewStack()
		a.Qf.PushContinue(continueStack)

		a.analyseDoWhileStatement(child, 0)

		//回填continue出口
		a.Qf.ClearContinueStack(a.CurrentJmpPos.ContinuePos)

		a.Qf.JmpPoint.Pop()
		if !a.Qf.JmpPoint.IsEmpty() {
			a.CurrentJmpPos = a.Qf.JmpPoint.Top().(*util.ForJmpPos)
		}
		//回填break出口
		a.Qf.ClearBreakStack(a.Qf.NextQuaFormId())
	case consts.FOR_STMT:
		jmpPos := util.NewForJmpPos()
		a.Qf.JmpPoint.Push(jmpPos)
		a.CurrentJmpPos = jmpPos

		breakStack := util.NewStack()
		a.Qf.PushBreakStack(breakStack)
		continueStack := util.NewStack()
		a.Qf.PushContinue(continueStack)

		a.analyseForStatement(child, 0)

		//回填continue出口
		a.Qf.ClearContinueStack(a.CurrentJmpPos.ContinuePos)

		a.Qf.JmpPoint.Pop()
		if !a.Qf.JmpPoint.IsEmpty() {
			a.CurrentJmpPos = a.Qf.JmpPoint.Top().(*util.ForJmpPos)
		}
		//回填break出口
		a.Qf.ClearBreakStack(a.Qf.NextQuaFormId())
	case consts.RETURN_STMT:
		a.analyseReturn(child, 0)
	case consts.BREAK_STMT:
		a.analyseBreak(child, 0)
	case consts.CONTINUE_STMT:
		a.analyseContinue(child, 0)
	}
	a.infoFlag()
	a.analyseControlStatement(node, next+1)
}

// analyseIfStatement 分析if语句
func (a *Analyser) analyseIfStatement(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "if":
		a.info.Type = "int"
	case "(":
		a.Qf.IfFlag = true
		a.calStacks.PushOpe(consts.QUA_LEFTSMALLBRACKET)
	case ")":
		a.calStacks.PushOpe(consts.QUA_RIGHTSMALLBRACKET)
		//a.calStacks.CalIf() //执行一次move操作，将括号算出的逻辑栈值移动到当前逻辑栈
		a.calStacks.CurrentStack.OpStack.Pop() // 弹出move操作
		//分析完if的判断条件后，需要回填真出口
		a.calStacks.ClearTrueStack(a.Qf.NextQuaFormId())
		a.Qf.IfFlag = false
	case consts.BOOLEAN_EXPR:
		a.analyseBoolExp(child, 0)
	case consts.COMPOUND_STMT:
		a.analyseCompoundStatement(child, 0)
	case consts.IF_TAIL:
		//如果ifTail为空，说明整个if语句结束，需要清空栈
		if child.Children[0].Value == consts.NULL {
			a.calStacks.ClearTrueStack(a.Qf.NextQuaFormId())
			a.calStacks.ClearFalseStack(a.Qf.NextQuaFormId())
			a.calStacks.ClearCurrentIfStack()
			a.calStacks.PopCurrentLogicStack()
		} else { //如果ifTail不为空，说明还有else语句，需要继续分析。回填假出口栈
			//if语句结束，跳出整个if语句
			id := a.Qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, nil)
			a.calStacks.CurrentIfQuaStack.Push(id)
			a.calStacks.ClearFalseStack(a.Qf.NextQuaFormId())
			a.calStacks.PopCurrentLogicStack()
		}
		a.analyseIfTail(child, 0)
	}
	a.infoFlag()
	a.analyseIfStatement(node, next+1)
}

// analyseIfTail 分析ifTail语句
func (a *Analyser) analyseIfTail(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "else":
	case consts.IF_TAIL_0:
		a.analyseIfTail0(child, 0)
	}
	a.infoFlag()
	a.analyseIfTail(node, next+1)
}

// analyseIfTail0 分析ifTail0语句
func (a *Analyser) analyseIfTail0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.IF_STMT:
		a.analyseIfStatement(child, 0)
	case consts.COMPOUND_STMT:
		a.info.Type = "int"
		stack := util.NewLogicStack(a.Qf)
		a.calStacks.PushLogicStack(stack)

		a.analyseCompoundStatement(child, 0)

		//else后紧跟的是复合语句，说明整个if语句结束，需要清空栈
		a.calStacks.ClearCurrentIfStack()
		a.calStacks.ClearTrueStack(a.Qf.NextQuaFormId())
		a.calStacks.ClearFalseStack(a.Qf.NextQuaFormId())
		a.calStacks.PopCurrentLogicStack()
	}
	a.infoFlag()
	a.analyseIfTail0(node, next+1)
}

// analyseWhileStatement 分析while语句
func (a *Analyser) analyseWhileStatement(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "while":
		a.info.Type = "int"
		//记录continue出口的位置
		a.CurrentJmpPos.ContinuePos = a.Qf.NextQuaFormId()
	case "(":
		a.Qf.IfFlag = true
		a.CurrentJmpPos.ConditionPos = a.Qf.NextQuaFormId()
		a.calStacks.PushOpe(consts.QUA_LEFTSMALLBRACKET)
	case ")":
		a.calStacks.PushOpe(consts.QUA_RIGHTSMALLBRACKET)
		a.calStacks.CalIf() //执行一次move操作，将括号算出的逻辑栈值移动到当前逻辑栈
		//分析完while的判断条件后，需要回填真出口
		a.calStacks.ClearTrueStack(a.Qf.NextQuaFormId())
		a.Qf.IfFlag = false
	case consts.BOOLEAN_EXPR:
		a.analyseBoolExp(child, 0)
	case consts.COMPOUND_STMT:
		a.analyseCompoundStatement(child, 0)
		//while语句结束，跳回到while的判断条件，然后回填假出口
		a.Qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, a.CurrentJmpPos.ConditionPos)
		a.calStacks.ClearTrueStack(a.Qf.NextQuaFormId())
		a.calStacks.ClearFalseStack(a.Qf.NextQuaFormId())
		a.calStacks.PopCurrentLogicStack()
	}
	a.infoFlag()
	a.analyseWhileStatement(node, next+1)
}

// analyseDoWhileStatement 分析do while语句
func (a *Analyser) analyseDoWhileStatement(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "do":
	case "while":
		a.info.Type = "int"
		//记录continue出口的位置
		a.CurrentJmpPos.ContinuePos = a.Qf.NextQuaFormId()
	case "(":
		a.Qf.IfFlag = true
		a.calStacks.PushOpe(consts.QUA_LEFTSMALLBRACKET)
	case ")":
		a.calStacks.PushOpe(consts.QUA_RIGHTSMALLBRACKET)
		a.calStacks.CalIf() //执行一次move操作，将括号算出的逻辑栈值移动到当前逻辑栈
		//在do while语句中，条件判断结束后，真出口跳转到语句开始位置，假出口跳转到下一条指令
		a.calStacks.ClearTrueStack(a.CurrentJmpPos.ConditionPos)
		a.calStacks.ClearFalseStack(a.Qf.NextQuaFormId())
		a.Qf.IfFlag = false
	case ";":
		a.flag = true
		a.calStacks.PopCurrentLogicStack()
	case consts.BOOLEAN_EXPR:
		a.analyseBoolExp(child, 0)
	case consts.COMPOUND_STMT:
		//记录语句开始的位置
		a.CurrentJmpPos.ConditionPos = a.Qf.NextQuaFormId()
		a.analyseCompoundStatement(child, 0)
	}
	a.infoFlag()
	a.analyseDoWhileStatement(node, next+1)
}

// analyseForStatement 分析for语句
func (a *Analyser) analyseForStatement(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "for":
	case "(":
	case ")":
	case ";":

	case consts.ASSIGNMENT_EXPR:
		a.analyseAssignmentExp(child, 0)
		if !a.err {
			a.clearCalStacks()
			//a.changeVarTable()
		}
		a.calStacks.Clear()
		a.flag = true
		a.err = false
		//for语句中的第一个赋值表达式，只执行一次
		if node.Children[next+1].Value == ";" {
			a.Qf.IfFlag = true
			//记录判断条件的位置
			a.CurrentJmpPos.ConditionPos = a.Qf.NextQuaFormId()
			a.calStacks.PushOpe(consts.QUA_LEFTSMALLBRACKET)
		} else {
			//每次循环结束后，先执行赋值表达式，再跳转到for语句中的条件判断
			a.Qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, a.CurrentJmpPos.ConditionPos)
		}
	case consts.BOOLEAN_EXPR:
		a.analyseBoolExp(child, 0)
		a.calStacks.PushOpe(consts.QUA_RIGHTSMALLBRACKET)
		a.calStacks.CalIf() //执行一次move操作，将括号算出的逻辑栈值移动到当前逻辑栈
		a.Qf.IfFlag = false
		//记录每次循环后需要执行的赋值表达式的位置
		a.CurrentJmpPos.AssignPos = a.Qf.NextQuaFormId()
		//记录continue出口的位置
		a.CurrentJmpPos.ContinuePos = a.Qf.NextQuaFormId()
	case consts.COMPOUND_STMT:
		//复合语句中的第一条语句即为真出口，所以在此处要回填真出口
		a.calStacks.ClearTrueStack(a.Qf.NextQuaFormId())
		a.analyseCompoundStatement(child, 0)
		//for语句的复合语句结束后，先跳转到for语句中的第二个赋值表达式，再跳转到for的判断条件，最后回填假出口
		a.Qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, a.CurrentJmpPos.AssignPos)
		a.calStacks.ClearFalseStack(a.Qf.NextQuaFormId())
		a.calStacks.PopCurrentLogicStack()
	}
	a.infoFlag()
	a.analyseForStatement(node, next+1)
}

// analyseBreak 分析break语句
func (a *Analyser) analyseBreak(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "break":
		//break跳转的位置是需要回填的
		id := a.Qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, nil)
		a.Qf.CurrentBreakStack.Push(id)
	}
	a.infoFlag()
	a.analyseBreak(node, next+1)
}

// analyseContinue 分析continue语句
func (a *Analyser) analyseContinue(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "continue":
		//continue跳转的位置是需要回填的
		id := a.Qf.AddQuaForm(consts.QuaFormMap[consts.QUA_JMP], nil, nil, nil)
		a.Qf.CurrentContinueStack.Push(id)
	}
	a.infoFlag()
	a.analyseBreak(node, next+1)
}

// analyseReturn 分析return语句
func (a *Analyser) analyseReturn(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case "return":
		if a.Level == 1 {
			a.retFlag = true
		}
	case consts.RETURN_STMT_0:
		a.analyseReturn0(child, 0)
	}
	a.infoFlag()
	a.analyseReturn(node, next+1)
}

// analyseReturn0 分析return0语句
func (a *Analyser) analyseReturn0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case ";":
		if next == 0 {
			a.Qf.AddQuaForm(consts.QuaFormMap[consts.QUA_RETURN], nil, nil, nil)
		} else {
			a.calStacks.CalAllUtilReturn()
			a.Qf.AddQuaForm(consts.QuaFormMap[consts.QUA_RETURN], nil, nil, a.calStacks.Result)
		}
	case consts.BOOLEAN_EXPR:
		a.calStacks.CurrentStack.OpStack.Push(consts.QUA_RETURN)
		a.analyseBoolExp(child, 0)
	}
	a.infoFlag()
	a.analyseReturn0(node, next+1)
}

// analyseAssignmentStatement 分析赋值语句
func (a *Analyser) analyseAssignmentStatement(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case ";":
		if !a.err {
			a.clearCalStacks()
			//a.changeVarTable()
		}
		a.calStacks.Clear()
		a.flag = true
		a.err = false
	case consts.ASSIGNMENT_EXPR:
		a.analyseAssignmentExp(child, 0)
	}
	a.infoFlag()
	a.analyseAssignmentStatement(node, next+1)
}

// analyseFuncCallStatement 分析函数调用语句
func (a *Analyser) analyseFuncCallStatement(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case ";":
		a.flag = true
		a.err = false
	case consts.FUNCTION_CALL:
		a.analyseFuncCall(child, 0, true)
	}
	a.infoFlag()
	a.analyseFuncCallStatement(node, next+1)
}

// analyseFuncCall 分析函数调用,flag用于标记是在函数调用语句中还是在表达式中，如果是一个<函数调用语句>只需要判断函数是否存在，参数个数是否匹配，参数类型是否匹配,如果是在<布尔表达式>中还需要判断返回类型是否匹配
func (a *Analyser) analyseFuncCall(node *util.TreeNode, next int, flag bool) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.VARIABLE:
		if flag { //在函数调用语句中
			if !a.funcIsExist(child.Children[0].Value) {
				a.Logger.AddAnalyseErr(child.Children[0].Token, "函数未定义")
				a.err = true
			} else {
				a.info.Name = child.Children[0].Value
				a.calStacks.PushFuncCall(child.Children[0].Value)
			}
		} else { //在布尔表达式中
			a.Qf.FuncCall++
			if a.checkFunc(child.Children[0]) {
				a.calStacks.PushFuncCall(child.Children[0].Value)
			} else {
				a.err = true
			}
		}
	case "(":

	case consts.ARGUMENTS:
		if child.Children[0].Value != consts.NULL {
			a.calStacks.PushOpe(consts.QUA_PARAM)
		}
		a.analyseActualParamList(child, 0)
	case ")":
		//如果是函数调用语句，最后需要清空栈；如果是在布尔表达式中，要对栈进行计算直到函数调用符号出栈
		if flag {
			a.clearCalStacks()
		} else {
			a.calStacks.CalAllUtilCall()
		}

	}
	a.infoFlag()
	a.analyseFuncCall(node, next+1, flag)
}

// analyseActualParamList 分析实参列表
func (a *Analyser) analyseActualParamList(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.ARGUMENT:
		a.analyseActualParam(child, 0)
	}
	a.infoFlag()
	a.analyseActualParamList(node, next+1)
}

// analyseActualParam 分析实参
func (a *Analyser) analyseActualParam(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.BOOLEAN_EXPR:
		a.analyseBoolExp(child, 0)
	case consts.ARGUMENT_0:
		a.analyseActualParam0(child, 0)
	}
	a.infoFlag()
	a.analyseActualParam(node, next+1)
}

// analyseActualParam0 分析实参0
func (a *Analyser) analyseActualParam0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case ",":
		a.calStacks.PushOpe(consts.QUA_PARAM)
	case consts.ARGUMENT:
		a.analyseActualParam(child, 0)
	}
	a.infoFlag()
	a.analyseActualParam0(node, next+1)
}

// analyseAssignmentExp 分析赋值表达式
func (a *Analyser) analyseAssignmentExp(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.VARIABLE:
		if a.constIsExist(child.Children[0].Value) {
			a.info.Name = child.Children[0].Value
			a.Logger.AddAnalyseErr(child.Children[0].Token, "常量不可赋值")
			a.err = true
		} else if !a.varIsExist(child.Children[0].Value) {
			a.Logger.AddAnalyseErr(child.Children[0].Token, "变量未定义")
			a.err = true
		} else {
			if a.checkVar(child.Children[0]) {
				a.info.Name = child.Children[0].Value
				a.calStacks.PushNum(child.Children[0].Value)
			} else {
				a.err = true
			}

		}
	case "=":
		a.calStacks.PushOpe(consts.QUA_ASSIGNMENT)
	case consts.BOOLEAN_EXPR:
		a.analyseBoolExp(child, 0)
	}
	a.infoFlag()
	a.analyseAssignmentExp(node, next+1)
}

// analyseFunctionBlock 分析函数块	TODO:return语句的处理?
func (a *Analyser) analyseFunctionBlock(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.FUNCTION_DEF:
		a.analyseFunctionDefine(child, 0)
	case consts.FUNCTION_BLOCK:
		a.analyseFunctionBlock(child, 0)
	}
	a.infoFlag()
	a.analyseFunctionBlock(node, next+1)
}

// analyseFunctionDefine 分析函数定义
func (a *Analyser) analyseFunctionDefine(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.FUNCTION_TYPE:
		a.info.Type = child.Children[0].Value
	case consts.VARIABLE:
		if !a.funcIsExist(child.Children[0].Value) {
			a.Logger.AddAnalyseErr(child.Children[0].Token, "函数未声明")
			a.err = true
		} else {
			a.info.Name = child.Children[0].Value
			a.currentFunc = a.info.Name
			a.Qf.AddQuaForm(a.info.Name, nil, nil, nil)
			info, _ := a.SymbolTable.FindFunction(a.info.Name)
			a.Scope = a.info.Name
			a.info.Scope = a.Scope
			if info.Type != a.info.Type {
				a.Logger.AddAnalyseErr(child.Children[0].Token, "函数返回类型不匹配")
				a.err = true
			}
		}
	case "(":
		a.params = make([]Param, 0)
	case ")":
		if a.paramFlag {
			a.params = append(a.params, Param{
				Name: a.info.Name,
				Type: a.info.Type,
			})
		}
		a.checkFuncParamList(a.currentFunc) //检查函数参数类型是否匹配，并加入符号表
	case consts.FUNCTION_PARAMS_DEF:
		//参数为空
		if child.Children[0].Value == consts.NULL {
			a.paramFlag = false
		} else {
			a.paramFlag = true
		}
		a.analyseDefineFormalParamList(child, 0)
	case consts.COMPOUND_STMT:
		a.analyseCompoundStatement(child, 0)
		if !a.retFlag {
			f, _ := a.SymbolTable.FindFunction(a.currentFunc)
			if f.Type == "void" {
				a.Qf.AddQuaForm(consts.QuaFormMap[consts.QUA_RETURN], nil, nil, nil)
			} else {
				a.Logger.AddErr("\t\t\t\t\t\t函数：" + a.currentFunc + " 缺少返回语句\n")
				a.err = true
			}
		}
		a.retFlag = false
	}
	a.infoFlag()
	a.analyseFunctionDefine(node, next+1)
}

// analyseDefineFormalParamList 分析函数定义形参列表
func (a *Analyser) analyseDefineFormalParamList(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.FUNCTION_PARAM_DEF:
		a.analyseDefineFormalParam(child, 0)

	}
	a.infoFlag()
	a.analyseDefineFormalParamList(node, next+1)
}

// analyseDefineFormalParam 分析函数定义形参
func (a *Analyser) analyseDefineFormalParam(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case consts.VARIABLE_TYPE:
		a.info.Type = child.Children[0].Value
	case consts.VARIABLE:
		a.info.Name = child.Children[0].Value
	case consts.FUNCTION_PARAM_0_DEF:
		a.analyseDefineFormalParam0(child, 0)
	}
	a.infoFlag()
	a.analyseDefineFormalParam(node, next+1)
}

// analyseDefineFormalParam0 分析函数定义形参0
func (a *Analyser) analyseDefineFormalParam0(node *util.TreeNode, next int) {
	if next >= len(node.Children) || !isLegalNode(node) {
		return
	}
	if a.info == nil {
		a.initInfo()
	}
	child := node.Children[next]
	switch child.Value {
	case ",":
		a.params = append(a.params, Param{
			Name: a.info.Name,
			Type: a.info.Type,
		})
	case consts.FUNCTION_PARAM_DEF:
		a.analyseDefineFormalParam(child, 0)

	}
	a.infoFlag()
	a.analyseDefineFormalParam0(node, next+1)
}
