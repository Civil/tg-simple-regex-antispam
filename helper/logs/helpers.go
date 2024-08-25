package logs

import (
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
