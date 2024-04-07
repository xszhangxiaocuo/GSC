package compiler

import "complier/pkg/consts"

// TokenNode token值和种别码
type TokenNode struct {
	Pos   Position
	Type  consts.Token
	Value string
}
