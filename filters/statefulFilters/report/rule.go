package report

import (
	"fmt"
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
	n      int
	logger *zap.Logger

	stateDir string

	actions []actions.Action

	db *badger.DB

	isFinal bool
}

func New(logger *zap.Logger, _ bannedDB.BanDB, config map[string]interface{},
	_ []interfaces.FilteringRule, actions []actions.Action) (interfaces.StatefulFilter, error) {
	var stateDir string
	var err error
	stateDir, err = config2.GetOptionString(config, "state_dir")
	if err != nil {
		return nil, err
	}
	if stateDir == "" {
		return nil, fmt.Errorf("state_dir cannot be empty")
	}

	n, err := config2.GetOptionInt(config, "n")
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, fmt.Errorf("n cannot be equal to 0")
	}

	isFinal, err := config2.GetOptionBool(config, "isFinal")
	if err != nil {
		return nil, err
	}

	badgerDB, err := badger.Open(badger.DefaultOptions(stateDir))
	if err != nil {
		return nil, err
	}

	f := &Filter{
		logger:   logger.With(zap.String("filter", "checkNevents")),
		stateDir: stateDir,
		db:       badgerDB,
		isFinal:  isFinal,
		actions:  actions,
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

func (r *Filter) removeState(userID int64) error {
	return r.db.Update(
		func(txn *badger.Txn) error {
			return txn.Delete(badgerHelper.UserIDToKey(userID))
		})
}

func (r *Filter) Score(msg telego.Message) int {
	if strings.HasPrefix("/report", msg.Text) || strings.HasPrefix("/spam", msg.Text) {
		return 100
	}
	return -1
	userID := msg.From.ID
	actualState, err := r.getState(userID)
	if err != nil || actualState == nil || len(actualState.MessageIds) == 0 {
		r.logger.Debug("failed to get state",
			zap.Int64("userID", userID),
			zap.Error(err),
		)
		actualState = &state.State{
			Verified:   false,
			MessageIds: []int64{int64(msg.MessageID)},
			LastUpdate: timestamppb.Now(),
		}
	}

	// We already verified that user
	if actualState.Verified {
		return -1
	}

	actualState.MessageIds = append(actualState.MessageIds, int64(msg.MessageID))
	actualState.LastUpdate = timestamppb.Now()

	maxScore := 0
	for _, filter := range r.filteringRules {
		score := filter.Score(msg)
		if score > maxScore {
			maxScore = score
			if maxScore == 100 {
				// We don't care about State of a spammer, but we need to track if they are banned (for some time)
				err = r.bannedUsers.BanUser(userID)
				if err != nil {
					r.logger.Error("failed to ban user", zap.Int64("userID", userID), zap.Error(err))
					return maxScore
				}

				for _, action := range r.actions {
					err = action.Apply(msg.Chat.ChatID(), actualState.MessageIds, userID)
					if err != nil {
						r.logger.Error("failed to apply action", zap.Any("action", action), zap.Error(err))
						return maxScore
					}
				}

				err = r.removeState(userID)
				if err != nil {
					r.logger.Error("failed to remove state", zap.Int64("userID", userID), zap.Error(err))
					return maxScore
				}
				return maxScore
			}
		}
	}
	if maxScore > 0 {
		return maxScore
	}
	if len(actualState.MessageIds) >= r.n {
		actualState.Verified = true
		actualState.MessageIds = nil
	}
	err = r.setState(userID, actualState)
	if err != nil {
		r.logger.Error("failed to set new state",
			zap.Int64("userID", userID),
			zap.Any("new_state", actualState),
			zap.Error(err),
		)
	}
	return 0
}

func (r *Filter) IsStateful() bool {
	return true
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
