package report

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	actions "github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/bannedDB"
	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/types/checkNeventsState"
	"github.com/Civil/tg-simple-regex-antispam/filters/types/scoringResult"
	badgerHelper "github.com/Civil/tg-simple-regex-antispam/helper/badger"
	"github.com/Civil/tg-simple-regex-antispam/helper/badger/badgerOpts"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
	"github.com/Civil/tg-simple-regex-antispam/helper/tg"
)

var (
	ErrStateDirEmpty = errors.New("state_dir cannot be empty")
	ErrNIsZero       = errors.New("n cannot be equal to 0")
)

type Filter struct {
	sync.RWMutex
	chainName string
	logger    *zap.Logger

	stateDir string

	filteringRules []interfaces.FilteringRule
	actions        []actions.Action

	db  *badger.DB
	bot *telego.Bot

	isFinal         bool
	removeReportMsg bool

	vacations map[string]time.Time

	tg.TGHaveAdminCommands
}

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
		vacations:       make(map[string]time.Time),
		TGHaveAdminCommands: tg.TGHaveAdminCommands{
			Handlers: make(map[string]tg.AdminCMDHandlerFunc),
		},
	}

	f.TGHaveAdminCommands.Handlers["vacation"] = f.vacationCmd
	go f.cleanVacationState()
	return f, nil
}

func Help() string {
	return "report requires `stateFile` parameter"
}

func (r *Filter) cleanVacationState() {
	for {
		for admin, vacationEndDate := range r.vacations {
			if time.Now().After(vacationEndDate) {
				_ = r.removeVacation(admin)
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func (r *Filter) vacationCmd(logger *zap.Logger, bot *telego.Bot, message *telego.Message, tokens []string) error {
	r.logger.Debug("vacation command", zap.Strings("tokens", tokens))
	if len(tokens) == 0 || tokens[0] == "list" {
		return r.listVacations(logger, bot, message, tokens)
	}
	switch tokens[0] {
	case "add":
		return r.addVacationCmd(logger, bot, message, tokens[1:])
	case "remove":
		return r.removeVacationCmd(logger, bot, message, tokens[1:])
	default:
		return fmt.Errorf("unknown subcommand: %v", tokens[0])
	}
}

func (r *Filter) listVacations(logger *zap.Logger, bot *telego.Bot, message *telego.Message, tokens []string) error {
	r.logger.Debug("list vacations command", zap.Strings("tokens", tokens))
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString("Admins on vacation:\n")
	r.logger.Debug("taking lock")
	r.RLock()
	r.logger.Debug("lock taken")
	if len(r.vacations) == 0 {
		buf.WriteString("No admins on vacation\n")
	} else {
		for admin, vacationEndDate := range r.vacations {
			buf.WriteString(fmt.Sprintf("%v is on vacation until %v\n", admin, vacationEndDate))
		}
	}
	r.RUnlock()
	r.logger.Debug("lock released")

	r.logger.Debug("sending message", zap.String("message", buf.String()))
	err := tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, buf.String())
	if err != nil {
		r.logger.Error("failed to send message", zap.Error(err))
	}
	return err
}

func (r *Filter) addVacationCmd(logger *zap.Logger, bot *telego.Bot, message *telego.Message, tokens []string) error {
	if len(tokens) == 0 {
		return fmt.Errorf("no admin username specified")
	}
	if len(tokens) == 1 {
		return fmt.Errorf("no vacation duration provided")
	}
	admin := tokens[0]
	vacationDuration, err := time.ParseDuration(tokens[1])
	if err != nil {
		return fmt.Errorf("failed to parse duration for admin %v: %v", admin, err)
	}
	r.Lock()
	r.vacations[admin] = time.Now().Add(vacationDuration)
	r.Unlock()
	err = tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, fmt.Sprintf("Admin %v is on vacation until %v", admin, r.vacations[admin]))
	if err != nil {
		r.logger.Error("failed to send message", zap.Error(err))
	}
	return err
}

func (r *Filter) removeVacationCmd(logger *zap.Logger, bot *telego.Bot, message *telego.Message,
	tokens []string) error {
	if len(tokens) == 0 {
		return fmt.Errorf("no admin username specified")
	}
	admin := tokens[0]
	err := r.removeVacation(admin)
	if err != nil {
		return err
	}
	err = tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
		fmt.Sprintf("Admin %v is no longer on vacation", admin))
	if err != nil {
		r.logger.Error("failed to send message", zap.Error(err))
	}
	return err
}

func (r *Filter) removeVacation(admin string) error {
	r.Lock()
	defer r.Unlock()
	_, ok := r.vacations[admin]
	if !ok {
		return fmt.Errorf("admin %v is not on vacation", admin)
	}
	delete(r.vacations, admin)
	return nil
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

func (r *Filter) Score(bot *telego.Bot, msg *telego.Message) *scoringResult.ScoringResult {
	score := &scoringResult.ScoringResult{}
	if !strings.HasPrefix(msg.Text, "/report") {
		r.logger.Debug("message does not start with /report")
		return score
	}
	if msg.ReplyToMessage == nil {
		r.logger.Debug("message does not have a reply")
		err := tg.SendMessage(r.bot, msg.Chat.ChatID(), &msg.MessageID, "Report must be a reply to a message")
		if err != nil {
			r.logger.Error("failed to send message", zap.Error(err))
		}
		return score
	}

	r.logger.Debug("got a message that start with /report",
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
		err = tg.SendMessage(r.bot, msg.Chat.ChatID(), &msg.MessageID, "Message/user was already reported")
		if err != nil {
			r.logger.Error("failed to send message", zap.Error(err))
		}

		if r.removeReportMsg {
			err = tg.DeleteMessage(r.bot, msg)
			if err != nil {
				r.logger.Error("failed to delete message", zap.Error(err))
			}
		}
		score.Score = -1
		score.Reason = "already reported"
		return score
	}
	if r.removeReportMsg {
		err = tg.DeleteMessage(r.bot, msg)
		if err != nil {
			r.logger.Error("failed to delete message", zap.Error(err))
		}
	}

	r.logger.Debug("applying actions...")
	score.Score = 100
	score.Reason = "reported command"
	for _, action := range r.actions {
		r.logger.Debug("trying to apply action",
			zap.Any("message_ids", actualState.MessageIds),
			zap.Any("action", action),
		)
		err = action.ApplyToMessage(r, score, reportedMsg, r.vacations)
		if err != nil {
			r.logger.Error("failed to apply action", zap.Any("action", action), zap.Error(err))
			return score
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

	return score
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
	return r.chainName
}

func (r *Filter) UnbanUser(_ int64) error {
	return nil
}
