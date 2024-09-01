package logs

import (
	"fmt"
	"regexp"

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

	tgApiTokenRE *regexp.Regexp
}

// Warningf is required for badgerdb, and it is required to have interface args...
func (l *StdLogger) Warningf(msg string, args ...interface{}) {
	l.Warn(zap.String("msg", fmt.Sprintf(msg, args...)))
}

func (l *StdLogger) Debugf(msg string, args ...interface{}) {
	for i := range args {
		if s, ok := args[i].(string); ok {
			args[i] = l.tgApiTokenRE.ReplaceAllString(s, "api.telegram.org/bot**TOKEN**/")
		}
	}
	l.SugaredLogger.Debugf(l.tgApiTokenRE.ReplaceAllString(msg, "api.telegram.org/bot**TOKEN**/"), args...)
}

func New(logger *zap.Logger) *StdLogger {
	return &StdLogger{
		SugaredLogger: *logger.Sugar(),
		tgApiTokenRE:  regexp.MustCompile(`api\.telegram\.org/bot[^/]+/`),
	}
}
