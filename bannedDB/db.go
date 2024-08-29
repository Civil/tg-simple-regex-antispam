package bannedDB

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"

	badgerHelper "github.com/Civil/tg-simple-regex-antispam/helper/badger"
	"github.com/Civil/tg-simple-regex-antispam/helper/stateful"
	"github.com/Civil/tg-simple-regex-antispam/tg/helpers"

	"github.com/dgraph-io/badger/v4"
	"go.uber.org/zap"
)

type BannedDB struct {
	logger   *zap.Logger
	stateDir string
	db       *badger.DB
}

var ErrRequiresStateDir = errors.New(
	"banDB requires `state_dir` configuration parameter",
)

var ErrStateDirNotString = errors.New(
	"state_dir is not a string",
)

func New(logger *zap.Logger, config map[string]any) (BanDB, error) {
	stateDirI, ok := config["state_dir"]
	if !ok {
		return nil, ErrRequiresStateDir
	}
	stateDir, ok := stateDirI.(string)
	if !ok {
		return nil, ErrStateDirNotString
	}

	badgerDB, err := badger.Open(badger.DefaultOptions(stateDir))
	if err != nil {
		return nil, err
	}

	db := &BannedDB{
		logger:   logger.With(zap.String("banDB", "bannedDB")),
		stateDir: stateDir,
		db:       badgerDB,
	}
	return db, nil
}

func (r *BannedDB) BanUser(userID int64) error {
	return r.db.Update(
		func(txn *badger.Txn) error {
			return txn.Set(badgerHelper.UserIDToKey(userID), []byte("1"))
		})
}

func (r *BannedDB) UnbanUser(userID int64) error {
	return r.db.Update(
		func(txn *badger.Txn) error {
			return txn.Delete(badgerHelper.UserIDToKey(userID))
		})
}

func (r *BannedDB) IsBanned(userID int64) bool {
	var exists bool
	err := r.db.View(
		func(tx *badger.Txn) error {
			key := badgerHelper.UserIDToKey(userID)
			if val, err := tx.Get(key); err != nil {
				return err
			} else if val != nil {
				exists = true
			}
			return nil
		})
	if err != nil {
		return false
	}
	return exists
}

func (r *BannedDB) ListUserIDs() ([]int64, error) {
	var userIDs []int64
	err := r.db.View(
		func(tx *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := tx.NewIterator(opts)
			defer it.Close()

			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				key := item.Key()
				userID, err := badgerHelper.KeyToUserID(key)
				if err != nil {
					return err
				}
				userIDs = append(userIDs, userID)
			}
			return nil

		})
	return userIDs, err
}

func (r *BannedDB) LoadState() error {
	return nil
}

func (r *BannedDB) SaveState() error {
	return r.db.Sync()
}

func (r *BannedDB) Close() error {
	return r.db.Close()
}

func (r *BannedDB) TGAdminPrefix() string {
	return "bandb"
}

func (r *BannedDB) HandleTGCommands(logger *zap.Logger, bot *telego.Bot, message *telego.Message,
	tokens []string) error {
	logger.Debug("ban db command", zap.Strings("tokens", tokens))
	switch strings.ToLower(tokens[0]) {
	case "list":
		list, err := r.ListUserIDs()
		if err != nil {
			logger.Error("failed to list banned users", zap.Error(err))
			return err
		}
		buf := bytes.NewBuffer([]byte{})
		buf.WriteString("Banned users:\n")
		for _, userID := range list {
			buf.WriteString(fmt.Sprintf("%v\n", userID))
		}
		err = helpers.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, buf.String())
		if err != nil {
			logger.Error("failed to send message", zap.Error(err))
		}
	case "unban":
		if len(tokens) < 2 {
			logger.Warn("invalid command", zap.Strings("tokens", tokens))
			return stateful.ErrNotSupported
		}
		userID := tokens[1]
		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			logger.Warn("invalid user id", zap.Strings("tokens", tokens), zap.Error(err))
			return stateful.ErrUserIDInvalid
		}
		err = r.UnbanUser(userIDInt)
		if err != nil {
			logger.Error("failed to unban user", zap.String("userID", userID), zap.Error(err))
			_ = helpers.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
				fmt.Sprintf("cannot unban user: %s",
					err.Error()))
			return err
		}
		_ = helpers.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
			fmt.Sprintf("user %v unbanned", userIDInt))
	case "bannodel":
		if len(tokens) < 2 {
			logger.Warn("invalid command", zap.Strings("tokens", tokens))
			return stateful.ErrInvalidCommand
		}
		userID := tokens[1]
		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			logger.Warn("invalid user id", zap.Strings("tokens", tokens), zap.Error(err))
			return err
		}
		err = r.BanUser(userIDInt)
		if err != nil {
			logger.Error("failed to ban user", zap.String("userID", userID), zap.Error(err))
			_ = helpers.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
				fmt.Sprintf("cannot ban user: %s",
					err.Error()))
			return err
		}
		err = helpers.BanUser(bot, message.Chat.ChatID(), userIDInt, false)
		if err != nil {
			logger.Error("failed to ban user in telegram", zap.Error(err))
			return err
		}
		_ = helpers.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
			fmt.Sprintf("user %v banned", userIDInt))
	case "ban":
		if len(tokens) < 2 {
			logger.Warn("invalid command", zap.Strings("tokens", tokens))
			return stateful.ErrInvalidCommand
		}
		userID := tokens[1]
		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			logger.Warn("invalid user id", zap.Strings("tokens", tokens), zap.Error(err))
			return err
		}
		err = r.BanUser(userIDInt)
		if err != nil {
			logger.Error("failed to add user to bandb", zap.String("userID", userID), zap.Error(err))
			_ = helpers.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
				fmt.Sprintf("cannot ban user: %s",
					err.Error()))
			return err
		}
		err = helpers.BanUser(bot, message.Chat.ChatID(), userIDInt, true)
		if err != nil {
			logger.Error("failed to ban user in telegram", zap.Error(err))
			return err
		}
		_ = helpers.SendMessage(bot, message.Chat.ChatID(), &message.MessageID,
			fmt.Sprintf("user %v banned", userIDInt))
	case "help":
		buf := bytes.NewBuffer([]byte{})
		buf.WriteString("Available commands:\n")
		buf.WriteString(" `list` - list all banned users (IDs only)")
		buf.WriteString(" `ban` - ban user by ID and delete all messages")
		buf.WriteString(" `banNoDel` - ban user by ID but keep all messages")
		buf.WriteString(" `unban` - unban user by ID")
		buf.WriteString(" `help` - this help")

		err := helpers.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, buf.String())
		if err != nil {
			logger.Error("failed to send message", zap.Error(err))
		}
		return err
	default:
		logger.Warn("unsupported command", zap.Strings("tokens", tokens))
		return stateful.ErrNotSupported
	}
	return nil
}
