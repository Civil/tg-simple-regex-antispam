package generic

import (
	"github.com/dgraph-io/badger/v4"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/constants"
	"github.com/Civil/tg-simple-regex-antispam/dbs/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/helper/badger/badgerOpts"
	"github.com/Civil/tg-simple-regex-antispam/helper/stateful"
)

type GenericDB struct {
	stateful.Stateful
	logger   *zap.Logger
	stateDir string
	db       *badger.DB
}

func New(logger *zap.Logger, name string, config map[string]any) (interfaces.SharedDB, error) {
	stateDirI, ok := config["state_dir"]
	if !ok {
		return nil, constants.ErrRequiresStateDir
	}
	stateDir, ok := stateDirI.(string)
	if !ok {
		return nil, constants.ErrStateDirNotString
	}

	badgerDB, err := badger.Open(badgerOpts.GetBadgerOptions(logger, "bannedDB", stateDir))
	if err != nil {
		return nil, err
	}

	db := &GenericDB{
		logger:   logger.With(zap.String("db_name", name)),
		stateDir: stateDir,
		db:       badgerDB,
	}
	return db, nil
}

// StoreValue stores a value in the database
func (db *GenericDB) StoreValue(key []byte, value []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (db *GenericDB) LoadValue(key []byte) ([]byte, error) {
	var value []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	return value, err
}

func (db *GenericDB) LoadState() error {
	return nil
}

func (db *GenericDB) SaveState() error {
	return db.db.Sync()
}

func (db *GenericDB) Close() error {
	return db.db.Close()
}
