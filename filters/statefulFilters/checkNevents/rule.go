package checkNevents

import (
	"errors"

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
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
	"github.com/Civil/tg-simple-regex-antispam/helper/tg"
)

type Filter struct {
	chainName      string
	n              int
	logger         *zap.Logger
	filteringRules []interfaces.FilteringRule

	bannedUsers bannedDB.BanDB

	stateDir string

	actions []actions.Action

	db *badger.DB

	isFinal bool

	tg.TGHaveAdminCommands
}

var (
	ErrStateDirEmpty = errors.New("state_dir cannot be empty")
	ErrNIsZero       = errors.New("n cannot be equal to 0")
)

func New(logger *zap.Logger, chainName string, banDB bannedDB.BanDB, _ *telego.Bot, config map[string]any,
	filteringRules []interfaces.FilteringRule, actions []actions.Action,
) (interfaces.StatefulFilter, error) {
	stateDir, err := config2.GetOptionString(config, "state_dir")
	if err != nil {
		return nil, err
	}
	if stateDir == "" {
		return nil, ErrStateDirEmpty
	}

	n, err := config2.GetOptionInt(config, "n")
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, ErrNIsZero
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
			zap.String("filter_type", "checkNevents"),
		),
		chainName:      chainName,
		stateDir:       stateDir,
		bannedUsers:    banDB,
		filteringRules: filteringRules,
		actions:        actions,
		db:             badgerDB,
		isFinal:        isFinal,
		n:              n,
		TGHaveAdminCommands: tg.TGHaveAdminCommands{
			Handlers: make(map[string]tg.AdminCMDHandlerFunc),
		},
	}

	for _, filter := range f.filteringRules {
		prefix := filter.TGAdminPrefix()
		if prefix != "" {
			f.TGHaveAdminCommands.Handlers[prefix] = filter.HandleTGCommands
		}
	}

	f.TGHaveAdminCommands.Handlers[f.bannedUsers.TGAdminPrefix()] = f.bannedUsers.HandleTGCommands

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
	r.logger.Debug("scoring message", zap.Any("message", msg))
	userID := msg.From.ID
	logger := r.logger.With(zap.Int64("userID", userID))
	if r.bannedUsers.IsBanned(userID) {
		logger.Warn("user is banned, but somehow sends messages, deleting them")
		err := r.applyActions(logger, msg.Chat.ChatID(), []int64{int64(msg.MessageID)}, userID)
		if err != nil {
			logger.Error("failed to apply actions", zap.Error(err))
		}
		return 100
	}

	actualState, err := r.getState(userID)
	if err != nil {
		logger.Error("failed to get state", zap.Error(err))
		actualState = &checkNeventsState.State{
			Verified:   false,
			MessageIds: make(map[int64]bool),
			LastUpdate: timestamppb.Now(),
		}
	} else if actualState == nil || (!actualState.Verified && len(actualState.MessageIds) == 0) {
		logger.Debug("state is empty, creating an empty one",
			zap.Error(err),
		)
		actualState = &checkNeventsState.State{
			Verified:   false,
			MessageIds: make(map[int64]bool),
			LastUpdate: timestamppb.Now(),
		}
	}

	// We already verified that user
	if actualState.Verified {
		logger.Debug("user is not a spammer, verified")
		return 0
	}

	actualState.MessageIds[int64(msg.MessageID)] = true
	actualState.LastUpdate = timestamppb.Now()

	maxScore := 0
	// Checking for the filters to match the message
	for _, filter := range r.filteringRules {
		score := filter.Score(msg)
		if score > maxScore {
			maxScore = score
			if filter.IsFinal() {
				break
			}
		}
	}
	if maxScore == 100 {
		// We don't care about State of a spammer, but we need to track if they are banned (at least for some time)
		logger.Debug("user is a spammer, banning them")
		err = r.bannedUsers.BanUser(userID)
		if err != nil {
			logger.Error("failed to ban user", zap.Error(err))
			return maxScore
		}

		messageIds := make([]int64, 0, len(actualState.MessageIds))
		for id := range actualState.MessageIds {
			messageIds = append(messageIds, id)
		}

		err = r.applyActions(logger, msg.Chat.ChatID(), messageIds, userID)
		if err != nil {
			logger.Error("failed to apply actions", zap.Error(err))
		}

		err = r.RemoveState(userID)
		if err != nil {
			logger.Error("failed to remove state", zap.Error(err))
			return maxScore
		}
		return maxScore
	}
	logger.Debug("message verified, updating state")
	if len(actualState.MessageIds) >= r.n {
		logger.Debug("reached threshold, marking user as verified", zap.Int("n", r.n))
		actualState.Verified = true
		actualState.MessageIds = nil
	}
	err = r.setState(userID, actualState)
	if err != nil {
		logger.Error("failed to set new state",
			zap.Any("new_state", actualState),
			zap.Error(err),
		)
	}
	return maxScore
}

func (r *Filter) applyActions(logger *zap.Logger, ChatID telego.ChatID, messageIds []int64, userID int64) error {
	for _, action := range r.actions {
		err := action.Apply(r, ChatID, messageIds, userID)
		if err != nil {
			logger.Error("failed to apply action", zap.Any("action", action), zap.Error(err))
			return err
		}
	}
	return nil
}

func (r *Filter) IsStateful() bool {
	return true
}

func (r *Filter) GetFilterName() string {
	return r.chainName
}

func (r *Filter) GetName() string {
	return "checkNEvents"
}

func (r *Filter) IsFinal() bool {
	return r.isFinal
}

func Help() string {
	return "checkNevents requires `stateFile` parameter"
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
