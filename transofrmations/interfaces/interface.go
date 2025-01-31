package interfaces

import (
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

type TransformationRule interface {
	Transform(string) string
	IsStateful() bool
	GetName() string
	GetTransformationName() string
	TGAdminPrefix() string
	HandleTGCommands(*zap.Logger, *telego.Bot, *telego.Message, []string) error
}
