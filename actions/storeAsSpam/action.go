package storeAsSpam

import (
	"crypto/sha256"
	"errors"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	interfaces2 "github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/types/scoringResult"
	"github.com/Civil/tg-simple-regex-antispam/sharedDBs/messages"
)

type Action struct {
	logger *zap.Logger
	bot    *telego.Bot

	messageDB *messages.MessageDB
}

func (r *Action) Apply(_ interfaces2.StatefulFilter, _ *scoringResult.ScoringResult, _ telego.ChatID, _ []int64,
	_ int64, _ any) error {
	return ErrNotSupported
}

var ErrNotSupported = errors.New("not supported")

func (r *Action) GetName() string {
	return "storeAsSpam"
}

func (r *Action) PerMessage() bool {
	return true
}

func (r *Action) ApplyToMessage(_ interfaces2.StatefulFilter, _ *scoringResult.ScoringResult,
	msg *telego.Message, _ any) error {

	r.messageDB.Lock()
	defer r.messageDB.Unlock()

	data := []byte(msg.Text)
	key := sha256.Sum256(data)

	err := r.messageDB.StoreValue(key[:], data)
	if err != nil {
		return err
	}

	return nil
}

func New(logger *zap.Logger, bot *telego.Bot, _ map[string]any) (interfaces.Action, error) {
	messageDB, err := messages.Get(logger, "spamMessages")
	if err != nil {
		return nil, err
	}

	return &Action{
		logger:    logger,
		bot:       bot,
		messageDB: messageDB,
	}, nil
}

func Help() string {
	return "storeAsSpam doesn't require any parameter"
}
