package stateful

import (
	"errors"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

type Stateful interface {
	SaveState() error
	LoadState() error
	Close() error
	TGAdminPrefix() string
	HandleTGCommands(*zap.Logger, *telego.Bot, *telego.Message, []string) error
}

var (
	ErrNotSupported   = errors.New("not supported")
	ErrUserIDInvalid  = errors.New("invalid user id")
	ErrInvalidCommand = errors.New("invalid command")
)
