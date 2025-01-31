package messages

import (
	"sync"

	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/dbs/generic"
	"github.com/Civil/tg-simple-regex-antispam/dbs/interfaces"
)

type MessageDB struct {
	name string
	sync.Mutex
	interfaces.SharedDB
}

var instance = map[string]*MessageDB{}
var lock = &sync.Mutex{}

func Get(logger *zap.Logger, name string) (*MessageDB, error) {
	lock.Lock()
	defer lock.Unlock()
	if db, ok := instance[name]; ok {
		return db, nil
	}

	logger = logger.With(zap.String("db_name", "messages"))
	db, err := generic.New(logger, name, nil)
	if err != nil {
		return nil, err
	}
	msgDB := &MessageDB{
		name:     name,
		SharedDB: db,
	}
	instance[name] = msgDB

	return msgDB, nil
}
