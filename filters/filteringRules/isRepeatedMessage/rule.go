package isRepeatedMessage

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/types/messageWithCounter"
	"github.com/Civil/tg-simple-regex-antispam/filters/types/scoringResult"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
	"github.com/Civil/tg-simple-regex-antispam/sharedDBs/messages"
)

type Filter struct {
	messageDB *messages.MessageDB
	logger    *zap.Logger
	chainName string
	isFinal   bool
	threshold uint64
}

func New(logger *zap.Logger, config map[string]any, chainName string) (interfaces.FilteringRule, error) {
	logger = logger.With(zap.String("filter", chainName), zap.String("filter_type", "isRepeatedMessage"))
	isFinal, err := config2.GetOptionBoolWithDefault(config, "isFinal", true)
	if err != nil {
		return nil, err
	}

	threshold, err := config2.GetOptionIntWithDefault(config, "n", 1)
	if err != nil {
		return nil, err
	}

	if threshold < 1 {
		return nil, errors.New("threshold should be greater than 1")
	}

	messageDB, err := messages.Get(logger, "repeatedMessages")
	if err != nil {
		return nil, err
	}

	return &Filter{
		logger:    logger,
		chainName: chainName,
		isFinal:   isFinal,
		messageDB: messageDB,

		threshold: uint64(threshold),
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
	msgSha256 := sha256.Sum256(msgByte)
	key := messageWithCounter.Key_builder{
		UserId: msg.From.ID,
		Sha256: msgSha256[:],
	}.Build()
	marshaledKey, err := proto.Marshal(key)
	if err != nil {
		return res
	}
	data, err := r.messageDB.LoadValue(marshaledKey)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			msgWithCounter := messageWithCounter.MessageWithCounter_builder{
				Message: msg.Text,
				Counter: 1,
			}.Build()
			data, err = proto.Marshal(msgWithCounter)
			if err != nil {
				return res
			}
			err = r.messageDB.StoreValue(marshaledKey, data)
			if err != nil {
				return res
			}
		}
		return res
	}

	msgWithCounter := &messageWithCounter.MessageWithCounter{}
	err = proto.Unmarshal(data, msgWithCounter)
	if err != nil {
		return res
	}

	if strings.Compare(msgWithCounter.GetMessage(), msg.Text) == 0 {
		counter := msgWithCounter.GetCounter() + 1
		msgWithCounter.SetCounter(counter)
		if counter > r.threshold {
			res.Score = 100
			res.Reason = fmt.Sprintf("this message was repeated by that user more than %v times",
				r.threshold)
		}
		data, err = proto.Marshal(msgWithCounter)
		if err != nil {
			return res
		}
		err = r.messageDB.StoreValue(marshaledKey, data)
		if err != nil {
			return res
		}
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
	return "isRepeatedMessage"
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
