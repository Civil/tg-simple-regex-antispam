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
	chainName      string
	n              int
	logger         *zap.Logger
	filteringRules []interfaces.FilteringRule

	bannedUsers bannedDB.BanDB

	stateDir string

	actions []actions.Action

	db *badger.DB

	isFinal                bool
	warnAboutAlreadyBanned bool

	tg.TGHaveAdminCommands
}

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

	warnAboutAlreadyBanned, err := config2.GetOptionBoolWithDefault(config, "warnAboutAlreadyBanned", false)
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
			zap.String("filter_type", "checkNevents"),
		),
		chainName:              chainName,
		stateDir:               stateDir,
		bannedUsers:            banDB,
		filteringRules:         filteringRules,
		actions:                actions,
		db:                     badgerDB,
		isFinal:                isFinal,
		n:                      n,
		warnAboutAlreadyBanned: warnAboutAlreadyBanned,
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

func Help() string {
	return "checkNevents requires `stateFile` parameter"
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
	r.logger.Debug("scoring message", zap.Any("message", msg))
	userID := msg.From.ID
	logger := r.logger.With(zap.Int64("userID", userID))
	maxScore := &scoringResult.ScoringResult{}
	if r.bannedUsers.IsBanned(userID) && r.warnAboutAlreadyBanned {
		logger.Warn("user is banned, but somehow sends messages, deleting them")
		maxScore.Score = 100
		maxScore.Reason = "user was already banned"
		err := r.applyActions(logger, maxScore, msg.Chat.ChatID(), msg, []int64{int64(msg.MessageID)}, userID)
		if err != nil {
			logger.Error("failed to apply actions", zap.Error(err))
		}

		return maxScore
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
		logger.Debug("user is not a spammer, already verified")
		return maxScore
	}

	actualState.MessageIds[int64(msg.MessageID)] = true
	actualState.LastUpdate = timestamppb.Now()

	// Checking for the filters to match the message
	for _, filter := range r.filteringRules {
		score := filter.Score(bot, msg)
		if score.Score > maxScore.Score {
			maxScore = score
			if filter.IsFinal() {
				break
			}
		}
	}
	if maxScore.Score == 100 {
		// We don't care about State of a spammer, but we need to track if they are banned (at least for some time)
		logger.Debug("user is a spammer, banning them",
			zap.String("username", msg.From.Username),
		)
		err = r.bannedUsers.BanUser(userID)
		if err != nil {
			logger.Error("failed to ban user", zap.Error(err))
			return maxScore
		}

		messageIds := make([]int64, 0, len(actualState.MessageIds))
		for id := range actualState.MessageIds {
			messageIds = append(messageIds, id)
		}

		err = r.applyActions(logger, maxScore, msg.Chat.ChatID(), msg, messageIds, userID)
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

func (r *Filter) applyActions(logger *zap.Logger, score *scoringResult.ScoringResult, ChatID telego.ChatID, msg *telego.Message, messageIds []int64, userID int64) error {
	for _, action := range r.actions {
		var err error
		if action.PerMessage() {
			err = action.ApplyToMessage(r, score, msg, nil)
		} else {
			err = action.Apply(r, score, ChatID, messageIds, userID, nil)
		}

		if err != nil {
			logger.Error("failed to apply action", zap.Any("action", action), zap.Error(err))
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

func (r *Filter) Close() error {
	return r.db.Close()
}

func (r *Filter) SaveState() error {
	return nil
}

func (r *Filter) LoadState() error {
	return nil
}

func (r *Filter) UnbanUser(userID int64) error {
	newState := &checkNeventsState.State{
		Verified:   true,
		MessageIds: make(map[int64]bool),
		LastUpdate: timestamppb.Now(),
	}
	return r.setState(userID, newState)
}

func (r *Filter) TGAdminPrefix() string {
	return r.chainName
}
