package badgerOpts

import (
	"github.com/dgraph-io/badger/v4"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/helper/logs"
)

func GetBadgerOptions(logger *zap.Logger, badgerDBname string, dir string) badger.Options {
	opts := badger.DefaultOptions(dir)
	stdLog := logs.New(logger.With(
		zap.String("type", "badgerDB"),
		zap.String("component", badgerDBname),
	))
	opts.Logger = stdLog
	return opts
}
