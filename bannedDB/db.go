package bannedDB

import (
	"fmt"

	badgerHelper "github.com/Civil/tg-simple-regex-antispam/helper/badger"

	"github.com/dgraph-io/badger/v4"
	"go.uber.org/zap"
)

type BannedDB struct {
	logger   *zap.Logger
	stateDir string
	db       *badger.DB
}

func New(logger *zap.Logger, config map[string]any) (BanDB, error) {
	stateDirI, ok := config["state_dir"]
	if !ok {
		return nil, fmt.Errorf("banDB requires `state_dir` configuration parameter")
	}
	stateDir, ok := stateDirI.(string)
	if !ok {
		return nil, fmt.Errorf("state_dir is not a string")
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

func (r *BannedDB) LoadState() error {
	return nil
}

func (r *BannedDB) SaveState() error {
	return r.db.Sync()
}

func (r *BannedDB) Close() error {
	return r.db.Close()
}
