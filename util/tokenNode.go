package util

import (
	"complier/pkg/consts"
)

// Position 当前读到的行列
type Position struct {
	Line   int
	Column int
}

// TokenNode token值和种别码
type TokenNode struct {
	Pos   Position
	Type  consts.Token
	Value string
}
