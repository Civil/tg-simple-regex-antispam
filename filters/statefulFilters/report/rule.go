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
	"github.com/Civil/tg-simple-regex-antispam/filters/types/checkNeventsState"
	badgerHelper "github.com/Civil/tg-simple-regex-antispam/helper/badger"
	"github.com/Civil/tg-simple-regex-antispam/helper/badger/badgerOpts"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
	"github.com/Civil/tg-simple-regex-antispam/helper/tg"
)

type Filter struct {
	chainName string
	logger    *zap.Logger

	stateDir string

	filteringRules []interfaces.FilteringRule
	actions        []actions.Action

	db  *badger.DB
	bot *telego.Bot

	isFinal         bool
	removeReportMsg bool
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

	removeReportMsg, err := config2.GetOptionBoolWithDefault(config, "removeReportMsg", true)
	if err != nil {
		return nil, err
	}

	badgerDB, err := badger.Open(badgerOpts.GetBadgerOptions(logger, chainName+"_DB", stateDir))
	if err != nil {
		return nil, err
	}

	f := &Filter{
		logger: logger.With(
			zap.String("filter", chainName),
			zap.String("filter_type", "report"),
		),
		chainName:       chainName,
		stateDir:        stateDir,
		db:              badgerDB,
		bot:             bot,
		isFinal:         isFinal,
		filteringRules:  filteringRules,
		removeReportMsg: removeReportMsg,
		actions:         actions,
	}
	return f, nil
}

func (r *Filter) setState(userID int64, s *checkNeventsState.State) error {
	b, err := proto.Marshal(s)
	if err != nil {
		return err
	}
	return r.db.Update(
		func(txn *badger.Txn) error {
			return txn.Set(badgerHelper.UserIDToKey(userID), b)
		})
}

func (r *Filter) getState(userID int64) (*checkNeventsState.State, error) {
	var s checkNeventsState.State
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
		r.logger.Debug("message does not start with /report or /spam")
		return 0
	}
	if msg.ReplyToMessage == nil {
		r.logger.Debug("message does not have a reply")
		return 0
	}
	r.logger.Debug("got a message that start with /report or /spam",
		zap.String("message_text", msg.Text),
		zap.String("from_user", msg.From.Username),
	)
	reportedMsg := msg.ReplyToMessage
	stateKey := int64(reportedMsg.MessageID)
	actualState, err := r.getState(stateKey)
	if err != nil || actualState == nil || len(actualState.MessageIds) == 0 {
		r.logger.Debug("failed to get state, creating a clean one",
			zap.Int("messageID", reportedMsg.MessageID),
			zap.Error(err),
		)
		actualState = &checkNeventsState.State{
			Verified:   false,
			MessageIds: make(map[int64]bool),
			LastUpdate: timestamppb.Now(),
		}
		actualState.MessageIds[stateKey] = true
	}

	// We already reported that message/user
	if actualState.Verified {
		r.logger.Debug("message/user already reported")
		sendMessageParams := &telego.SendMessageParams{
			ChatID: msg.Chat.ChatID(),
			Text:   "Message/user was already reported",
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

		if r.removeReportMsg {
			err = tg.DeleteMessage(r.bot, msg)
			if err != nil {
				r.logger.Error("failed to delete message", zap.Error(err))
			}
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

func (r *Filter) TGAdminPrefix() string {
	return ""
}

func (r *Filter) HandleTGCommands(logger *zap.Logger, bot *telego.Bot, message *telego.Message, tokens []string) error {
	return nil
}
