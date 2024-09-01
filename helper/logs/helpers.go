package logs

import (
	"fmt"

	"go.uber.org/zap"
)

/*ErrNST logs an error with a message and an error, increasing the stacktrace level to avoid having a stacktrace for errors.
 *
 * @param l *zap.Logger The zap logger instance.
 * @param msg string The error message.
 * @param e error The error to log.
 */
func ErrNST(l *zap.Logger, msg string, e error) {
	// increase stacktrace level to zap.DPanicLevel to avoid having stacktrace for errors
	l.WithOptions(zap.AddStacktrace(zap.DPanicLevel)).Error(msg, zap.Error(e))
}

type StdLogger struct {
	zap.SugaredLogger
}

// Warningf is required for badgerdb, and it is required to have interface args...
func (l *StdLogger) Warningf(msg string, args ...interface{}) {
	l.Warn(zap.String("msg", fmt.Sprintf(msg, args...)))
}

func New(logger *zap.Logger) *StdLogger {
	return &StdLogger{SugaredLogger: *logger.Sugar()}
}
