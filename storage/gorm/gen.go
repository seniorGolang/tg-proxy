package gorm

import (
	"gorm.io/cli/gorm/genconfig"
)

//go:generate go run gorm.io/cli/gorm@latest gen -i . -o ./generated

var _ = genconfig.Config{
	OutPath:        "./generated",
	IncludeStructs: []any{Project{}},
}
