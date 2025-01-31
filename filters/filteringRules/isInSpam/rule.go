package isInSpam

import (
	"bytes"
	"crypto/sha256"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/types/scoringResult"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
	"github.com/Civil/tg-simple-regex-antispam/sharedDBs/messages"
)

type Filter struct {
	messageDB *messages.MessageDB
	logger    *zap.Logger
	chainName string
	isFinal   bool
}

func New(logger *zap.Logger, config map[string]any, chainName string) (interfaces.FilteringRule, error) {
	logger = logger.With(zap.String("filter", chainName), zap.String("filter_type", "isInSpam"))
	isFinal, err := config2.GetOptionBoolWithDefault(config, "isFinal", true)
	if err != nil {
		return nil, err
	}

	messageDB, err := messages.Get(logger, "spamMessages")
	if err != nil {
		return nil, err
	}

	return &Filter{
		logger:    logger,
		chainName: chainName,
		isFinal:   isFinal,
		messageDB: messageDB,
	}, nil
}

func Help() string {
	return "isInSpam checks if the message is forwarded"
}

func (r *Filter) Score(_ *telego.Bot, msg *telego.Message) *scoringResult.ScoringResult {
	res := &scoringResult.ScoringResult{
		Score: 0,
	}
	msgByte := []byte(msg.Text)
	key := sha256.Sum256(msgByte)
	data, err := r.messageDB.LoadValue(key[:])
	if err != nil {
		return res
	}
	if bytes.Equal(data, msgByte) {
		res.Reason = "this message matches one that was marked as spam before"
		res.Score = 100
	} else {
		r.logger.Error("key collision found",
			zap.ByteString("current_message", msgByte),
			zap.ByteString("collided_with", data),
		)
	}
	return res
}

func (r *Filter) IsStateful() bool {
	return false
}

func (r *Filter) GetName() string {
	return "isInSpam"
}

func (r *Filter) GetFilterName() string {
	return ""
}

func (r *Filter) IsFinal() bool {
	return r.isFinal
}

func (r *Filter) TGAdminPrefix() string {
	return ""
}

func (r *Filter) HandleTGCommands(_ *zap.Logger, _ *telego.Bot, _ *telego.Message, _ []string) error {
	return nil
}
