package report

import (
	"errors"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	actions "github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/bannedDB"
	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/statefulFilters/state"
	badgerHelper "github.com/Civil/tg-simple-regex-antispam/helper/badger"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
)

type Filter struct {
	chainName string
	logger    *zap.Logger

	stateDir string

	filteringRules []interfaces.FilteringRule
	actions        []actions.Action

	db  *badger.DB
	bot *telego.Bot

	isFinal bool
}

var (
	ErrStateDirEmpty = errors.New("state_dir cannot be empty")
	ErrNIsZero       = errors.New("n cannot be equal to 0")
)

func New(logger *zap.Logger, chainName string, _ bannedDB.BanDB, bot *telego.Bot, config map[string]any,
	filteringRules []interfaces.FilteringRule, actions []actions.Action,
) (interfaces.StatefulFilter, error) {
	var stateDir string
	var err error
	stateDir, err = config2.GetOptionString(config, "state_dir")
	if err != nil {
		return nil, err
	}
	if stateDir == "" {
		return nil, ErrStateDirEmpty
	}

	isFinal, err := config2.GetOptionBoolWithDefault(config, "isFinal", false)
	if err != nil {
		return nil, err
	}

	badgerDB, err := badger.Open(badger.DefaultOptions(stateDir))
	if err != nil {
		return nil, err
	}

	f := &Filter{
		logger: logger.With(
			zap.String("filter", chainName),
			zap.String("filter_type", "report"),
		),
		chainName:      chainName,
		stateDir:       stateDir,
		db:             badgerDB,
		bot:            bot,
		isFinal:        isFinal,
		filteringRules: filteringRules,
		actions:        actions,
	}
	return f, nil
}

func (r *Filter) setState(userID int64, s *state.State) error {
	b, err := proto.Marshal(s)
	if err != nil {
		return err
	}
	return r.db.Update(
		func(txn *badger.Txn) error {
			return txn.Set(badgerHelper.UserIDToKey(userID), b)
		})
}

func (r *Filter) getState(userID int64) (*state.State, error) {
	var s state.State
	err := r.db.View(
		func(txn *badger.Txn) error {
			item, err := txn.Get(badgerHelper.UserIDToKey(userID))
			if err != nil {
				return err
			}
			return item.Value(func(val []byte) error {
				return proto.Unmarshal(val, &s)
			})
		})
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Filter) RemoveState(userID int64) error {
	return r.db.Update(
		func(txn *badger.Txn) error {
			return txn.Delete(badgerHelper.UserIDToKey(userID))
		})
}

func (r *Filter) Score(msg *telego.Message) int {
	if !strings.HasPrefix("/report", msg.Text) && !strings.HasPrefix("/spam", msg.Text) {
		return 0
	}
	if msg.ReplyToMessage == nil {
		r.logger.Debug("message does not have a reply")
		return 0
	}
	reportedMsg := msg.ReplyToMessage
	stateKey := int64(reportedMsg.MessageID)
	actualState, err := r.getState(stateKey)
	if err != nil || actualState == nil || len(actualState.MessageIds) == 0 {
		r.logger.Debug("failed to get state, creating a clean one",
			zap.Int("messageID", reportedMsg.MessageID),
			zap.Error(err),
		)
		actualState = &state.State{
			Verified:   false,
			MessageIds: []int64{int64(reportedMsg.MessageID)},
			LastUpdate: timestamppb.Now(),
		}
	}

	// We already reported that message/user
	if actualState.Verified {
		r.logger.Debug("message/user already reported")
		sendMessageParams := &telego.SendMessageParams{
			ChatID: msg.Chat.ChatID(),
			Text:   "Message/user already reported",
			ReplyParameters: &telego.ReplyParameters{
				MessageID: msg.MessageID,
			},
		}

		_, err = r.bot.SendMessage(
			sendMessageParams,
		)
		if err != nil {
			r.logger.Error("failed to send message", zap.Error(err))
		}
		return -1
	}

	r.logger.Debug("applying actions...")
	for _, action := range r.actions {
		r.logger.Debug("trying to apply action",
			zap.Any("message_ids", actualState.MessageIds),
			zap.Any("action", action),
		)
		err = action.ApplyToMessage(r, reportedMsg)
		if err != nil {
			r.logger.Error("failed to apply action", zap.Any("action", action), zap.Error(err))
			return 100
		}
	}

	actualState.Verified = true
	err = r.setState(stateKey, actualState)
	if err != nil {
		r.logger.Error("failed to set new state",
			zap.Any("new_state", actualState),
			zap.Error(err),
		)
	}

	return 100
}

func (r *Filter) IsStateful() bool {
	return true
}

func (r *Filter) GetName() string {
	return "report"
}

func (r *Filter) GetFilterName() string {
	return r.chainName
}

func (r *Filter) IsFinal() bool {
	return r.isFinal
}

func Help() string {
	return "report requires `stateFile` parameter"
}

func (r *Filter) Close() error {
	return r.db.Close()
}

func (r *Filter) SaveState() error {
	return nil
}

func (r *Filter) LoadState() error {
	return nil
}
